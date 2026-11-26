package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lashwrithe — Artifact — Equipment {4} (Commander 2014):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +1/+1 for each Swamp you control.
//	 Equip {B/P}{B/P} ({B/P} can be paid with either {B} or 2 life.)"
//
// "Swamp" is the land TYPE, so a Dryad Arbor-style Swamp counts and a
// Swamp turned into an Island does not (Blanchwood Armor's reading).
// "You" is the Equipment's controller (CR 109.5).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "066ba81e-b32d-4808-97ab-0b5f52d557e4",
		Name:         "Lashwrithe",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.LivingWeapon("Lashwrithe")},
		Static:       []game.StaticAbility{effects.PumpAttachedPer(1, 1, swampsControlledBy)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{B/P}{B/P}")},
	})
}

func swampsControlledBy(g *game.Game, source *game.Card) int {
	swamp := source.TextSubtype("Swamp")
	return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype(swamp) })
}
