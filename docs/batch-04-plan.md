# Batch 04 execution plan (issue #297, EDHREC ranks ~472–577)

Status: blocked pending catalog file contents. Attempt 1 failed on edit-format
compliance; Attempt 2 ran without a recorded baseline. This file records the
required procedure so the next run is verifiable.

## Baseline (must be recorded before any edit)

1. Run the full test suite on the target branch (`go test ./...` server-side,
   `npm test` client-side) and record pass/fail per test.
2. If the suite is already failing, isolate or fix pre-existing failures
   first and document the clean baseline before proceeding.

## Chunking

100 cards, split into chunks of 20–25, aligned with the slices in #297:

- Chunk 1: Slice 297-b (Landfall, 16 cards)
- Chunk 2: Slice 297-c + 297-d (Lands, 17 cards)
- Chunk 3: Slice 297-e (Mixed / ready, 11 cards) + start of 297-f
- Chunk 4: Slice 297-f remainder (Token makers)
- Chunk 5: Slice 297-g (Tribal: Clerics)
- Chunk 6: Slice 297-h (Tribal: Spirits)

Each chunk is emitted strictly as FILE/SEARCH/REPLACE blocks — no prose, no
fenced diffs inside the patch payload. After each chunk, re-run the test
suite and diff against the baseline; revert and retry on any regression.

## Prerequisites for the card edits themselves

The card catalog source files were not available to the patch author. Needed
before card definitions can be written:

1. An existing card-definition file showing the current `effects.Register` /
   `Spec` pattern (ideally a recently merged batch-03 card: one ETB token
   maker and one land).
2. The target file(s) where batch-04 cards should be added.
3. The card list with oracle IDs for ranks 472–577 (`retriage-slices.json`
   or the Scryfall data slice).

## Known state from the issue re-triage (2026-09-24)

- 52 of 100 cards already catalogued — no action needed.
- 47 buildable now, 0 blocked, 1 unsure (Birgi, God of Storytelling —
  'boast' not tracked in roadmap registry.go; needs a human check).
- Slice 297-a (16 cards) already merged in #1786.
