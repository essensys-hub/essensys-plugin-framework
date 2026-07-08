// Gestionnaire de plugins pour l'écran Paramètres. Générique et partagé par
// les deux frontends jumeaux : liste le catalogue compilé, active/désactive,
// désinstalle (purge des données). L'installation et la mise à jour du code
// passent par une release (les plugins sont compilés, pas de chargement
// dynamique) — l'écran l'explique plutôt que de le simuler.
import React from "react";
import { PluginClient } from "./client";
import type { PluginInfo } from "./descriptor";

export interface PluginManagerProps {
  /** true si l'utilisateur courant peut administrer (sinon lecture seule). */
  canAdmin?: boolean;
  client?: PluginClient;
}

export function PluginManager({ canAdmin = false, client }: PluginManagerProps): React.JSX.Element {
  const api = React.useMemo(() => client ?? new PluginClient(), [client]);
  const [plugins, setPlugins] = React.useState<PluginInfo[] | undefined>();
  const [error, setError] = React.useState<string | undefined>();
  const [busy, setBusy] = React.useState<string | undefined>();
  const [confirming, setConfirming] = React.useState<string | undefined>();

  React.useEffect(() => {
    let alive = true;
    api
      .list()
      .then((l) => alive && (setPlugins(l), setError(undefined)))
      .catch((e) => alive && setError(String(e)));
    return () => {
      alive = false;
    };
  }, [api]);

  const run = (id: string, action: "enable" | "disable" | "purge") => {
    setBusy(id);
    setConfirming(undefined);
    api[action](id)
      .then((l) => {
        setPlugins(l);
        setError(undefined);
      })
      .catch((e) => setError(String(e)))
      .finally(() => setBusy(undefined));
  };

  if (error && !plugins) {
    return <p className="ess-pm__note">Catalogue de plugins indisponible ({error}).</p>;
  }
  if (!plugins) return <p className="ess-pm__note">Chargement…</p>;

  return (
    <div className="ess-pm">
      {error && <p className="ess-pm__error">{error}</p>}
      <ul className="ess-pm__list">
        {plugins.map((p) => (
          <li key={p.id} className="ess-pm__item">
            <div className="ess-pm__id">
              <div className="ess-pm__name">
                {p.name}
                {p.version && <span className="ess-pm__version">{p.version}</span>}
                <span className={p.enabled ? "ess-pm__badge ess-pm__badge--on" : "ess-pm__badge"}>
                  {p.enabled ? "Activé" : "Désactivé"}
                </span>
                {p.write_scope === "read-only" && <span className="ess-pm__ro">lecture seule</span>}
              </div>
              {p.description && <p className="ess-pm__desc">{p.description}</p>}
              <p className="ess-pm__meta">
                <code>{p.id}</code>
                {p.capabilities?.length ? ` · ${p.capabilities.join(", ")}` : ""}
              </p>
            </div>
            {canAdmin && (
              <div className="ess-pm__actions">
                {p.enabled ? (
                  <button type="button" disabled={busy === p.id} onClick={() => run(p.id, "disable")}>
                    Désactiver
                  </button>
                ) : (
                  <button
                    type="button"
                    className="ess-pm__primary"
                    disabled={busy === p.id}
                    onClick={() => run(p.id, "enable")}
                  >
                    Activer
                  </button>
                )}
                {confirming === p.id ? (
                  <span className="ess-pm__confirm">
                    Effacer aussi les données ?
                    <button type="button" className="ess-pm__danger" disabled={busy === p.id} onClick={() => run(p.id, "purge")}>
                      Confirmer
                    </button>
                    <button type="button" onClick={() => setConfirming(undefined)}>Annuler</button>
                  </span>
                ) : (
                  <button type="button" className="ess-pm__danger" disabled={busy === p.id} onClick={() => setConfirming(p.id)}>
                    Désinstaller…
                  </button>
                )}
              </div>
            )}
          </li>
        ))}
      </ul>
      <p className="ess-pm__note">
        Les plugins sont compilés dans la passerelle : l'installation d'un nouveau plugin ou la mise à
        jour de version se fait par une release Essensys (déploiement), puis se gère ici.
      </p>
    </div>
  );
}
