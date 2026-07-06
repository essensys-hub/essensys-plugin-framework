// Renderer générique server-driven. Consommé à l'identique par
// essensys-server-frontend et essensys-user-portal-frontend (jumeaux).
// Aucun composant spécifique à un plugin : tout vient du descripteur.
import React from "react";
import { TONE_VAR } from "./descriptor";
import type { Descriptor, Reading, MetricDisplay } from "./descriptor";

export interface RenderProps {
  descriptor: Descriptor;
  reading?: Reading;
  /** false si le plugin n'est pas disponible sur ce périmètre. */
  available?: boolean;
}

function sampleValue(reading: Reading | undefined, metric: string): string {
  const s = reading?.samples.find((x) => x.metric === metric);
  if (!s) return "—";
  const v = Number.isInteger(s.value) ? s.value : Math.round(s.value * 100) / 100;
  return `${v} ${s.unit}`;
}

function Unavailable({ title }: { title: string }): React.JSX.Element {
  return (
    <div className="ess-plugin ess-plugin--off" role="note">
      <b>{title}</b>
      <p>Indisponible sur ce périmètre d'installation.</p>
    </div>
  );
}

function StaleBadge(): React.JSX.Element {
  return (
    <span className="ess-plugin__stale" title="Dernière valeur connue">
      obsolète
    </span>
  );
}

/** Tuile compacte affichée dans l'accueil. */
export function PluginTile({ descriptor, reading, available = true }: RenderProps): React.JSX.Element {
  if (!available) return <Unavailable title={descriptor.title} />;
  const primary = descriptor.tile?.primary ?? descriptor.metrics[0]?.name ?? "";
  const tone = descriptor.metrics.find((m) => m.name === primary)?.tone ?? "";
  return (
    <div className="ess-plugin ess-plugin--tile" data-plugin={descriptor.plugin_id}>
      <header>
        <span className="ess-plugin__title">{descriptor.title}</span>
        {reading?.stale && <StaleBadge />}
      </header>
      <div className="ess-plugin__primary" style={{ color: TONE_VAR[tone] }}>
        {sampleValue(reading, primary)}
      </div>
    </div>
  );
}

/** Page/panneau détail : toutes les métriques déclarées. */
export function PluginPanel({ descriptor, reading, available = true }: RenderProps): React.JSX.Element {
  if (!available) return <Unavailable title={descriptor.title} />;
  return (
    <section className="ess-plugin ess-plugin--panel" data-plugin={descriptor.plugin_id}>
      <header>
        <h3>{descriptor.title}</h3>
        {reading?.stale && <StaleBadge />}
        {descriptor.read_only && <span className="ess-plugin__ro">lecture seule</span>}
      </header>
      <dl className="ess-plugin__metrics">
        {descriptor.metrics.map((m: MetricDisplay) => (
          <div key={m.name} className="ess-plugin__metric">
            <dt style={{ color: TONE_VAR[m.tone ?? ""] }}>{m.label}</dt>
            <dd>{sampleValue(reading, m.name)}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
