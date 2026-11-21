package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morcant's Loyalist — Creature — Elf Warrior {1}{B}{G} (Lorwyn Eclipsed, #4458):
//
//	"Other Elves you control get +1/+1.
//	 When this creature dies, return another target Elf card from your
//	 graveyard to your hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2c30683-cc7f-4341-add1-b08cb9b93eeb",
		Name:         "Morcant's Loyalist",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Morcant's Loyalist — return another target Elf card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.Another(effects.TargetCardInGraveyard("another target Elf card from your graveyard", effects.YouOwn(), effects.HasSubtype("Elf")))),
		},
	})
}
