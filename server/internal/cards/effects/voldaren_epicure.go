package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voldaren Epicure — Creature — Vampire {R}, 1/1 (Innistrad: Crimson Vow):
//
//	"When this creature enters, it deals 1 damage to each opponent. Create a
//	 Blood token."
func init() {
	Register(Spec{
		OracleID:     "cef82674-1f61-407b-801b-a76a3803e598",
		Name:         "Voldaren Epicure",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Voldaren Epicure — 1 damage to each opponent and create a Blood token",
				Do(DamageEachOpponent{N: 1}, s82Vow03CreateBlood(1))),
		},
	})
}
