#!/bin/bash

CNF_DIR="${1}"

if [ ! -d "$CNF_DIR" ]; then
  echo "Directory '$CNF_DIR' not found." >&2
  exit 1
fi

exit_code=0

for file in "$CNF_DIR"/*.cnf; do
  [ -e "$file" ] || continue

  if ! go run satsolver -f "$file" -c 0; then
    exit_code=1
  fi
done

exit $exit_code
