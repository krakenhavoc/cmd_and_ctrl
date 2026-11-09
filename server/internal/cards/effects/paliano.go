package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paliano — Plane — Fiora (Commander Masters, #4870):
//
//	"When one or more creatures you control deal combat damage to a
//	 player, if there is no monarch, you become the monarch.
//	 Whenever chaos ensues, create a 1/1 black Assassin creature token
//	 with deathtouch and haste."
//
// A plane (ADR 0216 §6). The first ability has an intervening "if" (CR
// 603.4): it triggers only when there is no monarch and does nothing if
// someone has become it by the time it resolves; it triggers once for each
// player the planar controller's creatures hit in a combat damage step
// (CR 603.2c). The chaos ability (CR 311.7) makes one Assassin.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5f0da12-b76f-42fb-a52e-cd09535e72fb",
		Name:         "Paliano",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			onlyIfThereIsNoMonarch(InterveningIf(CondNoMonarch{}, WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Paliano — if there is no monarch, you become the monarch", Do(BecomeTheMonarch{})))),
			WheneverChaosEnsues("Paliano — create a 1/1 black Assassin creature token with deathtouch and haste",
				Do(CreateToken{Template: TokenCard("1/1 black Assassin with deathtouch and haste"), N: 1})),
		},
	})
}
