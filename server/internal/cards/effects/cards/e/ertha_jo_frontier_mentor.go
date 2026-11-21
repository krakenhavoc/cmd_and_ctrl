package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0a66ce8b-af99-411f-8ecb-52a5d2f6af3d",
		Name:         "Ertha Jo, Frontier Mentor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ertha Jo, Frontier Mentor — create a 1/1 red Mercenary", effects.CreateAMercenary()),
			effects.On(game.EventActivateAbility, effects.B35YouActivatedAnAbilityTargetingACreatureOrPlayer,
				"Ertha Jo, Frontier Mentor — copy that ability", effects.CopyTheTriggeringAbility),
		},
	})
}
