package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lux Artillery — Artifact {4} (EDHREC rank 3706):
//
//	"Whenever you cast an artifact creature spell, it gains sunburst.
//	 (It enters with a +1/+1 counter on it for each color of mana
//	 spent to cast it.)
//	 At the beginning of your end step, if there are thirty or more
//	 counters among artifacts and creatures you control, this
//	 artifact deals 10 damage to each opponent."
//
// The counters deck's alternate win. The end-step half is Lathiel's
// shape: the count — every counter of every kind on the artifacts
// and creatures the controller controls, a permanent that is both
// counted once — is a CR 603.4 intervening-if, checked when the
// trigger would go on the stack and again on resolution, so the
// trigger never appears below thirty and a board shrunk in response
// stops the damage. The Artillery is the damage source, so a
// Fog-class shield stops it.
//
// SANDBOX SIMPLIFICATION — the sunburst grant is NOT implemented.
// Sunburst counts the colours of mana spent to cast the spell, and
// the cost engine discards the tokens it pays with and reports
// nothing about their colours — the same missing seam that keeps
// Scaled Nurturer's rider and converge out of the catalog. An
// artifact creature cast with the Artillery out enters with no
// sunburst counters. Weaker than printed, never stronger: the
// end-step half, the card's payoff, is whole.
func init() {
	Register(Spec{
		OracleID:     "cbd76b22-d04e-48e7-bcab-c7f5466d67d8",
		Name:         "Lux Artillery",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Artifact creature spells you cast don't gain sunburst — only the end-step 10 damage at thirty counters works."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b35EndStepAndThirtyCounters(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Lux Artillery — 10 damage to each opponent",
					b35TenDamageToEachOpponentIfThirtyCounters)
			},
		}},
	})
}
