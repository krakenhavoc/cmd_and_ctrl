package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of Vitality — Enchantment {2}{G}{G} (Magic 2011, #5579):
//
//	"If this card is in your opening hand, you may begin the game with it
//	 on the battlefield.
//	 Creatures you control get +0/+1.
//	 Whenever a creature you control enters, you may gain 1 life."
//
// The opening-hand clause is Leyline of Hope's. The anthem is Angelic
// Shield's +0/+1 over creatures its controller controls, read on every
// layer pass. The trigger is Soul's Attendant's optional life gain over
// "a creature you control enters" (Impact Tremors' condition), so a
// creature entering under an opponent's control does not trigger it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbf2b19e-32a3-489f-9979-ba2748c52845",
		Name:         "Leyline of Vitality",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{YoursOnly: true}, 0, 1),
		},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, CreatureEnteredUnderYourControl, "Leyline of Vitality — gain 1 life", Do(GainLife{Amount: 1})),
				"Leyline of Vitality — gain 1 life?"),
		},
	})
}
