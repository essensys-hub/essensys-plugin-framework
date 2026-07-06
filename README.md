# essensys-plugin-framework

Framework de plugins Essensys — ajoute des **options/intégrations** aux quatre applications modernes jumelles (`essensys-server-backend`, `essensys-user-portal-backend`, `essensys-server-frontend`, `essensys-user-portal-frontend`) **sans forker chaque app**.

## Principes
- **Manifest déclaratif** (`plugin.manifest.json`), extension de `features/schema/feature.schema.json`.
- **Aucun chargement dynamique de code** : registre Go compilé.
- **Server-driven UI** : un renderer TS générique partagé, rendu identique LAN et cloud (jumeaux).
- Réutilise l'infra existante : **Mosquitto** (transport), **Redis** (last-value), **Prometheus** (séries), **SOPS** (secrets).
- Frontière **dual-protocol** garantie en CI : le protocole legacy IoT n'est jamais touché.

## Contenu
- `go/` — SDK Go (interface `PluginAdapter`, registre compilé, routes `/api/plugins/<id>/*`).
- `ts/` — package du renderer générique (tuile, page détail, panneau réglages).
- `schema/` — schéma du manifest de plugin.

## Références
- Spécification OpenSpec : `essensys-memory/openspec/changes/essensys-plugin-framework-2026-07-035`.
- Plugins : [`essensys-plugin-example`](https://github.com/essensys-hub/essensys-plugin-example) (template), [`essensys-plugin-sungrow`](https://github.com/essensys-hub/essensys-plugin-sungrow) (premier plugin).
