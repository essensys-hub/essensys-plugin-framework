package plugin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AuthFunc extrait l'identité d'une requête HTTP. Fournie par l'application
// hôte, elle réutilise le middleware d'auth existant (JWT/session cloud,
// trusted-devices LAN). Le framework n'invente aucun schéma d'autorisation.
type AuthFunc func(*http.Request) (Identity, bool)

// IsAdmin indique si l'identité porte un rôle d'administration.
func (i Identity) IsAdmin() bool {
	for _, r := range i.Roles {
		if r == RoleAdminLocal || r == RoleAdminGlobal || r == RoleLANAdmin {
			return true
		}
	}
	return false
}

// Mount enregistre les routes /api/plugins/* sur mux :
//   - GET  /api/plugins/                    catalogue (écran Paramètres)
//   - GET  /api/plugins/<id>/descriptor|current|history
//   - POST /api/plugins/<id>/enable|disable|purge   (admin uniquement)
//
// Chaque route de lecture applique le RBAC déclaratif du manifest via auth.
func (r *Registry) Mount(mux *http.ServeMux, auth AuthFunc) {
	mux.HandleFunc("/api/plugins/", func(w http.ResponseWriter, req *http.Request) {
		ident, authed := auth(req)
		if !authed {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Catalogue : GET /api/plugins/
		if strings.Trim(strings.TrimPrefix(req.URL.Path, "/api/plugins"), "/") == "" {
			if req.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, r.List())
			return
		}

		id, action, ok := parsePath(req.URL.Path)
		if !ok {
			http.NotFound(w, req)
			return
		}

		// Actions d'administration (plugin désactivé inclus).
		switch action {
		case "enable", "disable", "purge":
			if req.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if !ident.IsAdmin() {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			switch action {
			case "enable":
				if err := r.Enable(id); err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
			case "disable":
				r.Disable(id)
			case "purge":
				r.Disable(id)
				if ps, isPurge := r.store.(PurgeStore); isPurge {
					ps.Purge(id)
				}
			}
			writeJSON(w, r.List())
			return
		}

		m, active := r.manifest(id)
		if !active {
			http.NotFound(w, req)
			return
		}
		if !m.Allows(ident.Roles) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		switch action {
		case "descriptor":
			writeJSON(w, r.descriptor(id))
		case "current":
			writeJSON(w, r.store.Current(id))
		case "history":
			hp, isHist := r.store.(HistoryProvider)
			if !isHist {
				http.NotFound(w, req)
				return
			}
			metric := req.URL.Query().Get("metric")
			if metric == "" {
				http.Error(w, "metric requis", http.StatusBadRequest)
				return
			}
			hours := 24
			if h, err := strconv.Atoi(req.URL.Query().Get("hours")); err == nil && h > 0 && h <= 48 {
				hours = h
			}
			pts := hp.History(id, metric, time.Now().Add(-time.Duration(hours)*time.Hour))
			if pts == nil {
				pts = []Point{}
			}
			writeJSON(w, map[string]any{"metric": metric, "points": pts})
		default:
			http.NotFound(w, req)
		}
	})
}

func (r *Registry) descriptor(id string) Descriptor {
	r.mu.RLock()
	a := r.adapters[id]
	r.mu.RUnlock()
	if a == nil {
		return Descriptor{PluginID: id}
	}
	d := a.Descriptor()
	d.ReadOnly = true // MVP: lecture seule, garanti côté framework.
	return d
}

// parsePath extrait (id, action) de /api/plugins/<id>/<action>.
func parsePath(p string) (id, action string, ok bool) {
	rest := strings.TrimPrefix(p, "/api/plugins/")
	if rest == p {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
