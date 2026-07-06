#!/usr/bin/env python3
"""Lint bloquant : aucun secret en clair dans un manifest de plugin.

Un manifest ne doit exposer que des références SOPS. On refuse toute clé qui
ressemble à un secret en clair (password, passwd, token, api_key, secret...)
portant une valeur littérale, et tout champ 'sops_ref' vide.

Usage: lint_secrets.py <manifest.json> [...]  ->  code 1 si fuite détectée.
"""
import json
import re
import sys

SUSPECT = re.compile(r"(pass(word|wd)?|token|api[_-]?key|secret|credential)", re.I)


def scan(node, path=""):
    hits = []
    if isinstance(node, dict):
        for k, v in node.items():
            here = f"{path}.{k}" if path else k
            if k == "sops_ref":
                if not isinstance(v, str) or not v.strip():
                    hits.append(f"{here}: sops_ref vide")
                continue
            if SUSPECT.search(k) and isinstance(v, str) and v.strip():
                hits.append(f"{here}: valeur en clair suspecte ({v[:3]}…)")
            hits += scan(v, here)
    elif isinstance(node, list):
        for i, v in enumerate(node):
            hits += scan(v, f"{path}[{i}]")
    return hits


def main(argv):
    if len(argv) < 2:
        print("usage: lint_secrets.py <manifest.json> ...", file=sys.stderr)
        return 2
    ok = True
    for path in argv[1:]:
        try:
            with open(path, encoding="utf-8") as f:
                data = json.load(f)
        except (OSError, json.JSONDecodeError) as e:
            print(f"✗ {path}: {e}")
            ok = False
            continue
        hits = scan(data)
        if hits:
            ok = False
            print(f"✗ {path}")
            for h in hits:
                print(f"    - {h}")
        else:
            print(f"✓ {path}")
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
