package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weathered Wayfarer — Creature — Human Nomad Cleric (EDHREC rank 1354):
//
//	"{W}, {T}: Search your library for a land card, reveal it, put it
//	 into your hand, then shuffle. Activate only if an opponent
//	 controls more lands than you."
//
// White's catch-up tutor. The gate is the activation condition (CR
// 602.1b, #743): OpponentControlsMore compares each opponent's land
// count with yours, so one opponent ahead of you is enough. A {T}
// ability on a creature, so summoning sickness applies (CR 302.6). Any
// land card — a nonbasic, a legendary land, an MDFC whose front is a
// land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1093fdb6-bd4f-46a8-92e3-5aa46b52bb4f",
		Name:         "Weathered Wayfarer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{W}, {T}: Search your library for a land card, reveal it, put it into your hand, then shuffle. Activate only if an opponent controls more lands than you.",
			Cost:      Plus(ManaCost("{W}"), TapCost()),
			Condition: OpponentControlsMore(MatchLand),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b06TutorToHand("Weathered Wayfarer — a land card", func(c game.Card) bool { return c.IsLand() })(item, NewContext(g, item))
			},
		}},
	})
}
