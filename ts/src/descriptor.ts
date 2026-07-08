// Types du descripteur server-driven. Miroir exact des types Go du SDK
// (plugin.go). Le backend décrit l'UI ; le renderer générique l'affiche à
// l'identique sur les deux frontends jumeaux.

export type Tone = "solar" | "grid" | "load" | "battery" | "";

export interface MetricDisplay {
  name: string;
  label: string;
  unit: string;
  tone?: Tone;
}

export interface TileSpec {
  icon: string;
  /** nom de la métrique mise en avant dans la tuile */
  primary: string;
}

export interface PageSpec {
  chart: "flow" | "area" | "gauge";
}

/** Sous-ligne d'une carte KPI : "label valeur unité". */
export interface SubRef {
  label?: string;
  metric: string;
}

/** Carte KPI du tableau de bord. */
export interface CardSpec {
  label: string;
  icon?: string; // "sun" | "home" | "arrow-up" | "battery"
  tone?: Tone;
  value_tone?: Tone;
  metric: string;
  sub?: SubRef[];
  sub_text?: string;
}

/** Jauge circulaire de ratio (ex. autoconsommation). */
export interface GaugeSpec {
  title: string;
  numerator: string;
  denominator: string;
  invert?: boolean;
  label?: string;
  legend_a?: string;
  legend_b?: string;
  tone?: Tone;
}

export interface StatRef {
  label: string;
  metric?: string;
  peak?: boolean;
  tone?: Tone;
}

/** Courbe du jour d'une métrique (route history). */
export interface ChartSpec {
  title: string;
  metric: string;
  unit?: string;
  tone?: Tone;
  stats?: StatRef[];
}

/** Tableau de bord riche server-driven (cartes, jauge, courbe). */
export interface DashboardSpec {
  cards?: CardSpec[];
  gauge?: GaugeSpec;
  chart?: ChartSpec;
}

export interface Descriptor {
  plugin_id: string;
  title: string;
  tile?: TileSpec;
  page?: PageSpec;
  dashboard?: DashboardSpec;
  metrics: MetricDisplay[];
  read_only: boolean;
}

/** Point historisé pour les courbes. */
export interface Point {
  ts: string;
  value: number;
}

export interface History {
  metric: string;
  points: Point[];
}

export interface Sample {
  metric: string;
  value: number;
  unit: string;
  machine_id: string;
  ts: string;
}

export interface Reading {
  plugin_id: string;
  samples: Sample[];
  updated_at: string;
  stale: boolean;
}

/** Couleur sémantique alignée sur la charte Essensys (tokens --essensys-*). */
export const TONE_VAR: Record<Tone, string> = {
  solar: "var(--essensys-warning, #f59e0b)",
  grid: "var(--essensys-primary, #2563eb)",
  load: "var(--essensys-primary, #2563eb)",
  battery: "var(--essensys-success, #16a34a)",
  "": "var(--essensys-text-main, #111827)",
};
