package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var idRe = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)

// SecretRef référence un secret via SOPS (jamais en clair).
type SecretRef struct {
	Key     string `json:"key"`
	SOPSRef string `json:"sops_ref"`
}

// Manifest est la représentation Go de plugin.manifest.json.
type Manifest struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Version          string      `json:"version,omitempty"`     // version du plugin (release)
	Description      string      `json:"description,omitempty"` // une phrase pour l'écran Paramètres
	ManifestVersion  int         `json:"manifest_version"`
	FrameworkVersion string      `json:"framework_version"`
	Capabilities     []string    `json:"capabilities"`
	Perimeters       []Perimeter `json:"perimeters"`
	Surfaces         Surfaces    `json:"surfaces"`
	Visibility       []Role      `json:"visibility"`
	Secrets          []SecretRef `json:"secrets"`
	WriteScope       string      `json:"write_scope"`
}

// Surfaces décrit les briques activées par le plugin.
type Surfaces struct {
	Collector *struct {
		Runtime          string `json:"runtime"`
		RunsOn           string `json:"runs_on"`
		HeartbeatSeconds int    `json:"heartbeat_seconds"`
	} `json:"collector,omitempty"`
	Backend *struct {
		Adapter string `json:"adapter"`
	} `json:"backend,omitempty"`
	UI *struct {
		Tile     bool `json:"tile"`
		Page     bool `json:"page"`
		Settings bool `json:"settings"`
	} `json:"ui,omitempty"`
}

// LoadManifest lit et valide un manifest depuis un fichier.
func LoadManifest(path string) (Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return Manifest{}, fmt.Errorf("manifest illisible: %w", err)
	}
	return m, m.Validate()
}

// Validate applique les règles minimales (le schéma JSON reste la référence en CI).
func (m Manifest) Validate() error {
	if !idRe.MatchString(m.ID) {
		return fmt.Errorf("id invalide %q (kebab-case attendu)", m.ID)
	}
	if m.ManifestVersion < 1 {
		return fmt.Errorf("manifest_version manquant")
	}
	if m.FrameworkVersion == "" {
		return fmt.Errorf("framework_version manquant")
	}
	if len(m.Perimeters) == 0 {
		return fmt.Errorf("perimeters manquant")
	}
	if len(m.Visibility) == 0 {
		return fmt.Errorf("visibility manquant")
	}
	if m.WriteScope != "" && m.WriteScope != "read-only" {
		return fmt.Errorf("write_scope %q non autorisé (MVP: read-only)", m.WriteScope)
	}
	// Un plugin device-LAN (collecteur sur cm5/gateway) ne peut pas supporter armoire-wan.
	if c := m.Surfaces.Collector; c != nil && (c.RunsOn == "cm5" || c.RunsOn == "gateway") {
		if m.supportsPerimeter(PerimeterArmoireWAN) {
			return fmt.Errorf("plugin device-LAN incompatible avec le périmètre armoire-wan")
		}
	}
	return nil
}

func (m Manifest) supportsPerimeter(p Perimeter) bool {
	for _, x := range m.Perimeters {
		if x == p {
			return true
		}
	}
	return false
}

// CompatibleFramework vérifie que la version majeure requise est satisfaite.
func (m Manifest) CompatibleFramework() bool {
	want := strings.TrimPrefix(m.FrameworkVersion, "^")
	wantMajor := strings.SplitN(want, ".", 2)[0]
	haveMajor := strings.SplitN(FrameworkVersion, ".", 2)[0]
	return wantMajor == haveMajor
}

// Allows indique si un des rôles fournis figure dans la visibilité du plugin.
func (m Manifest) Allows(roles []Role) bool {
	set := map[Role]bool{}
	for _, r := range m.Visibility {
		set[r] = true
	}
	for _, r := range roles {
		if set[r] {
			return true
		}
	}
	return false
}

// HeartbeatSeconds renvoie le seuil de heartbeat déclaré (défaut 30s).
func (m Manifest) HeartbeatSeconds() int {
	if c := m.Surfaces.Collector; c != nil && c.HeartbeatSeconds > 0 {
		return c.HeartbeatSeconds
	}
	return 30
}
