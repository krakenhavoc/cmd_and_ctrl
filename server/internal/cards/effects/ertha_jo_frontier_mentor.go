package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ertha Jo, Frontier Mentor — Legendary Creature — Kor Advisor {2}{R}{W},
// 2/3 (Outlaws of Thunder Junction, #4468):
//
//	"When Ertha Jo enters, create a 1/1 red Mercenary creature token with
//	 '{T}: Target creature you control gets +1/+0 until end of turn.
//	 Activate only as a sorcery.'
//	 Whenever you activate an ability that targets a creature or player,
//	 copy that ability. You may choose new targets for the copy."
//
// The copy trigger is Illusionist's Bracers' (copyTheTriggeringAbility,
// CR 707.10), watching every ability its controller activates: the
// activation's own targets are read off the stack item, so a Mercenary's
// pump, an equip ability and a planeswalker's loyalty ability all count,
// and an ability that targets only a land or an artifact does not. A copy
// is not an activation, so it doesn't trigger Ertha Jo again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a66ce8b-af99-411f-8ecb-52a5d2f6af3d",
		Name:         "Ertha Jo, Frontier Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ertha Jo, Frontier Mentor — create a 1/1 red Mercenary", createAMercenary()),
			On(game.EventActivateAbility, b35YouActivatedAnAbilityTargetingACreatureOrPlayer,
				"Ertha Jo, Frontier Mentor — copy that ability", copyTheTriggeringAbility),
		},
	})
}
