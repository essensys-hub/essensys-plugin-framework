package plugin

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Registry est le registre compilé des plugins. Aucun chargement dynamique:
// les adaptateurs sont enregistrés au build via Register, puis activés selon
// leur manifest.
type Registry struct {
	mu        sync.RWMutex
	adapters  map[string]PluginAdapter
	manifests map[string]Manifest
	enabled   map[string]bool
	store     Store
	sink      MetricSink
}

// New crée un registre. store et sink sont fournis par l'application hôte
// (Redis + Prometheus en prod, MemStore + MemSink en test).
func New(store Store, sink MetricSink) *Registry {
	return &Registry{
		adapters:  map[string]PluginAdapter{},
		manifests: map[string]Manifest{},
		enabled:   map[string]bool{},
		store:     store,
		sink:      sink,
	}
}

// Register enregistre un adaptateur au build. Idempotent par ID.
func (r *Registry) Register(a PluginAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[a.ID()] = a
}

// StateStore est un port optionnel du Store : persistance de l'état
// activé/désactivé des plugins entre redémarrages.
type StateStore interface {
	SetEnabled(pluginID string, enabled bool)
	// PersistedEnabled renvoie (état, true) si un état a été persisté.
	PersistedEnabled(pluginID string) (bool, bool)
}

// PurgeStore est un port optionnel du Store : effacement des données d'un
// plugin (snapshot + historique) lors d'une désinstallation.
type PurgeStore interface {
	Purge(pluginID string)
}

// Configure associe un manifest à un adaptateur enregistré et l'active
// (sauf état désactivé persisté). Refuse une version de framework
// incompatible ou un adaptateur absent.
func (r *Registry) Configure(m Manifest) error {
	if err := m.Validate(); err != nil {
		return err
	}
	if !m.CompatibleFramework() {
		return fmt.Errorf("plugin %q exige framework %s, incompatible avec %s", m.ID, m.FrameworkVersion, FrameworkVersion)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.adapters[m.ID]; !ok {
		return fmt.Errorf("aucun adaptateur compilé pour le plugin %q", m.ID)
	}
	r.manifests[m.ID] = m
	enabled := true
	if ss, ok := r.store.(StateStore); ok {
		if persisted, has := ss.PersistedEnabled(m.ID); has {
			enabled = persisted
		}
	}
	r.enabled[m.ID] = enabled
	return nil
}

// Enable active un plugin configuré et persiste l'état.
func (r *Registry) Enable(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.manifests[id]; !ok {
		return fmt.Errorf("plugin %q non configuré", id)
	}
	r.enabled[id] = true
	if ss, ok := r.store.(StateStore); ok {
		ss.SetEnabled(id, true)
	}
	return nil
}

// Disable désactive un plugin sans effacer ses séries (l'historique est
// conservé) et persiste l'état.
func (r *Registry) Disable(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled[id] = false
	if ss, ok := r.store.(StateStore); ok {
		ss.SetEnabled(id, false)
	}
}

// PluginInfo est l'entrée du catalogue exposé à l'écran Paramètres.
type PluginInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Enabled     bool     `json:"enabled"`
	WriteScope  string   `json:"write_scope,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// List renvoie le catalogue des plugins compilés et configurés.
func (r *Registry) List() []PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]PluginInfo, 0, len(r.manifests))
	for id, m := range r.manifests {
		out = append(out, PluginInfo{
			ID: id, Name: m.Name, Version: m.Version, Description: m.Description,
			Enabled: r.enabled[id], WriteScope: m.WriteScope, Capabilities: m.Capabilities,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Enabled indique si un plugin est actif.
func (r *Registry) Enabled(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.enabled[id]
}

// Subscribe abonne chaque plugin configuré au bus ; l'état activé/désactivé
// est vérifié à l'ingestion, pour qu'un Enable/Disable à chaud soit effectif
// sans redémarrage.
func (r *Registry) Subscribe(bus Bus) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, a := range r.adapters {
		if _, configured := r.manifests[id]; !configured {
			continue
		}
		adapter := a
		filter := TopicFilter(id)
		err := bus.Subscribe(filter, func(msg BusMessage) {
			if !r.Enabled(adapter.ID()) {
				return
			}
			r.ingest(adapter, msg)
		})
		if err != nil {
			return fmt.Errorf("abonnement %q: %w", id, err)
		}
	}
	return nil
}

// Ingest est exposé pour les tests et l'ingestion directe (sans bus réel).
func (r *Registry) Ingest(a PluginAdapter, msg BusMessage) { r.ingest(a, msg) }

func (r *Registry) ingest(a PluginAdapter, msg BusMessage) {
	samples, err := a.OnMessage(msg)
	if err != nil || len(samples) == 0 {
		return
	}
	now := time.Now()
	for i := range samples {
		if samples[i].TS.IsZero() {
			samples[i].TS = now
		}
		if samples[i].MachineID == "" {
			samples[i].MachineID = msg.MachineID
		}
		r.sink.Observe(a.ID(), samples[i])
	}
	r.store.Put(a.ID(), msg.MachineID, samples, now)
}

// manifest renvoie le manifest actif d'un plugin.
func (r *Registry) manifest(id string) (Manifest, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.manifests[id]
	return m, ok && r.enabled[id]
}
