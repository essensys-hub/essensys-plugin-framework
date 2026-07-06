package plugin

import (
	"encoding/json"
	"net/http"
	"strings"
)

// AuthFunc extrait l'identité d'une requête HTTP. Fournie par l'application
// hôte, elle réutilise le middleware d'auth existant (JWT/session cloud,
// trusted-devices LAN). Le framework n'invente aucun schéma d'autorisation.
type AuthFunc func(*http.Request) (Identity, bool)

// Mount enregistre les routes /api/plugins/<id>/{descriptor,current} sur mux.
// Chaque route applique le RBAC déclaratif du manifest via auth.
func (r *Registry) Mount(mux *http.ServeMux, auth AuthFunc) {
	mux.HandleFunc("/api/plugins/", func(w http.ResponseWriter, req *http.Request) {
		id, action, ok := parsePath(req.URL.Path)
		if !ok {
			http.NotFound(w, req)
			return
		}
		m, active := r.manifest(id)
		if !active {
			http.NotFound(w, req)
			return
		}
		ident, authed := auth(req)
		if !authed || !m.Allows(ident.Roles) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		switch action {
		case "descriptor":
			writeJSON(w, r.descriptor(id))
		case "current":
			writeJSON(w, r.store.Current(id))
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
