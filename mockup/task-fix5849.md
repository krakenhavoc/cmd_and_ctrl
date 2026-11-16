Branch: feat/5848-draft-drag-preview (PR #5849 into develop): pointer-event drag into Main/Sideboard and a large hover preview in the draft room (spec sections 1 and 2).

CI fails one client unit test: "tallies the picks for Main and for the whole pool, and follows a move" (find it with `grep -rn "tallies the picks for Main" client/src`). It probably still moves a card with the removed HTML5 drag and drop, or reads markup the branch changed. Decide whether the test or the code is wrong: the behaviour the test protects (stats follow a move between Main and Sideboard) must still work. Fix it, then run `npx vitest run` on the draft files (`src/lib/components/draft src/routes/Draft.render.test.ts src/lib/draft*.test.ts src/lib/draftDrag* src/lib/draftPreview*`), then the whole suite once (`npx vitest run`), `npm run lint` and `npx svelte-check`.

Also harden one gap the builder reported: if a drag-to-Sideboard pick is refused by the server, the card id queued as "aside" must not put a later card in the Sideboard; prune queued aside ids that are still in the pack (or clear them on a pick error), with a test.
Commit (`fix(client): …`), push, stop.
