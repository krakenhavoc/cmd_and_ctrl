package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hologram Janeway — Legendary Artifact Creature — Illusion Officer {2}{W},
// 3/3 (Star Trek, #4973):
//
//	"Whenever you attack while creatures you control have total power 8
//	 or greater, put a +1/+1 counter on target creature."
//
// "Whenever you attack" is once per declaration however many creatures
// attack (OncePerBatch, The Chief Warg's shape). "While" is a condition
// of the trigger event, not an intervening "if" (CR 603.4 needs the word
// "if"): the total is read as attackers are declared, over every
// creature you control (Mosswort Bridge's creaturesYouControlTotalPowerAtLeast,
// current power after the layers), and not checked again on resolution.
//
// No simplification.
func init() {
	totalPowerEight := effects.CreaturesYouControlTotalPowerAtLeast(8)
	effects.Register(effects.Spec{
		OracleID:     "b572be16-ad13-4f9f-bcf6-085f81e5292c",
		Name:         "Hologram Janeway",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && totalPowerEight(g, source.Controller)
			}, "Hologram Janeway — put a +1/+1 counter on target creature", effects.PutPlusOneCounterOnEachLegalTarget)),
				effects.TargetCreature("target creature")),
		},
	})
}
