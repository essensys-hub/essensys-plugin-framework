// Garde no-armoire : par construction, un plugin ne peut émettre aucune
// mutation domotique vers l'armoire. Toute tentative est bloquée sauf dry-run
// explicite. Le renderer n'utilise que guardedFetch.

const ARMOIRE_MUTATIONS: RegExp[] = [
  /\/api\/admin\/inject/,
  /\/api\/portal\/inject/,
  /\/api\/web\/actions/,
  /\/scenarios\/[^/]+\/launch/,
];

const MUTATING = new Set(["POST", "PUT", "PATCH", "DELETE"]);

export class ArmoireMutationBlocked extends Error {}

/** Vérifie qu'un appel ne mute pas l'armoire (hors dry-run). */
export function assertNoArmoire(url: string, method = "GET", dryRun = false): void {
  if (dryRun) return;
  if (!MUTATING.has(method.toUpperCase())) return;
  if (ARMOIRE_MUTATIONS.some((re) => re.test(url))) {
    throw new ArmoireMutationBlocked(
      `Mutation armoire bloquée pour un plugin (lecture seule) : ${method} ${url}`
    );
  }
}

/** fetch encapsulé : refuse toute mutation armoire. À utiliser dans tout plugin. */
export function guardedFetch(input: string, init?: RequestInit): Promise<Response> {
  const method = (init?.method || "GET").toUpperCase();
  const dryRun = Boolean((init as any)?.dryRun);
  assertNoArmoire(input, method, dryRun);
  return fetch(input, init);
}
