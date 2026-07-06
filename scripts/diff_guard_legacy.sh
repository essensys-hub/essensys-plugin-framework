#!/usr/bin/env bash
# Diff-guard bloquant : interdit toute référence au protocole legacy IoT depuis
# du code de plugin. Les plugins vivent exclusivement côté API moderne.
#
# Usage: diff_guard_legacy.sh [chemin ...]   (défaut: répertoire courant)
# Sortie: code 1 si une référence legacy est trouvée.
set -euo pipefail

roots=("${@:-.}")

# Motifs gelés (endpoints/handlers legacy — voir okf/protocols/legacy-http.md).
patterns='serverinfos|mystatus|myactions|/api/web/actions|table.?d.?echange|BP_MQX_ETH|/done\b'

# On ignore les artefacts et la doc.
found=$(grep -RInE "$patterns" "${roots[@]}" \
  --include='*.go' --include='*.ts' --include='*.tsx' --include='*.py' \
  --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=dist 2>/dev/null || true)

if [ -n "$found" ]; then
  echo "✗ Référence legacy interdite dans du code de plugin :"
  echo "$found"
  echo ""
  echo "Les plugins doivent utiliser uniquement l'API moderne /api/plugins/*."
  exit 1
fi
echo "✓ Aucune référence legacy détectée."
