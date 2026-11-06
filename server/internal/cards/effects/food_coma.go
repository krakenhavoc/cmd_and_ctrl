package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Food Coma — Enchantment {3}{W} (Wilds of Eldraine, #4471):
//
//	"When this enchantment enters, exile target creature an opponent
//	 controls until this enchantment leaves the battlefield. Create a Food
//	 token. (It's an artifact with "{2}, {T}, Sacrifice this token: You
//	 gain 3 life.")"
//
// The Oblivion Ring family's entry trigger over a creature an opponent
// controls (exileChosenTargetUntilThisLeaves: the card comes back under
// its owner's control when Food Coma leaves, CR 610.3c), then the
// catalog's Food token. If Food Coma has already left by the time the
// trigger resolves, nothing is exiled (CR 610.3b) but the Food is still
// made; if the creature is no longer a legal target, the trigger does
// nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	exile := exileChosenTargetUntilThisLeaves("Food Coma — the exiled card returns when Food Coma leaves the battlefield")
	Register(Spec{
		OracleID:     "aab3774c-4ec3-4946-a09b-af0fab7cacb7",
		Name:         "Food Coma",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Food Coma — exile target creature an opponent controls until this enchantment leaves the battlefield, then create a Food token",
				func(g *game.Game, item *game.StackItem) error {
					if err := exile(g, item); err != nil {
						return err
					}
					return CreateToken{Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
