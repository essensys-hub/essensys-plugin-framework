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

export interface Descriptor {
  plugin_id: string;
  title: string;
  tile?: TileSpec;
  page?: PageSpec;
  metrics: MetricDisplay[];
  read_only: boolean;
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
