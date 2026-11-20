package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harbor Bandit — Creature — Human Rogue {2}{B}, 2/2 (Magic 2013, #5568):
//
//	"This creature gets +1/+1 as long as you control an Island.
//	 {1}{U}: This creature can't be blocked this turn."
//
// Coastal Bulwark's Island pump and Drownyard Amalgam's evasion for
// mana. The Bandit can't be blocked only while it is still the
// permanent that activated: one that left and came back in response is a
// new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18c65823-01ae-4314-a254-d89a1a362633",
		Name:         "Harbor Bandit",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control an Island.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Island")),
			Apply:     pumpBy(1, 1),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Harbor Bandit — can't be blocked this turn"}.Apply(NewContext(g, item))
			},
		}},
	})
}
