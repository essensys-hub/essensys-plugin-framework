// Client de l'API moderne /api/plugins/*. Lecture seule via guardedFetch.
import { Descriptor, Reading } from "./descriptor";
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
    const res = await guardedFetch(`${this.base}${path}`, { method: "GET" });
    if (!res.ok) throw new Error(`${path} → HTTP ${res.status}`);
    return (await res.json()) as T;
  }

  descriptor(pluginId: string): Promise<Descriptor> {
    return this.get<Descriptor>(`/api/plugins/${pluginId}/descriptor`);
  }

  current(pluginId: string): Promise<Reading> {
    return this.get<Reading>(`/api/plugins/${pluginId}/current`);
  }
}
