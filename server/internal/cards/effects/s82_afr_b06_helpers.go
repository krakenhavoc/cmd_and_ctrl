package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// afrB06EnterAttach is "When this Equipment enters, attach it to target
// creature you control." (Paladin's Shield, Dueling Rapier).
func AfrB06EnterAttach(name string) game.TriggeredAbility {
	return Targeting(
		WhenThisEnters(name+" — attach it to target creature you control", AttachSourceToTarget),
		PermanentYouControl("target creature you control", Creature()),
	)
}
