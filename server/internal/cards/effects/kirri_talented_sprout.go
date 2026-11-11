package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kirri, Talented Sprout — Legendary Creature — Plant Druid {1}{R}{G}{W}, 0/3
// (Outlaws of Thunder Junction Commander, #5368):
//
//	"Other Plants and Treefolk you control get +2/+0.
//	 At the beginning of each of your postcombat main phases, return target
//	 Plant, Treefolk, or land card from your graveyard to your hand."
//
// The lord half is a tribal anthem (Canopy Tactician's shape) over two
// creature types, read through the layered view; Kirri herself is excluded.
// The trigger is "each of your postcombat main phases", so an added
// postcombat main phase fires it again (AtYourPostcombatMain, not the second-main
// variant). The target is any card in your graveyard with the Plant or
// Treefolk subtype or the land type, a card of any of the three; with none the
// trigger is removed rather than prompting (CR 603.3d), and one that left in
// response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f03bc648-08e4-4ecd-b339-ff72b00367c2",
		Name:         "Kirri, Talented Sprout",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Plant", "Treefolk"}, Others: true, YoursOnly: true}, 2, 0),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourPostcombatMain("Kirri, Talented Sprout — return target Plant, Treefolk, or land card from your graveyard to your hand",
					b35ReturnChosenGraveyardCardToHand),
				TargetCardInGraveyard("target Plant, Treefolk, or land card from your graveyard", cmdr2B15PlantTreefolkOrLand(), YouOwn())),
		},
	})
}

// cmdr2B15PlantTreefolkOrLand is "a Plant, Treefolk, or land card".
func cmdr2B15PlantTreefolkOrLand() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsLand() || c.HasSubtype("Plant") || c.HasSubtype("Treefolk")
	}
}
