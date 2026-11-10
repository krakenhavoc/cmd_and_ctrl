# draftpacks testdata

- `fixture_dump.json` is a small file in the shape of Scryfall's
  default-cards dump. Its two sets, `c2a` and `c2b`, are **made up for
  the tests**: the cards are real cards the catalog holds as Full (real
  oracle IDs and names), but the set, the rarities, the collector numbers
  and the printing IDs are invented, and the basic lands and the
  `C2 Fixture …` cards have invented oracle IDs. It is not Scryfall data
  and says nothing about any real set.
- `draftable_fixture.txt` is the ledger of the fixture's draftable sets.
  CI has no Scryfall dump, so this is the ledger CI checks: it fails when
  a card of `c2a` stops being Full in the catalog.
- `draftable_sets.txt` is the ledger of the real draftable sets, checked
  when `CMDCTRL_SCRYFALL_DUMP` points at the real dump (nightly).

Both ledgers are generated. Regenerate one only on purpose, from
`server/`:

    go test ./internal/draftpacks/ -run TestDraftableLedger -update-draftable
    CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json \
      go test ./internal/draftpacks/ -run TestDraftableLedger -update-draftable

The first rewrites the fixture ledger, the second the real one.
