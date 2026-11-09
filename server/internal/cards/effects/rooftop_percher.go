package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rooftop Percher — Creature — Shapeshifter {5}, 3/3 (Lorwyn Eclipsed,
// #4458):
//
//	"Changeling (This card is every creature type.)
//	 Flying
//	 When this creature enters, exile up to two target cards from
//	 graveyards. You gain 3 life."
//
// Changeling and flying are Gangly Stompling's declared keywords. The
// trigger is Graveyard Glutton's "up to two target cards from graveyards"
// (graveyards of any players, not a single one) with the life after the
// exile, as printed. With no target chosen the life is still gained; if
// every chosen target is illegal the ability does not resolve, so no life
// (the engine's CR 608.2b check runs before the body).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3dc73517-0767-46ae-bc8d-ee378237a509",
		Name:            "Rooftop Percher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"changeling", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Rooftop Percher — exile up to two target cards from graveyards, then you gain 3 life",
					func(g *game.Game, item *game.StackItem) error {
						if err := ExileTargetCards(g, item); err != nil {
							return err
						}
						return GainLife{Amount: 3}.Apply(NewContext(g, item))
					}),
				TargetCardInGraveyard("up to two target cards from graveyards").WithCount(0, 2)),
		},
	})
}
