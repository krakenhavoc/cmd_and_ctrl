package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	totalPowerEight := creaturesYouControlTotalPowerAtLeast(8)
	Register(Spec{
		OracleID:     "b572be16-ad13-4f9f-bcf6-085f81e5292c",
		Name:         "Hologram Janeway",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && totalPowerEight(g, source.Controller)
			}, "Hologram Janeway — put a +1/+1 counter on target creature", putPlusOneCounterOnEachLegalTarget)),
				TargetCreature("target creature")),
		},
	})
}
