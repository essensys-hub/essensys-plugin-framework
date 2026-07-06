package plugin

import (
	"fmt"
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

// Configure associe un manifest à un adaptateur enregistré et l'active.
// Refuse une version de framework incompatible ou un adaptateur absent.
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
	r.enabled[m.ID] = true
	return nil
}

// Disable désactive un plugin sans effacer ses séries (l'historique est conservé).
func (r *Registry) Disable(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled[id] = false
}

// Enabled indique si un plugin est actif.
func (r *Registry) Enabled(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.enabled[id]
}

// Subscribe abonne chaque plugin actif au bus. Un message reçu est traduit en
// échantillons (idempotence via clé de série), persistés dans store + sink.
func (r *Registry) Subscribe(bus Bus) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, a := range r.adapters {
		if !r.enabled[id] {
			continue
		}
		adapter := a
		filter := TopicFilter(id)
		err := bus.Subscribe(filter, func(msg BusMessage) {
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
