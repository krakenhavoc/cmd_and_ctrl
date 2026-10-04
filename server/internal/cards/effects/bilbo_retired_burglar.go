package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bilbo, Retired Burglar — Legendary Creature — Halfling Rogue
// {1}{U}{R}, 1/3:
//
//	"When Bilbo enters or leaves the battlefield, the Ring tempts you.
//	 Whenever Bilbo deals combat damage to a player, create a Treasure
//	 token."
//
// One trigger for both events, as Cryogen Relic's. The leaves half
// looks back (CR 603.10a): its controller is whoever controlled Bilbo
// as it left, and Bilbo is gone by the time the tempt asks for a
// creature, so it can't be chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e7e3be5-04dd-4ff4-b121-3e65b3e00699",
		Name:         "Bilbo, Retired Burglar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrLeaves("Bilbo, Retired Burglar — the Ring tempts you", Do(TheRingTemptsYou{})),
			WheneverThisDealsCombatDamageToAPlayer("Bilbo, Retired Burglar — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
