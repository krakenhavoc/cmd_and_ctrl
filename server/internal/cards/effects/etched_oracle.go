package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Etched Oracle — Artifact Creature — Wizard {4}, 0/0 (EDHREC rank
// 6423):
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for
//	 each color of mana spent to cast it.)
//	 {1}, Remove four +1/+1 counters from this creature: Target
//	 player draws three cards."
//
// Both halves of this PR on one card: sunburst reads the mana that
// paid (#761), and the draw ability spends the counters it made
// (#625's self form, now on a card that can actually make four).
//
// Sunburst (CR 702.44a) is SunburstCounters, and since #1002 it is a
// CR 614.1c entry clause like Hangarback Walker's X rather than an
// OnResolve: the engine seeds it onto the entry event off the
// resolving stack item, so the counters are part of the ENTRY. They
// are on the permanent before EventETB, a doubler applies, and a
// "whenever one or more counters are put on a permanent you control"
// payoff sees them.
//
// One declared simplification is left, and it is the paid-cost
// record's rather than the entry's: with strict mana off the engine
// never saw what paid, so the record claims no colours (ADR 0068 §3,
// unknown answers weaker than printed) and the Oracle enters as a 0/0
// that dies at once.
//
// Spec.WantsDistinctColors makes the cast gate spread the payment
// across colours, so an Oracle cast off four different lands really
// is a 4/4. Cast off four colourless mana it is a 0/0 and dies at
// once, which is exactly what the card does in paper.
//
// The draw ability is instant speed and has no {T}, so a freshly
// landed Oracle with four counters can cash them the turn it arrives.
func init() {
	Register(Spec{
		OracleID:            "7ecfa47e-1165-46a6-884c-290b1c14d020",
		Name:                "Etched Oracle",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			SunburstCounters(game.CounterPlusOne),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}, Remove four +1/+1 counters from this creature: Target player draws three cards.",
			Cost: Plus(
				ManaCost("{1}"),
				RemoveCountersFromThis(game.CounterPlusOne, 4),
			),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, ref := range ctx.LegalTargets() {
					if ref.Kind == game.TargetPlayer {
						return DrawCards{Player: ref.ID, N: 3}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
