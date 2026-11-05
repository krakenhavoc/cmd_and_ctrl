---
title: "A token creation after a replaced move"
date: 2026-11-05
issues: [4522]
---
**A token creation after a replaced move** (#4522, [ADR 0178](decisions/0178-life-instructions-after-a-replaced-move.md) amendment of 2026-11-05, CR 608.2c, CR 614.6, CR 616.1) — "instead exile that card and create a 2/2 black Zombie creature token". The replacement declares the creation on the move it modifies (`Game.CreateTokensAfterReplacedMoveForEffect`), as it declares a life change, and the move's finisher creates the tokens through the token pipeline once the card has landed. The token is part of the modified event, so nothing goes on the stack. The move's continuation (the next leg of a wipe, the destroying spell's next instruction) waits for the creation and every token's entry, including a CR 616 question on either.
- **The rest of the move** is parked on the game while the creation runs (`Game.movesAwaitingTokens`, a carrier event holding the route and the instructions not yet run), and the creation's keyed tail takes it back, the way an entry choice's frame is frozen behind its discard. The creation survives a later redirect or cancel of the move (CR 614.6) and is skipped for a controller who has left the game.
- **Snapshot:** additive in v9. A paused move with a declared creation writes `replacementResume.event.tokensAfterMove` under its own key, `replacement-resume/move-tokens`; a parked move is written as `movesAwaitingTokens`. **Closure ratchet:** three new edges, keyed and rebuilt; no counter moves.
- **Cards (1, Full):** Kalitas, Traitor of Ghet.
- **Tests:** `game/move_tokens_test.go` (landing order, a creation prompt, an entry prompt, a wipe, a cancelled move, a cost, a controller who left, undo, a parked move restored, a paused move restored both ways, a stale key refused, malformed declarations) and `cards/effects/kalitas_traitor_of_ghet_test.go`.
- **Still open:** a wipe that destroys the replacement's source loses it for the later legs (#4521). A move paid as a cost creates its tokens without asking, but a token's entry during that cost can still ask.
