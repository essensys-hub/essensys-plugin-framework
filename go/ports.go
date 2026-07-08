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
	history  map[string][]Point // clé = pluginID|metric
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
	var base []Sample
	if prev, ok := s.readings[pluginID]; ok {
		base = prev.Samples
	}
	s.readings[pluginID] = &Reading{PluginID: pluginID, Samples: MergeSamples(base, samples), UpdatedAt: at}
	if s.history == nil {
		s.history = map[string][]Point{}
	}
	for _, sm := range samples {
		k := pluginID + "|" + sm.Metric
		ts := sm.TS
		if ts.IsZero() {
			ts = at
		}
		s.history[k] = append(s.history[k], Point{TS: ts, Value: sm.Value})
	}
}

// History renvoie la série d'une métrique depuis since (implémente HistoryProvider).
func (s *MemStore) History(pluginID, metric string, since time.Time) []Point {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Point
	for _, p := range s.history[pluginID+"|"+metric] {
		if !p.TS.Before(since) {
			out = append(out, p)
		}
	}
	return out
}

// MergeSamples fusionne des échantillons entrants dans un snapshot existant,
// avec upsert par clé de série (machine_id|metric) : un message MQTT ne porte
// qu'une série, le snapshot courant doit conserver les autres.
func MergeSamples(existing, incoming []Sample) []Sample {
	idx := make(map[string]int, len(existing))
	out := make([]Sample, len(existing))
	copy(out, existing)
	for i, s := range out {
		idx[s.MachineID+"|"+s.Metric] = i
	}
	for _, s := range incoming {
		key := s.MachineID + "|" + s.Metric
		if i, ok := idx[key]; ok {
			out[i] = s
		} else {
			idx[key] = len(out)
			out = append(out, s)
		}
	}
	return out
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
