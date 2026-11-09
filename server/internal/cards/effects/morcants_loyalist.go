package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morcant's Loyalist — Creature — Elf Warrior {1}{B}{G} (Lorwyn Eclipsed, #4458):
//
//	"Other Elves you control get +1/+1.
//	 When this creature dies, return another target Elf card from your
//	 graveyard to your hand."
func init() {
	Register(Spec{
		OracleID:     "b2c30683-cc7f-4341-add1-b08cb9b93eeb",
		Name:         "Morcant's Loyalist",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Morcant's Loyalist — return another target Elf card from your graveyard to your hand", returnTargetCardToHand),
				Another(TargetCardInGraveyard("another target Elf card from your graveyard", YouOwn(), HasSubtype("Elf")))),
		},
	})
}
