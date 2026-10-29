package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smaug the Magnificent — Legendary Creature — Dragon {2}{R}{R}, 4/3
// (The Hobbit):
//
//	"Flying, haste
//	 Whenever Smaug attacks, he deals damage equal to the number of
//	 Treasures you control to any target.
//	 At the beginning of your upkeep, create a Treasure token."
//
// Flying and haste are printed keywords. The attack trigger chooses any
// target as it goes on the stack and counts the Treasures as it resolves
// (CR 608.2b), so a Treasure sacrificed in response shrinks the hit. Smaug
// is the damage source; none, or an illegal target, does nothing. The upkeep
// trigger is Awakening Zone's shape with a Treasure.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c53b7602-5e36-4900-896c-844bf6eac22e",
		Name:            "Smaug the Magnificent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Smaug the Magnificent — deals damage equal to the number of Treasures you control to any target",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := len(permanentsControlledByMatching(g, item.Controller, Subtype("Treasure")))
					if n == 0 {
						return nil
					}
					return damageToFirstTarget(n)(item, ctx)
				}), TargetAny()),
			AtYourUpkeep("Smaug the Magnificent — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
