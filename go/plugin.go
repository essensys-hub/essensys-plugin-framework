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
	PluginID  string          `json:"plugin_id"`
	Title     string          `json:"title"`
	Tile      *TileSpec       `json:"tile,omitempty"`
	Page      *PageSpec       `json:"page,omitempty"`
	Dashboard *DashboardSpec  `json:"dashboard,omitempty"`
	Metrics   []MetricDisplay `json:"metrics"`
	ReadOnly  bool            `json:"read_only"`
}

// DashboardSpec décrit le tableau de bord riche d'un plugin (cartes KPI,
// jauge de ratio, courbe du jour). Entièrement server-driven : le renderer
// générique ne connaît aucun plugin.
type DashboardSpec struct {
	Cards []CardSpec `json:"cards,omitempty"`
	Gauge *GaugeSpec `json:"gauge,omitempty"`
	Chart *ChartSpec `json:"chart,omitempty"`
}

// CardSpec est une carte KPI : valeur principale + sous-ligne.
type CardSpec struct {
	Label   string   `json:"label"`
	Icon    string   `json:"icon,omitempty"` // "sun" | "home" | "arrow-up" | "battery"
	Tone    string   `json:"tone,omitempty"` // teinte de la pastille ; la valeur suit ValueTone
	ValueTone string `json:"value_tone,omitempty"`
	Metric  string   `json:"metric"`             // valeur principale
	Sub     []SubRef `json:"sub,omitempty"`      // sous-ligne "label valeur"
	SubText string   `json:"sub_text,omitempty"` // sous-ligne statique sinon
}

// SubRef référence une métrique affichée en sous-ligne d'une carte.
type SubRef struct {
	Label  string `json:"label,omitempty"`
	Metric string `json:"metric"`
}

// GaugeSpec est une jauge circulaire de ratio entre deux métriques du
// snapshot. pourcent = numerator/denominator borné 0..100 ; Invert affiche
// 1-ratio (ex. autoconsommation = 1 - injecté_du_jour/produit_du_jour).
type GaugeSpec struct {
	Title       string `json:"title"`
	Numerator   string `json:"numerator"`
	Denominator string `json:"denominator"`
	Invert      bool   `json:"invert,omitempty"`
	Label       string `json:"label,omitempty"`    // libellé sous la valeur
	LegendA     string `json:"legend_a,omitempty"` // part affichée
	LegendB     string `json:"legend_b,omitempty"` // part complémentaire
	Tone        string `json:"tone,omitempty"`
}

// ChartSpec est la courbe du jour d'une métrique (source: route history).
type ChartSpec struct {
	Title  string    `json:"title"`
	Metric string    `json:"metric"`
	Unit   string    `json:"unit,omitempty"`
	Tone   string    `json:"tone,omitempty"`
	Stats  []StatRef `json:"stats,omitempty"` // méta sous la courbe
}

// StatRef est une statistique affichée sous la courbe. Si Peak est vrai, le
// renderer calcule le pic de la série (valeur + heure) au lieu de lire une
// métrique du snapshot.
type StatRef struct {
	Label  string `json:"label"`
	Metric string `json:"metric,omitempty"`
	Peak   bool   `json:"peak,omitempty"`
	Tone   string `json:"tone,omitempty"`
}

// Point est un échantillon historisé (courbes).
type Point struct {
	TS    time.Time `json:"ts"`
	Value float64   `json:"value"`
}

// HistoryProvider est un port optionnel du Store : s'il est implémenté, la
// route /api/plugins/<id>/history est exposée pour alimenter les courbes.
type HistoryProvider interface {
	History(pluginID, metric string, since time.Time) []Point
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
