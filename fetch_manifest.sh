#!/bin/bash

# DIM's API key. insanity_wolf.jpg: it's public anyways.
API_KEY=5ec01caf6aee450d9dabe646294ffdc9


path="$(curl --silent --header 'X-API-Key: $(API_KEY)' https://www.bungie.net/Platform/Destiny2/Manifest/ \
  | jq --raw-output .Response.jsonWorldContentPaths.en)"
url="https://www.bungie.net${path}"
curl --silent --header 'X-API-Key: $(API_KEY)' --header 'Content-Type: application/json' "${url}" > ${1:-manifest.json}
