package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kona, Rescue Beastie — Legendary Creature — Beast Survivor {3}{G}, 4/3
// (Duskmourn: House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if Kona is
//	 tapped, you may put a permanent card from your hand onto the
//	 battlefield."
//
// The Survival trigger (survival.go) with the put-from-hand primitive
// (PutFromHandOntoBattlefield): any permanent card, a land included, put
// without being cast, so no mana is paid and no land drop is used.
// "You may" lets the player decline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76372327-41ec-4516-abff-3b612e073629",
		Name:         "Kona, Rescue Beastie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Survival("Kona, Rescue Beastie — you may put a permanent card from your hand onto the battlefield",
				Do(PutFromHandOntoBattlefield{
					Optional: true,
					Label:    "Kona, Rescue Beastie — put a permanent card from your hand onto the battlefield",
				})),
		},
	})
}
