package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wayfarer's Bauble — Artifact {1}:
//
//	"{2}, {T}, Sacrifice Wayfarer's Bauble: Search your library for a
//	basic land card, put it onto the battlefield tapped, then
//	shuffle."
//
// A one-mana rock that turns into a Rampant Growth later. The tap
// component matters here in a way it doesn't on Burnished Hart: the
// Bauble must have been on the battlefield since the start of the
// turn is NOT required (it isn't a creature, so no summoning
// sickness), but it must be untapped, so it can't be used the same
// turn something else tapped it.
func init() {
	Register(Spec{
		OracleID: "31f15274-301b-47c5-ba19-0ced04520878",
		Name:     "Wayfarer's Bauble",
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice Wayfarer's Bauble: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Cost: game.AbilityCost{
				Tap:           true,
				SacrificeSelf: true,
				Mana:          "{2}",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Wayfarer's Bauble — a basic land",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
