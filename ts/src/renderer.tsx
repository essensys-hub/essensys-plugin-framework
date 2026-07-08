// Renderer générique server-driven. Consommé à l'identique par
// essensys-server-frontend et essensys-user-portal-frontend (jumeaux).
// Aucun composant spécifique à un plugin : tout vient du descripteur.
import React from "react";
import { TONE_VAR } from "./descriptor";
import type {
  CardSpec,
  Descriptor,
  FlowSpec,
  GaugeSpec,
  ChartSpec,
  MetricDisplay,
  Point,
  Reading,
  Tone,
} from "./descriptor";

export interface RenderProps {
  descriptor: Descriptor;
  reading?: Reading;
  /** série pour dashboard.chart (fournie par l'hôte via client.history). */
  history?: Point[];
  /** false si le plugin n'est pas disponible sur ce périmètre. */
  available?: boolean;
}

function findSample(reading: Reading | undefined, metric: string) {
  return reading?.samples.find((x) => x.metric === metric);
}

function fmt(v: number, digits = 2): string {
  return v.toLocaleString("fr-FR", { maximumFractionDigits: digits });
}

function sampleValue(reading: Reading | undefined, metric: string): string {
  const s = findSample(reading, metric);
  if (!s) return "—";
  return `${fmt(s.value)} ${s.unit}`;
}

function toneColor(tone?: Tone): string {
  return TONE_VAR[tone ?? ""] ?? TONE_VAR[""];
}

/** Fond doux dérivé de la teinte (pastille d'icône). */
function toneSoft(tone?: Tone): string {
  return `color-mix(in srgb, ${toneColor(tone)} 14%, transparent)`;
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

/** Icônes déclarables dans le descripteur (trait 2px, style charte). */
const ICONS: Record<string, React.JSX.Element> = {
  sun: (
    <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
      <circle cx="12" cy="12" r="4" />
      <path d="M12 3v2m0 14v2M3 12h2m14 0h2M5.6 5.6 7 7m10 10 1.4 1.4M18.4 5.6 17 7M7 17l-1.4 1.4" />
    </svg>
  ),
  home: (
    <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinejoin="round">
      <path d="M4 21V10l8-7 8 7v11" />
      <rect x="10" y="14" width="4" height="7" />
    </svg>
  ),
  "arrow-up": (
    <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 19V5m0 0-6 6m6-6 6 6" />
    </svg>
  ),
  battery: (
    <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinejoin="round">
      <rect x="4" y="7" width="15" height="10" rx="2" />
      <path d="M22 10v4" />
      <rect x="6" y="9" width="11" height="6" fill="currentColor" stroke="none" />
    </svg>
  ),
};

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

function KpiCard({ card, reading }: { card: CardSpec; reading?: Reading }): React.JSX.Element {
  const s = findSample(reading, card.metric);
  const valueColor = card.value_tone ? toneColor(card.value_tone) : undefined;
  return (
    <div className="ess-kpi">
      <div className="ess-kpi__top">
        <span className="ess-kpi__label">{card.label}</span>
        {card.icon && ICONS[card.icon] && (
          <span className="ess-kpi__chip" style={{ background: toneSoft(card.tone), color: toneColor(card.tone) }}>
            {ICONS[card.icon]}
          </span>
        )}
      </div>
      <div className="ess-kpi__value" style={{ color: valueColor }}>
        {s ? (
          <>
            {fmt(s.value)}
            <small>{s.unit}</small>
          </>
        ) : (
          "—"
        )}
      </div>
      <div className="ess-kpi__sub">
        {card.sub?.length
          ? card.sub.map((ref, i) => {
              const sub = findSample(reading, ref.metric);
              return (
                <span key={ref.metric}>
                  {i > 0 && " · "}
                  {ref.label && `${ref.label} `}
                  <b>{sub ? `${fmt(sub.value)} ${sub.unit}` : "—"}</b>
                </span>
              );
            })
          : card.sub_text ?? " "}
      </div>
    </div>
  );
}

function GaugeCard({ gauge, reading }: { gauge: GaugeSpec; reading?: Reading }): React.JSX.Element {
  const num = findSample(reading, gauge.numerator)?.value;
  const den = findSample(reading, gauge.denominator)?.value;
  let pct: number | undefined;
  if (num !== undefined && den !== undefined && den > 0) {
    pct = Math.min(100, Math.max(0, (num / den) * 100));
    if (gauge.invert) pct = 100 - pct;
    pct = Math.round(pct);
  }
  const color = toneColor(gauge.tone ?? "battery");
  const C = 2 * Math.PI * 82; // circonférence r=82 (viewBox 200)
  const offset = pct === undefined ? C : C * (1 - pct / 100);
  return (
    <div className="ess-gauge">
      <div className="ess-panel-hd">
        <h4>{gauge.title}</h4>
      </div>
      <div className="ess-gauge__wrap">
        <svg viewBox="0 0 200 200" role="img" aria-label={gauge.title}>
          <circle cx="100" cy="100" r="82" fill="none" stroke="var(--ess-track, #eef1f5)" strokeWidth="20" />
          <circle
            cx="100"
            cy="100"
            r="82"
            fill="none"
            stroke={color}
            strokeWidth="20"
            strokeLinecap="round"
            transform="rotate(-90 100 100)"
            strokeDasharray={C}
            strokeDashoffset={offset}
            style={{ transition: "stroke-dashoffset .6s ease" }}
          />
        </svg>
        <div className="ess-gauge__center">
          <div className="ess-gauge__val" style={{ color }}>
            {pct === undefined ? "—" : `${pct} %`}
          </div>
          {gauge.label && <div className="ess-gauge__lab">{gauge.label}</div>}
        </div>
      </div>
      {pct !== undefined && (gauge.legend_a || gauge.legend_b) && (
        <div className="ess-gauge__legend">
          {gauge.legend_a && (
            <span>
              <i style={{ background: color }} /> {gauge.legend_a} {pct} %
            </span>
          )}
          {gauge.legend_b && (
            <span>
              <i style={{ background: "var(--ess-track, #eef1f5)" }} /> {gauge.legend_b} {100 - pct} %
            </span>
          )}
        </div>
      )}
    </div>
  );
}

/** Fenêtre du jour : 06 h → 21 h locale. */
function dayWindow(): { start: Date; end: Date } {
  const start = new Date();
  start.setHours(6, 0, 0, 0);
  const end = new Date();
  end.setHours(21, 0, 0, 0);
  return { start, end };
}

function ChartCard({
  chart,
  reading,
  history,
  pluginId,
}: {
  chart: ChartSpec;
  reading?: Reading;
  history?: Point[];
  pluginId: string;
}): React.JSX.Element {
  const { start, end } = dayWindow();
  const pts = (history ?? [])
    .map((p) => ({ t: new Date(p.ts).getTime(), v: p.value }))
    .filter((p) => p.t >= start.getTime() && p.t <= end.getTime())
    .sort((a, b) => a.t - b.t);

  const W = 560;
  const H = 220;
  const yTop = 15;
  const yBase = 205;
  const vMax = Math.max(0.5, ...pts.map((p) => p.v)) * 1.08;
  const x = (t: number) => ((t - start.getTime()) / (end.getTime() - start.getTime())) * W;
  const y = (v: number) => yBase - (v / vMax) * (yBase - yTop);
  const line = pts.map((p, i) => `${i === 0 ? "M" : "L"}${x(p.t).toFixed(1)} ${y(p.v).toFixed(1)}`).join(" ");
  const area = pts.length
    ? `${line} L${x(pts[pts.length - 1].t).toFixed(1)} ${H} L${x(pts[0].t).toFixed(1)} ${H} Z`
    : "";
  const peak = pts.reduce<{ t: number; v: number } | undefined>(
    (best, p) => (best === undefined || p.v > best.v ? p : best),
    undefined,
  );
  const color = toneColor(chart.tone ?? "solar");
  const gradId = `ess-fill-${pluginId}-${chart.metric}`;

  return (
    <div className="ess-chart">
      <div className="ess-panel-hd">
        <h4>{chart.title}</h4>
        <span className="ess-kpi__label">{chart.unit ?? ""} · 06 h → 21 h</span>
      </div>
      {pts.length >= 2 ? (
        <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" role="img" aria-label={chart.title}>
          <defs>
            <linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0" stopColor={color} stopOpacity="0.34" />
              <stop offset="1" stopColor={color} stopOpacity="0" />
            </linearGradient>
          </defs>
          {[55, 105, 155].map((gy) => (
            <line key={gy} x1="0" y1={gy} x2={W} y2={gy} stroke="var(--ess-grid-line, #eceef1)" />
          ))}
          <path d={area} fill={`url(#${gradId})`} />
          <path d={line} fill="none" stroke={color} strokeWidth="2.6" strokeLinecap="round" />
          {peak && (
            <>
              <circle cx={x(peak.t)} cy={y(peak.v)} r="4.5" fill={color} />
              <circle cx={x(peak.t)} cy={y(peak.v)} r="9" fill={color} opacity=".18" />
            </>
          )}
        </svg>
      ) : (
        <p className="ess-chart__empty">Historique en construction — la courbe apparaîtra au fil de la journée.</p>
      )}
      {chart.stats?.length ? (
        <div className="ess-chart__meta">
          {chart.stats.map((st) => {
            let value = "—";
            let unit = "";
            let sub = st.label;
            if (st.peak) {
              if (peak) {
                value = fmt(peak.v);
                unit = chart.unit ?? "";
                const d = new Date(peak.t);
                sub = `${st.label} à ${d.getHours()} h ${String(d.getMinutes()).padStart(2, "0")}`;
              }
            } else if (st.metric) {
              const s = findSample(reading, st.metric);
              if (s) {
                value = fmt(s.value);
                unit = s.unit;
              }
            }
            return (
              <div className="ess-chart__stat" key={st.label}>
                <div className="ess-chart__stat-v" style={{ color: st.tone ? toneColor(st.tone) : undefined }}>
                  {value} <small>{unit}</small>
                </div>
                <div className="ess-chart__stat-l">{sub}</div>
              </div>
            );
          })}
        </div>
      ) : null}
    </div>
  );
}

/** Valeur numérique d'une métrique du snapshot (0 si absente). */
function metricNum(reading: Reading | undefined, metric?: string): number {
  if (!metric) return 0;
  return findSample(reading, metric)?.value ?? 0;
}

const FLOW_EPS = 0.05; // kW : en-dessous, le lien est considéré inactif

function FlowLink({
  d,
  color,
  active,
  reverse,
}: {
  d: string;
  color: string;
  active: boolean;
  reverse?: boolean;
}): React.JSX.Element {
  return (
    <>
      <path d={d} fill="none" stroke="var(--ess-track, #eef1f5)" strokeWidth="6" strokeLinecap="round" />
      {active && (
        <path
          d={d}
          fill="none"
          stroke={color}
          strokeWidth="4"
          strokeLinecap="round"
          className={reverse ? "ess-flow__dash ess-flow__dash--rev" : "ess-flow__dash"}
        />
      )}
    </>
  );
}

function FlowValue({ x, y, v }: { x: number; y: number; v: number }): React.JSX.Element | null {
  if (v <= FLOW_EPS) return null;
  return (
    <text x={x} y={y} fontSize="15" fontWeight="700" fill="var(--essensys-text-main, #111827)">
      {fmt(v)}
      <tspan fontSize="11" fill="var(--essensys-text-muted, #6b7280)">
        {" "}
        kW
      </tspan>
    </text>
  );
}

/** Schéma de flux d'énergie : PV / maison / batterie / réseau (maquette vue 1). */
function FlowCard({ flow, reading }: { flow: FlowSpec; reading?: Reading }): React.JSX.Element {
  const pv = metricNum(reading, flow.pv);
  const load = metricNum(reading, flow.load);
  const gridIn = metricNum(reading, flow.grid_import);
  const gridOut = metricNum(reading, flow.grid_export);
  const battIn = metricNum(reading, flow.battery_charge);
  const battOut = metricNum(reading, flow.battery_discharge);
  const soc = flow.battery_soc ? findSample(reading, flow.battery_soc)?.value : undefined;

  const solar = toneColor("solar");
  const primary = toneColor("grid");
  const success = toneColor("battery");
  const text = "var(--essensys-text-main, #111827)";
  const muted = "var(--essensys-text-muted, #6b7280)";
  const faint = "var(--essensys-text-faint, #9ca3af)";
  const nodeBg = "var(--essensys-bg-card, #fff)";
  const nodeRing = "var(--essensys-border, #e2e8f0)";

  const pvToHouse = Math.min(pv, load);
  const battLabel =
    battOut > FLOW_EPS ? `décharge ${fmt(battOut)} kW` : battIn > FLOW_EPS ? `charge ${fmt(battIn)} kW` : "0 W";
  const gridLabel = gridOut > FLOW_EPS ? "RÉSEAU · INJECTION" : gridIn > FLOW_EPS ? "RÉSEAU · SOUTIRAGE" : "RÉSEAU";
  const gridValue = gridOut > FLOW_EPS ? gridOut : gridIn;

  return (
    <div className="ess-flow">
      <svg viewBox="0 0 640 420" role="img" aria-label="Schéma de flux d'énergie">
        <defs>
          <filter id="ess-flow-soft" x="-20%" y="-20%" width="140%" height="140%">
            <feDropShadow dx="0" dy="6" stdDeviation="7" floodColor="#0b1220" floodOpacity="0.10" />
          </filter>
        </defs>

        {/* PV -> Maison */}
        <FlowLink d="M320 150 L320 250" color={solar} active={pvToHouse > FLOW_EPS} />
        <FlowValue x={332} y={205} v={pvToHouse} />
        {/* PV -> Réseau (injection) */}
        <FlowLink d="M392 118 H520 V288" color={success} active={gridOut > FLOW_EPS} />
        {gridOut > FLOW_EPS && <FlowValue x={430} y={106} v={gridOut} />}
        {/* Batterie <-> Maison */}
        <FlowLink
          d="M214 331 H256"
          color={success}
          active={battOut > FLOW_EPS || battIn > FLOW_EPS}
          reverse={battIn > FLOW_EPS && battOut <= FLOW_EPS}
        />
        {/* Réseau -> Maison (soutirage) */}
        <FlowLink d="M456 331 H384" color={primary} active={gridIn > FLOW_EPS} />
        {gridIn > FLOW_EPS && <FlowValue x={396} y={318} v={gridIn} />}

        {/* Nœud PV */}
        <g filter="url(#ess-flow-soft)">
          <rect x="256" y="70" width="128" height="80" rx="16" fill={nodeBg} stroke={nodeRing} />
        </g>
        <g transform="translate(276,86)">
          <rect x="0" y="0" width="48" height="34" rx="4" fill="none" stroke={solar} strokeWidth="2.4" />
          <path d="M12 0v34M24 0v34M36 0v34M0 11.3h48M0 22.6h48" stroke={solar} strokeWidth="1.6" opacity=".75" />
        </g>
        <text x="332" y="100" fontSize="20" fontWeight="750" fill={text} letterSpacing="-.5">
          {fmt(pv)}
          <tspan fontSize="12" fill={muted} fontWeight="650">
            {" "}
            kW
          </tspan>
        </text>
        <text x="332" y="122" fontSize="11" fill={faint} fontWeight="700" letterSpacing=".5">
          PANNEAUX PV
        </text>

        {/* Nœud Maison */}
        <g filter="url(#ess-flow-soft)">
          <rect x="256" y="250" width="128" height="86" rx="16" fill={nodeBg} stroke={nodeRing} />
        </g>
        <g transform="translate(276,268)" stroke={primary} strokeWidth="2.4" fill="none" strokeLinejoin="round">
          <path d="M4 20 L20 6 L36 20" />
          <path d="M8 18v14h24V18" />
          <rect x="16" y="22" width="8" height="10" fill="color-mix(in srgb, currentColor 12%, transparent)" />
        </g>
        <text x="332" y="292" fontSize="20" fontWeight="750" fill={text} letterSpacing="-.5">
          {fmt(load)}
          <tspan fontSize="12" fill={muted} fontWeight="650">
            {" "}
            kW
          </tspan>
        </text>
        <text x="332" y="314" fontSize="11" fill={faint} fontWeight="700" letterSpacing=".5">
          MAISON
        </text>

        {/* Nœud Batterie */}
        <g filter="url(#ess-flow-soft)">
          <rect x="86" y="288" width="128" height="86" rx="16" fill={nodeBg} stroke={nodeRing} />
        </g>
        <g transform="translate(104,306)" stroke={success} strokeWidth="2.4" fill="none">
          <rect x="4" y="2" width="30" height="34" rx="4" />
          <path d="M14 0h10" strokeLinecap="round" />
          <rect x="8" y="24" width="22" height="8" fill={success} stroke="none" />
          {soc !== undefined && soc > 40 && <rect x="8" y="14" width="22" height="7" fill={success} opacity=".55" stroke="none" />}
        </g>
        <text x="150" y="330" fontSize="20" fontWeight="750" fill={text} letterSpacing="-.5">
          {soc === undefined ? "—" : fmt(soc, 1)}
          <tspan fontSize="12" fill={muted} fontWeight="650">
            {" "}
            %
          </tspan>
        </text>
        <text x="150" y="352" fontSize="11" fill={faint} fontWeight="700" letterSpacing=".5">
          BATTERIE · {battLabel}
        </text>

        {/* Nœud Réseau */}
        <g filter="url(#ess-flow-soft)">
          <rect x="456" y="288" width="128" height="86" rx="16" fill={nodeBg} stroke={nodeRing} />
        </g>
        <g transform="translate(486,306)" stroke={primary} strokeWidth="2.2" fill="none" strokeLinecap="round">
          <path d="M14 0 L4 34 M14 0 L24 34 M7 12 H21 M6 20 H22 M2 34 H26" />
        </g>
        <text x="520" y="330" fontSize="20" fontWeight="750" fill={text} letterSpacing="-.5">
          {fmt(gridValue)}
          <tspan fontSize="12" fill={muted} fontWeight="650">
            {" "}
            kW
          </tspan>
        </text>
        <text x="520" y="352" fontSize="11" fill={faint} fontWeight="700" letterSpacing=".5">
          {gridLabel}
        </text>
      </svg>
      <div className="ess-flow__legend">
        <span>
          <i style={{ background: solar }} /> Production
        </span>
        <span>
          <i style={{ background: primary }} /> Consommation
        </span>
        <span>
          <i style={{ background: success }} /> Injection
        </span>
      </div>
    </div>
  );
}

/** Tableau de bord riche décrit par descriptor.dashboard. */
function PluginDashboard({ descriptor, reading, history }: RenderProps): React.JSX.Element {
  const dash = descriptor.dashboard!;
  return (
    <div className="ess-dash">
      {dash.cards?.length ? (
        <div className="ess-dash__cards">
          {dash.cards.map((c) => (
            <KpiCard key={c.metric + c.label} card={c} reading={reading} />
          ))}
        </div>
      ) : null}
      {(dash.gauge || dash.chart) && (
        <div className="ess-dash__lower">
          {dash.gauge && <GaugeCard gauge={dash.gauge} reading={reading} />}
          {dash.chart && (
            <ChartCard chart={dash.chart} reading={reading} history={history} pluginId={descriptor.plugin_id} />
          )}
        </div>
      )}
    </div>
  );
}

type PanelView = "dash" | "flow";

function initialView(pluginId: string): PanelView {
  try {
    return window.localStorage.getItem(`ess-plugin-view-${pluginId}`) === "flow" ? "flow" : "dash";
  } catch {
    return "dash";
  }
}

/** Page/panneau détail : dashboard riche si déclaré, sinon liste des métriques. */
export function PluginPanel({ descriptor, reading, history, available = true }: RenderProps): React.JSX.Element {
  const [view, setView] = React.useState<PanelView>(() => initialView(descriptor.plugin_id));
  const dash = descriptor.dashboard;
  const hasFlow = Boolean(dash?.flow);
  const hasDash = Boolean(dash?.cards?.length || dash?.gauge || dash?.chart);

  const switchView = (v: PanelView) => {
    setView(v);
    try {
      window.localStorage.setItem(`ess-plugin-view-${descriptor.plugin_id}`, v);
    } catch {
      /* stockage indisponible : la préférence ne survit pas au rechargement */
    }
  };

  if (!available) return <Unavailable title={descriptor.title} />;
  const showFlow = hasFlow && (view === "flow" || !hasDash);
  return (
    <section className="ess-plugin ess-plugin--panel" data-plugin={descriptor.plugin_id}>
      <header>
        <h3>{descriptor.title}</h3>
        {reading?.stale && <StaleBadge />}
        {hasFlow && hasDash && (
          <div className="ess-plugin__views" role="tablist" aria-label="Vue du plugin">
            <button
              type="button"
              role="tab"
              aria-selected={!showFlow}
              className={showFlow ? "" : "is-active"}
              onClick={() => switchView("dash")}
            >
              Tableau
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={showFlow}
              className={showFlow ? "is-active" : ""}
              onClick={() => switchView("flow")}
            >
              Schéma
            </button>
          </div>
        )}
        {descriptor.read_only && <span className="ess-plugin__ro">lecture seule</span>}
      </header>
      {dash ? (
        showFlow && dash.flow ? (
          <div className="ess-dash">
            <FlowCard flow={dash.flow} reading={reading} />
          </div>
        ) : (
          <PluginDashboard descriptor={descriptor} reading={reading} history={history} />
        )
      ) : (
        <dl className="ess-plugin__metrics">
          {descriptor.metrics.map((m: MetricDisplay) => (
            <div key={m.name} className="ess-plugin__metric">
              <dt style={{ color: TONE_VAR[m.tone ?? ""] }}>{m.label}</dt>
              <dd>{sampleValue(reading, m.name)}</dd>
            </div>
          ))}
        </dl>
      )}
    </section>
  );
}
