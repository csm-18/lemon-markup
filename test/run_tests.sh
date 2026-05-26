#!/usr/bin/env bash
set -u

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/bin/lm"
PROJECTS_DIR="$ROOT/test/projects"

mkdir -p "$ROOT/bin"

echo "Building lm binary..."
if ! go build -o "$BIN" "$ROOT"; then
  echo "Failed to build lm. Aborting."
  exit 2
fi

total=0
failures=0

for proj in "$PROJECTS_DIR"/*; do
  [ -d "$proj" ] || continue
  total=$((total+1))
  name=$(basename "$proj")
  echo
  echo "=== Testing: $name ==="
  rm -rf "$proj/dist"

  pushd "$proj" > /dev/null
  if "$BIN" compile > compile.log 2>&1; then
    status=0
  else
    status=1
  fi
  popd > /dev/null

  expect_err=0
  case "$name" in
    *_err) expect_err=1 ;;
  esac

  if [ "$expect_err" -eq 1 ]; then
    if [ "$status" -eq 0 ]; then
      echo "FAIL: expected compilation error but succeeded"
      failures=$((failures+1))
    else
      echo "PASS: compilation failed as expected"
    fi
    continue
  fi

  # For _ok projects ensure compile succeeded and outputs for files with DOCTYPE exist
  if [ "$status" -ne 0 ]; then
    echo "FAIL: compilation failed"
    echo "--- log ---"
    sed -n '1,200p' "$proj/compile.log" || true
    failures=$((failures+1))
    continue
  fi

  # Determine if project contains any DOCTYPEs
  project_has_doctype=0
  if grep -I -q "DOCTYPE" "$proj"/*.lm >/dev/null 2>&1; then
    project_has_doctype=1
  fi

  # If project has no DOCTYPE anywhere, expect no generated outputs
  if [ "$project_has_doctype" -eq 0 ]; then
    if [ -d "$proj/dist" ] && find "$proj/dist" -type f -name '*.html' | read; then
      echo "FAIL: unexpected outputs generated for project without DOCTYPE"
      failures=$((failures+1))
    else
      echo "PASS: $name (no DOCTYPE, no outputs)"
    fi
    continue
  fi

  # Project has at least one DOCTYPE: check per-file expectations
  pass=1
  while IFS= read -r -d '' file; do
    rel=${file#"$proj/"}
    if grep -qi "DOCTYPE" "$file"; then
      out="$proj/dist/${rel%.lm}.html"
      if [ ! -f "$out" ]; then
        echo "FAIL: expected output for $rel but not found -> $out"
        pass=0
      fi
    else
      out="$proj/dist/${rel%.lm}.html"
      if [ -f "$out" ]; then
        echo "FAIL: unexpected output for $rel (no DOCTYPE present) -> $out"
        pass=0
      fi
    fi
  done < <(find "$proj" -maxdepth 1 -type f -name '*.lm' -print0)

  if [ "$pass" -eq 1 ]; then
    echo "PASS: $name"
  else
    failures=$((failures+1))
  fi
done

echo
echo "Ran $total project(s). Failures: $failures"
if [ "$failures" -ne 0 ]; then
  exit 1
fi

echo "All tests passed."
