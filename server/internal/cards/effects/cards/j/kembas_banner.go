package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kemba's Banner — Artifact — Equipment {3}{W} (Phyrexia: All Will Be One
// Commander):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+1 for each creature you control.
//	 Equip {2}{W}"
//
// #5654, S82. The shared ForMirrodin trigger, one equip ability and
// PumpAttachedPer over the creatures the Banner's controller controls
// (CR 109.5: "you" is the Equipment's controller, not the host's),
// counted on every layer pass. The equipped creature counts itself while
// that player controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd31c248-8b3f-4fa5-a639-19b0e37b15c3",
		Name:         "Kemba's Banner",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ForMirrodin("Kemba's Banner")},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, kembasBannerCreaturesYouControl),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{W}")},
	})
}

// kembasBannerCreaturesYouControl counts "each creature you control" for
// the Equipment's controller, on post-layer-4 types.
func kembasBannerCreaturesYouControl(g *game.Game, source *game.Card) int {
	return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() })
}
