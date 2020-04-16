#!/bin/bash
CONTRACT=$1
RAND=$2
UPDATE_ECOSTATE="{\"update_ecostate\":{\"ecostate\": $2}}"
xrncli tx wasm execute $CONTRACT "$UPDATE_ECOSTATE" --gas auto --fees 5000utree --from oracle --chain-id kontraua --node http://159.89.249.168:26657 -y