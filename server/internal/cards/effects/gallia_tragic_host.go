package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallia, Tragic Host — Legendary Creature — Zombie Satyr {1}{B}, 2/1:
//
//	"Menace
//	 {4}{B}, Exile another creature card from your graveyard: Return
//	 this card from your graveyard to the battlefield tapped with a
//	 +1/+1 counter on her."
//
// Reassembling Skeleton's graveyard activation (Zones: graveyard) with
// Drivnod's exile-another-creature-card cost component (the source is
// never a legal pick). The counter is a self-replacement on entry that
// applies only when the move is out of a graveyard, so a Gallia cast
// from hand enters as the plain 2/1; it is on the card before any state
// check sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58285b70-0d13-4723-b5a4-91fb6f9ccf03",
		Name:            "Gallia, Tragic Host",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Replacements: []game.ReplacementEffect{
			rfCreatureBEntersWithCounterFromGraveyard(game.CounterPlusOne, 1,
				"Gallia, Tragic Host: enters with a +1/+1 counter"),
		},
		Activated: []ActivatedAbility{{
			Label: "{4}{B}, Exile another creature card from your graveyard: Return this card from your graveyard to the battlefield tapped with a +1/+1 counter on her.",
			Cost: Plus(ManaCost("{4}{B}"),
				ExileFromGraveyard(1, "another creature card", func(c game.Card) bool { return c.IsCreature() })),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnThisFromGraveyardTapped(g, item)
			},
		}},
	})
}
