# work/s82-cards — data only, never merge

`s82-cards.json.gz` is the Scryfall default-cards dump (refreshed 2026-11-04) cut down to every printing of every card in the S82 ten (MAT, MOM/MOC, ONE/ONC, BRO/BRC/BRR, DMU/DMC, SNC/NCC, NEO/NEC, VOW/VOC, MID/MIC, AFR/AFC, and their token sets), plus the related tokens and all their printings. Same format as `data/scryfall/default-cards.json`.

Use it in a cloud worker:

    git fetch origin work/s82-cards
    mkdir -p data/scryfall && git show FETCH_HEAD:s82-cards.json.gz | gunzip > data/scryfall/default-cards.json
    cd server && CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json go test -p 1 ./internal/cards/coverage/ -run TestOracleFixtureCoversRegistry -update-oracle -oracle-ids=<id>,<id>

Only cards from these sets are in it; a dump-wide audit run against it will report other cards as missing.

The pinned Comprehensive Rules text (AGENTS.md §6) is `cr.txt`: `git show FETCH_HEAD:cr.txt > data/cr/cr.txt`.
