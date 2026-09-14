package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza's Bauble — Artifact {0} (EDHREC rank 2840):
//
//	"{T}, Sacrifice this artifact: Look at a card at random in target
//	 player's hand. You draw a card at the beginning of the next
//	 turn's upkeep."
//
// The free artifact that cantrips a turn late — an artifact count
// for Mox Opal and a storm count for the price of nothing. The
// activation is a tap-and-sacrifice ability targeting a player; at
// resolution the activator is marked a knower of one card in that
// hand ("look at", not "reveal" — Gitaxian Probe's posture, so the
// rest of the table learns nothing), and a CR 603.7 delayed trigger
// is scheduled for the very next upkeep at the table, whoever's it
// is (Arcane Denial's "the next turn's upkeep"). The draw goes on
// the stack when that upkeep begins, so it can be responded to, and
// it is controlled by the Bauble's controller (CR 603.7d).
//
// Sandbox simplification, declared: "at random" is a deterministic
// pick keyed on the game log (b27PseudoRandomIndex) rather than a
// draw from the engine's seeded RNG, which no effect can reach. The
// activator cannot steer it — they do not know the hand's order —
// and it never favours the most recently drawn card the way a fixed
// index would.
func init() {
	Register(Spec{
		OracleID:     "17dbbca3-ac1c-4d4e-9618-3e66ac3ccd24",
		Name:         "Urza's Bauble",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The card you look at is picked by a fixed rule from the game's history rather than truly at random."},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Look at a card at random in target player's hand. You draw a card at the beginning of the next turn's upkeep.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						b27LookAtRandomCardInHand(ctx, item.Controller, t.ID)
					}
				}
				return ScheduleDelayedTrigger{
					At:     game.StepUpkeep,
					Label:  "Urza's Bauble — draw a card",
					Effect: b27DrawOne,
				}.Apply(ctx)
			},
		}},
	})
}
