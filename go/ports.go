package plugin

import (
	"sync"
	"time"
)

// ---- Ports (implémentés par l'application hôte) ----

// BusMessage est un message reçu du bus (Mosquitto en prod).
type BusMessage struct {
	Topic     string
	PluginID  string
	MachineID string
	Payload   []byte
}

// Bus est le transport d'ingestion (Mosquitto). Fourni par l'hôte.
type Bus interface {
	// Subscribe abonne un handler à un filtre de topic.
	Subscribe(topicFilter string, h func(BusMessage)) error
}

// MetricSink persiste les séries temporelles (Prometheus en prod).
type MetricSink interface {
	Observe(pluginID string, s Sample)
}

// Store conserve la dernière valeur connue (Redis en prod).
type Store interface {
	Put(pluginID, machineID string, samples []Sample, at time.Time)
	Current(pluginID string) Reading
}

// Reading est l'instantané renvoyé par l'API /current.
type Reading struct {
	PluginID  string    `json:"plugin_id"`
	Samples   []Sample  `json:"samples"`
	UpdatedAt time.Time `json:"updated_at"`
	Stale     bool      `json:"stale"`
}

// Identity est l'identité résolue par le middleware d'auth existant.
type Identity struct {
	Roles []Role
}

// ---- Implémentations en mémoire (tests + périmètre sans Redis) ----

// MemStore est un Store en mémoire, thread-safe, avec calcul de fraîcheur.
type MemStore struct {
	mu       sync.RWMutex
	readings map[string]*Reading
	deadline time.Duration
	now      func() time.Time
}

// NewMemStore crée un store dont les lectures deviennent stale après staleAfter.
func NewMemStore(staleAfter time.Duration) *MemStore {
	return &MemStore{readings: map[string]*Reading{}, deadline: staleAfter, now: time.Now}
}

func (s *MemStore) Put(pluginID, machineID string, samples []Sample, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.readings[pluginID] = &Reading{PluginID: pluginID, Samples: samples, UpdatedAt: at}
}

func (s *MemStore) Current(pluginID string) Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.readings[pluginID]
	if !ok {
		return Reading{PluginID: pluginID, Stale: true}
	}
	out := *r
	out.Stale = s.now().Sub(r.UpdatedAt) > s.deadline
	return out
}

// MemSink capture les échantillons observés (pour tests/idempotence).
type MemSink struct {
	mu     sync.Mutex
	Series map[string][]Sample // clé = pluginID|machineID|metric
}

// NewMemSink crée un sink en mémoire.
func NewMemSink() *MemSink { return &MemSink{Series: map[string][]Sample{}} }

func (m *MemSink) Observe(pluginID string, s Sample) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := pluginID + "|" + s.MachineID + "|" + s.Metric
	m.Series[key] = append(m.Series[key], s)
}
