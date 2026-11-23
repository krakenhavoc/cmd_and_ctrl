package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elderfang Ritualist — Creature — Elf Cleric {2}{B}, 3/1 (Kaldheim,
// #5505):
//
//	"When this creature dies, return another target Elf card from your
//	 graveyard to your hand."
//
// Morcant's Loyalist's dies trigger without its anthem: "another" is
// object identity (Another), so the Ritualist, an Elf card in the
// graveyard by then, can't return itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29c1f80e-f7e2-425e-a225-e88351fcbaab",
		Name:         "Elderfang Ritualist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Elderfang Ritualist — return another target Elf card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.Another(effects.TargetCardInGraveyard("another target Elf card from your graveyard", effects.YouOwn(), effects.HasSubtype("Elf")))),
		},
	})
}
