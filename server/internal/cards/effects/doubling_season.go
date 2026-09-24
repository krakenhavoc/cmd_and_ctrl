package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doubling Season — "If an effect would create one or more tokens
// under your control, it creates twice that many of those tokens
// instead. If an effect would put one or more counters on a
// permanent you control, it puts twice that many of those counters
// on that permanent instead."
//
// First S17 catalog card with a replacement effect. It shipped with
// the counter half only, because token creation was not a replaceable
// event; #762 made it one and the second half landed here, which is
// the whole card.
//
// How the pipeline sees it:
//   - Watches EventCounterPlaced (pre-event, via
//     game.RepEventCounter).
//   - AppliesTo: the counter target is a permanent the Season's
//     controller controls. Any counter name (+1/+1, loyalty, charge,
//     etc.) — "counters" in CR is uncategorised.
//   - Replace: ev.CounterDelta *= 2.
//
// "If AN EFFECT would put" is the narrow half, and ADR 0056 Decision 5
// is where it starts to matter: COMBAT DAMAGE is a turn-based action,
// not an effect, so the -1/-1 counters a wither or infect attacker puts
// on a blocker are NOT doubled, while the same keyword on a spell, or a
// fight, or a proliferate, is. Judges are consistent on this and the
// ADR cites them. The gate is ev.CounterFromCombatDamage, which the
// damage tail sets; nothing sets it yet, so this arm is a no-op change
// today and the right answer the day the tail branches.
//
// The passive "would be put" cards in the catalog — Hardened Scales,
// Winding Constrictor, Primal Vigor — name no effect and are
// deliberately NOT gated the same way.
//
// "WOULD PUT" is a placement, not a removal (CR 122.6, CR 614.1): a
// negative ev.CounterDelta (a counter coming OFF) is gated out before
// anything else runs. #1291 found this half missing — a "remove N
// counters" cost or effect was getting doubled, and two removal
// replacements on the same board could pause on a CR 616 ordering
// prompt that a removal should never see.
//
// And the token half (CR 701.7b, #762):
//
//   - Watches EventTokenCreated (pre-event, via
//     game.RepEventCreateTokens), which is one CREATION INSTRUCTION
//     rather than one token, so "create two Treasures" is one event
//     that becomes four Treasures.
//   - AppliesTo: the creation is under the Season's controller.
//   - Replace: every group's count doubles.
//
// Multi-replacement ordering (CR 616): with Hardened Scales also in
// play, the affected player (counter target's controller) picks the
// order — [HS, DS] → (1+1)*2 = 4; [DS, HS] → (1*2)+1 = 3. The CR
// 616 order prompt queues; the exit-criterion test asserts both
// results in effects/doubling_season_test.go.
func init() {
	Register(Spec{
		OracleID:     "01546b7d-a233-4176-8843-d732074dc5b6",
		Name:         "Doubling Season",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			TokensDoubled("Doubling Season: double tokens"),
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventCounter {
						return false
					}
					if ev.CounterDelta <= 0 {
						// "If AN EFFECT WOULD PUT" — CR 122.6 / CR
						// 614.1 only replace the named event. A
						// removal is not a placement (#1291).
						return false
					}
					if ev.CounterFromCombatDamage {
						// Combat damage is a turn-based action, not
						// "an effect" — see the card comment.
						return false
					}
					target, ok := g.LookupCardForEffect(ev.CounterTarget)
					if !ok {
						return false
					}
					return target.Controller == src.Controller
				},
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.CounterDelta *= 2
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Doubling Season: double counters",
			},
		},
	})
}
