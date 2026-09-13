package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burnished Hart — Artifact Creature — Elk {3}, 2/2:
//
//	"{3}, Sacrifice Burnished Hart: Search your library for up to two
//	basic land cards, put them onto the battlefield tapped, then
//	shuffle."
//
// Colourless ramp any deck can run, which is why it shows up in
// Commander lists that play no green at all. Six mana total across
// two turns for two lands is slow; being an artifact creature that
// any colour can cast is the point.
//
// Cost is mana + sacrifice with NO tap, so the Hart can cash itself
// in the turn it arrives.
//
// S22: one search with Limit 2 — the two single-card passes this
// used to make were a workaround for the first-match picker, and the
// chooser makes "up to two basic land cards" literal.
func init() {
	Register(Spec{
		OracleID:     "893fed41-c144-433f-af88-bc7d419b7fb3",
		Name:         "Burnished Hart",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{3}, Sacrifice Burnished Hart: Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
			Cost: game.AbilityCost{
				SacrificeSelf: true,
				Mana:          "{3}",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         2,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Burnished Hart — up to two basic lands",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
