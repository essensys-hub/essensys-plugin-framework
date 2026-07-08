package plugin

import (
	"testing"
	"time"
)

// Un message MQTT ne porte qu'une série : Put doit fusionner avec le snapshot
// courant (upsert par machine|metric), pas le remplacer.
func TestMemStorePutMergesSeries(t *testing.T) {
	s := NewMemStore(time.Minute)
	s.Put("p", "m1", []Sample{{Metric: "a", Value: 1, MachineID: "m1"}}, time.Now())
	s.Put("p", "m1", []Sample{{Metric: "b", Value: 2, MachineID: "m1"}}, time.Now())
	s.Put("p", "m1", []Sample{{Metric: "a", Value: 3, MachineID: "m1"}}, time.Now())

	r := s.Current("p")
	if len(r.Samples) != 2 {
		t.Fatalf("attendu 2 séries après fusion, obtenu %d", len(r.Samples))
	}
	byMetric := map[string]float64{}
	for _, sm := range r.Samples {
		byMetric[sm.Metric] = sm.Value
	}
	if byMetric["a"] != 3 {
		t.Fatalf("la série a doit être upsertée à 3, obtenu %v", byMetric["a"])
	}
	if byMetric["b"] != 2 {
		t.Fatalf("la série b doit être conservée à 2, obtenu %v", byMetric["b"])
	}
}

func TestMergeSamplesDistinctMachines(t *testing.T) {
	out := MergeSamples(
		[]Sample{{Metric: "a", Value: 1, MachineID: "m1"}},
		[]Sample{{Metric: "a", Value: 9, MachineID: "m2"}},
	)
	if len(out) != 2 {
		t.Fatalf("les machines distinctes sont des séries distinctes, obtenu %d", len(out))
	}
}
