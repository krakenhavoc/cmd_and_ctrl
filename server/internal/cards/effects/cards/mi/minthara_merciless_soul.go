package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minthara, Merciless Soul — {2}{W}{B} Legendary Creature — Elf Cleric 2/2
//
//	"Minthara has ward {X}, where X is the number of experience counters you
//	 have.
//	 At the beginning of your end step, if a permanent you controlled left the
//	 battlefield this turn, you get an experience counter.
//	 Creatures you control get +1/+0 for each experience counter you have."
//
// The ward's X is read as the ward trigger resolves (a ward with no counters
// asks nothing). The end-step ability is revolt's intervening "if"
// (CR 603.4). The anthem is a layer 7c modification that re-reads the
// controller's experience counters whenever one is gained.
func init() {
	anthem := game.StaticAbility{
		Label:    "Creatures you control get +1/+0 for each experience counter you have.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			c.Power += effects.ExperienceCounters(g, source.Controller)
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "fad0c174-4c4a-4662-bd1f-224886e826e1",
		Name:         "Minthara, Merciless Soul",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{anthem},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.GapB215WardXExperience(), "Minthara, Merciless Soul — ward {X}"),
			effects.AtYourEndStepIfRevolt("Minthara, Merciless Soul — you get an experience counter", effects.YouGetAnExperienceCounter),
		},
	})
}
