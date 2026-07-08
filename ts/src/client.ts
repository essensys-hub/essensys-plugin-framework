// Client de l'API moderne /api/plugins/*. Lecture seule via guardedFetch.
import type { Descriptor, History, PluginInfo, Reading } from "./descriptor";
import { guardedFetch } from "./noArmoire";

export interface PluginClientOptions {
  /** base URL (défaut: origine courante). */
  baseUrl?: string;
}

export class PluginClient {
  private base: string;
  constructor(opts: PluginClientOptions = {}) {
    this.base = (opts.baseUrl ?? "").replace(/\/$/, "");
  }

  private async get<T>(path: string): Promise<T> {
    // credentials: include -> envoie le cookie de session (auth LAN/portail).
    const res = await guardedFetch(`${this.base}${path}`, { method: "GET", credentials: "include" });
    if (!res.ok) throw new Error(`${path} → HTTP ${res.status}`);
    return (await res.json()) as T;
  }

  descriptor(pluginId: string): Promise<Descriptor> {
    return this.get<Descriptor>(`/api/plugins/${pluginId}/descriptor`);
  }

  current(pluginId: string): Promise<Reading> {
    return this.get<Reading>(`/api/plugins/${pluginId}/current`);
  }

  /** Série historisée d'une métrique (48 h max) pour les courbes. */
  history(pluginId: string, metric: string, hours = 24): Promise<History> {
    return this.get<History>(
      `/api/plugins/${pluginId}/history?metric=${encodeURIComponent(metric)}&hours=${hours}`,
    );
  }

  private async post<T>(path: string): Promise<T> {
    const res = await guardedFetch(`${this.base}${path}`, { method: "POST", credentials: "include" });
    if (!res.ok) throw new Error(`${path} → HTTP ${res.status}`);
    return (await res.json()) as T;
  }

  /** Catalogue des plugins compilés (écran Paramètres). */
  list(): Promise<PluginInfo[]> {
    return this.get<PluginInfo[]>(`/api/plugins/`);
  }

  /** Active un plugin (admin). Renvoie le catalogue à jour. */
  enable(pluginId: string): Promise<PluginInfo[]> {
    return this.post<PluginInfo[]>(`/api/plugins/${pluginId}/enable`);
  }

  /** Désactive un plugin (admin), données conservées. */
  disable(pluginId: string): Promise<PluginInfo[]> {
    return this.post<PluginInfo[]>(`/api/plugins/${pluginId}/disable`);
  }

  /** Désinstalle : désactive et efface snapshot + historique (admin). */
  purge(pluginId: string): Promise<PluginInfo[]> {
    return this.post<PluginInfo[]>(`/api/plugins/${pluginId}/purge`);
  }
}
