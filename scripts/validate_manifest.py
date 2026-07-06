#!/usr/bin/env python3
"""Valide un plugin.manifest.json contre le schéma Essensys.

Sans dépendance externe : reproduit les contraintes clés du JSON Schema
(schema/plugin.manifest.schema.json) — champs requis, enums, règle
device-LAN ⇒ pas de périmètre armoire-wan, et interdiction de secret en clair.

Usage: validate_manifest.py <manifest.json> [<manifest.json> ...]
Sortie: code 0 si tous valides, 1 sinon.
"""
import json
import sys

REQUIRED = ["id", "manifest_version", "framework_version", "capabilities", "perimeters", "surfaces", "visibility"]
CAPS = {"metrics", "device-poll", "cloud-relay", "ui-tile", "ui-page", "settings"}
PERIMETERS = {"lan-cm5", "hub-cloudsync", "armoire-wan"}
ROLES = {"user", "admin_local", "admin_global", "lan_user", "lan_admin"}
ID_OK = lambda s: isinstance(s, str) and s[:1].isalpha() and all(c.islower() or c.isdigit() or c == "-" for c in s) and not s.endswith("-")


def validate(path):
    errs = []
    try:
        with open(path, encoding="utf-8") as f:
            m = json.load(f)
    except (OSError, json.JSONDecodeError) as e:
        return [f"illisible: {e}"]

    for k in REQUIRED:
        if k not in m:
            errs.append(f"champ requis manquant: {k}")
    if "id" in m and not ID_OK(m["id"]):
        errs.append(f"id invalide (kebab-case attendu): {m.get('id')!r}")
    if m.get("manifest_version", 0) < 1:
        errs.append("manifest_version doit être >= 1")
    for c in m.get("capabilities", []):
        if c not in CAPS:
            errs.append(f"capability inconnue: {c}")
    perims = set(m.get("perimeters", []))
    for p in perims:
        if p not in PERIMETERS:
            errs.append(f"perimeter inconnu: {p}")
    if not perims:
        errs.append("perimeters ne peut être vide")
    for r in m.get("visibility", []):
        if r not in ROLES:
            errs.append(f"rôle de visibilité inconnu: {r}")
    if not m.get("visibility"):
        errs.append("visibility ne peut être vide")
    if m.get("write_scope", "read-only") != "read-only":
        errs.append("write_scope doit être 'read-only' (MVP)")

    # Règle périmètre: un collecteur device-LAN exclut armoire-wan.
    coll = (m.get("surfaces") or {}).get("collector")
    if coll and coll.get("runs_on") in ("cm5", "gateway") and "armoire-wan" in perims:
        errs.append("plugin device-LAN incompatible avec le périmètre armoire-wan")

    # Secrets: uniquement par référence SOPS.
    for s in m.get("secrets", []):
        if "sops_ref" not in s or not s.get("sops_ref"):
            errs.append(f"secret {s.get('key')!r} sans sops_ref")
    return errs


def main(argv):
    if len(argv) < 2:
        print("usage: validate_manifest.py <manifest.json> ...", file=sys.stderr)
        return 2
    ok = True
    for path in argv[1:]:
        errs = validate(path)
        if errs:
            ok = False
            print(f"✗ {path}")
            for e in errs:
                print(f"    - {e}")
        else:
            print(f"✓ {path}")
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
