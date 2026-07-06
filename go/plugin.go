// Package plugin est le SDK Go du framework de plugins Essensys.
//
// Principe: inversion de dépendances. Le SDK ne dépend d'aucun client MQTT,
// Redis ou Prometheus concret. Les applications hôtes (essensys-server-backend,
// essensys-user-portal-backend) fournissent les implémentations de Bus, Store,
// MetricSink et AuthFunc. Cela garde le SDK compilable et testable hors-ligne,
// et interdit par construction tout couplage au protocole legacy.
package plugin

import "time"

// FrameworkVersion est la version majeure.mineure du contrat exposé par ce SDK.
const FrameworkVersion = "1.0.0"

// Perimeter est un périmètre de déploiement Essensys.
type Perimeter string

const (
	PerimeterLANCM5       Perimeter = "lan-cm5"
	PerimeterHubCloudsync Perimeter = "hub-cloudsync"
	PerimeterArmoireWAN   Perimeter = "armoire-wan"
)

// Role réutilise le RBAC existant des applications.
type Role string

const (
	RoleUser        Role = "user"
	RoleAdminLocal  Role = "admin_local"
	RoleAdminGlobal Role = "admin_global"
	RoleLANUser     Role = "lan_user"
	RoleLANAdmin    Role = "lan_admin"
)

// Sample est une valeur de métrique pour une armoire donnée.
type Sample struct {
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	MachineID string    `json:"machine_id"`
	TS        time.Time `json:"ts"`
}

// Descriptor est la description server-driven de l'UI d'un plugin. Les deux
// frontends jumeaux rendent ce descripteur avec le même renderer générique.
type Descriptor struct {
	PluginID string          `json:"plugin_id"`
	Title    string          `json:"title"`
	Tile     *TileSpec       `json:"tile,omitempty"`
	Page     *PageSpec       `json:"page,omitempty"`
	Metrics  []MetricDisplay `json:"metrics"`
	ReadOnly bool            `json:"read_only"`
}

// MetricDisplay décrit comment afficher une métrique (label, unité, couleur sémantique).
type MetricDisplay struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Unit  string `json:"unit"`
	Tone  string `json:"tone,omitempty"` // "solar" | "grid" | "load" | "battery" | ""
}

// TileSpec décrit la tuile compacte affichée dans l'accueil.
type TileSpec struct {
	Icon    string `json:"icon"`
	Primary string `json:"primary"` // nom de la métrique mise en avant
}

// PageSpec décrit la page détail (série historique via MetricSink).
type PageSpec struct {
	Chart string `json:"chart"` // "flow" | "area" | "gauge"
}

// PluginAdapter est le contrat compilé d'un plugin côté backend.
// Aucun chargement dynamique: les adaptateurs sont enregistrés au build.
type PluginAdapter interface {
	// ID doit correspondre au champ id du manifest.
	ID() string
	// Descriptor renvoie la description UI server-driven.
	Descriptor() Descriptor
	// OnMessage traduit un message du bus en échantillons à persister.
	OnMessage(msg BusMessage) ([]Sample, error)
}
