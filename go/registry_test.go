package plugin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeAdapter struct{ id string }

func (f fakeAdapter) ID() string { return f.id }
func (f fakeAdapter) Descriptor() Descriptor {
	return Descriptor{PluginID: f.id, Title: "Test", Tile: &TileSpec{Primary: "pv_power"}}
}
func (f fakeAdapter) OnMessage(msg BusMessage) ([]Sample, error) {
	return []Sample{{Metric: "pv_power", Value: 5.81, Unit: "kW", MachineID: msg.MachineID}}, nil
}

func testManifest() Manifest {
	m := Manifest{
		ID: "test-plugin", ManifestVersion: 1, FrameworkVersion: "^1.0",
		Capabilities: []string{"metrics"}, Perimeters: []Perimeter{PerimeterLANCM5},
		Visibility: []Role{RoleUser}, WriteScope: "read-only",
	}
	m.Surfaces.Backend = &struct {
		Adapter string `json:"adapter"`
	}{Adapter: "test-plugin"}
	return m
}

func newTestRegistry() (*Registry, *MemStore, *MemSink) {
	store := NewMemStore(2 * time.Second)
	sink := NewMemSink()
	r := New(store, sink)
	r.Register(fakeAdapter{id: "test-plugin"})
	return r, store, sink
}

func TestConfigureRejectsUnknownAdapter(t *testing.T) {
	r := New(NewMemStore(time.Second), NewMemSink())
	if err := r.Configure(testManifest()); err == nil {
		t.Fatal("attendu: erreur adaptateur absent")
	}
}

func TestConfigureRejectsIncompatibleFramework(t *testing.T) {
	r, _, _ := newTestRegistry()
	m := testManifest()
	m.FrameworkVersion = "^2.0"
	if err := r.Configure(m); err == nil {
		t.Fatal("attendu: erreur version framework incompatible")
	}
}

func TestIngestPersistsAndKeyIsStable(t *testing.T) {
	r, store, sink := newTestRegistry()
	if err := r.Configure(testManifest()); err != nil {
		t.Fatalf("configure: %v", err)
	}
	a := fakeAdapter{id: "test-plugin"}
	msg := BusMessage{PluginID: "test-plugin", MachineID: "A254"}
	r.Ingest(a, msg)
	r.Ingest(a, msg)

	// clé de série idempotente (plugin|machine|metric) : une seule série, 2 points.
	key := "test-plugin|A254|pv_power"
	if got := len(sink.Series); got != 1 {
		t.Fatalf("séries attendues=1, got=%d (%v)", got, sink.Series)
	}
	if got := len(sink.Series[key]); got != 2 {
		t.Fatalf("points attendus=2 pour %s, got=%d", key, got)
	}
	cur := store.Current("test-plugin")
	if len(cur.Samples) != 1 || cur.Samples[0].Value != 5.81 {
		t.Fatalf("current inattendu: %+v", cur)
	}
	if cur.Stale {
		t.Fatal("lecture fraîche marquée stale")
	}
}

func TestStaleAfterDeadline(t *testing.T) {
	store := NewMemStore(2 * time.Second)
	base := time.Now()
	store.now = func() time.Time { return base }
	store.Put("p", "A254", []Sample{{Metric: "x", Value: 1}}, base.Add(-5*time.Second))
	if !store.Current("p").Stale {
		t.Fatal("lecture ancienne devrait être stale")
	}
}

func TestDisableKeepsSeries(t *testing.T) {
	r, _, sink := newTestRegistry()
	_ = r.Configure(testManifest())
	r.Ingest(fakeAdapter{id: "test-plugin"}, BusMessage{MachineID: "A254"})
	r.Disable("test-plugin")
	if r.Enabled("test-plugin") {
		t.Fatal("plugin devrait être désactivé")
	}
	if len(sink.Series) == 0 {
		t.Fatal("les séries doivent être conservées après désactivation")
	}
}

func TestRBAC(t *testing.T) {
	r, _, _ := newTestRegistry()
	_ = r.Configure(testManifest()) // visibility = [user]
	mux := http.NewServeMux()

	roles := []Role{RoleAdminGlobal} // pas dans la visibilité
	r.Mount(mux, func(*http.Request) (Identity, bool) { return Identity{Roles: roles}, true })

	// hors visibilité -> 403
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/plugins/test-plugin/current", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("attendu 403, got %d", rec.Code)
	}

	// dans la visibilité -> 200
	roles = []Role{RoleUser}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/plugins/test-plugin/descriptor", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("attendu 200, got %d", rec.Code)
	}
}
