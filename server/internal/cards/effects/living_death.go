package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Living Death — Sorcery {3}{B}{B} (EDHREC rank 460):
//
//	"Each player exiles all creature cards from their graveyard, then
//	 sacrifices all creatures they control, then puts all cards they
//	 exiled this way onto the battlefield."
//
// The reanimator deck's board swap: everything in every graveyard
// trades places with everything on every battlefield. Three passes,
// each over a snapshot taken before anything moves, in the printed
// order across the whole table:
//
//  1. every seat's graveyard creature cards → exile, as one batch
//     (ExileCardsThenForEffect);
//  2. every creature on the battlefield → sacrificed, as one batch
//     (SacrificeAllThenForEffect) — a sacrifice, not a destroy, so
//     indestructible doesn't save it and "whenever you sacrifice"
//     payoffs fire;
//  3. the cards exiled in step 1 → the battlefield under their
//     OWNER's control, as ONE entry for every player
//     (ReturnFromExileTogether with Controller zero).
//
// The creatures sacrificed in step 2 land in graveyards and stay
// there — they were not exiled "this way" — which is the whole
// trick of the card. Everything that comes back is a new object
// (fresh InstanceID, ETB triggers fire), as printed.
//
// #894: step 1 is ONE BATCH with steps 2 and 3 hanging off its
// continuation, because an exile can PAUSE. A commander card in any
// graveyard stops to ask its owner about the command zone (CR 903.9),
// and the old per-card loop walked straight past the question: the
// swap sacrificed and reanimated with the prompt still open, so when
// the owner finally declined, their commander landed in exile with
// nothing left to return it. Now nothing moves in steps 2 and 3 until
// every leg of step 1 has settled, and step 3 returns the LANDED list
// — the cards that really reached exile (CR 400.7), which is exactly
// what "all cards they exiled this way" means. A commander that takes
// the command zone is not among them and does not come back, which is
// the right answer rather than a stranded card.
//
// #910: step 2 is a BATCH too, and step 3 hangs off ITS continuation.
// Two things follow. The creatures leave as one simultaneous exit, so a
// Blood Artist caught in the swap sees every death including its own
// (CR 603.10) instead of only the ones after it; and a sacrificed
// COMMANDER stops the reanimation until its owner has answered CR 903.9,
// where before the return pass ran with the question still open. The
// exiled list is carried into the inner continuation BY VALUE, the
// contract every continuation in the engine signs, so an undo across
// either prompt replays identically.
//
// Step 3 reads nothing of step 2's result — "all cards THEY EXILED this
// way" is step 1's list — so the sacrifice batch's landed list is
// deliberately ignored. It waits for it, it does not read it.
//
// #1872: step 3 is ONE simultaneous entry for the whole table, not a
// loop of single returns. Each player puts their cards onto the
// battlefield at the same time (CR 101.4: the actions happen
// simultaneously), so every returning creature sees every other one
// enter (CR 603.6a) — two returned Soul Wardens each gain a life for
// the other. A loop announced the first card before the second moved.
// An entry that stops for a prompt (a Clone asked what to copy)
// holds the whole batch until it is answered, and then everything
// lands together; the cards wait in exile meanwhile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e6a3df4-67a3-452e-a6ef-f04dbadb21ef",
		Name:         "Living Death",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// BOTH sides are snapshotted before anything moves, and
			// the graveyard side before the exile in particular: the
			// creatures sacrificed below land in graveyards, and a set
			// re-gathered afterwards would reanimate them too.
			var dead []uuid.UUID
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.IsCreature() {
						dead = append(dead, c.InstanceID)
					}
				}
			}
			doomed := ctx.CreatureIDs()
			// The context is rebuilt inside the continuation from the
			// live *Game, the contract every continuation in the engine
			// follows: an undo restores this game's fields in place, so
			// a captured *Game would be the wrong one.
			return ctx.Game.ExileCardsThenForEffect(dead, func(g *game.Game, exiled []uuid.UUID) error {
				// A creature that left while the exile was paused is
				// skipped by the batch rather than routed: the list was
				// taken before the first move, and a leg with nothing
				// to do must not open a window for a move that cannot
				// happen.
				return g.SacrificeAllThenForEffect(item.SourceCardID, doomed, func(g *game.Game, _ []uuid.UUID) error {
					return ReturnFromExileTogether{Targets: exiled}.Apply(NewContext(g, item))
				})
			})
		},
	})
}
