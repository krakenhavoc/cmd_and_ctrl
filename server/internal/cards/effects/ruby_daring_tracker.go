package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruby, Daring Tracker — Legendary Creature — Human Scout {R}{G}, 1/2
// (Wilds of Eldraine, #4471):
//
//	"Haste
//	 Whenever Ruby attacks while you control a creature with power 4 or
//	 greater, Ruby gets +2/+2 until end of turn.
//	 {T}: Add {R} or {G}."
//
// The attack trigger is Nighthowl Pursuer's (hobbitFerociousAttack):
// "while" is a condition on the trigger event, not an intervening "if",
// so the power check is made once, as the attack is declared, and the
// +2/+2 resolves whatever the board looks like by then. Ruby counts
// toward the check if she has power 4 or greater herself. The mana
// ability is a plain {T} for either colour; haste lets her use it the
// turn she arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d1b0eee-3a7a-4f22-a40d-7658a368962a",
		Name:            "Ruby, Daring Tracker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, hobbitFerociousAttack, "Ruby, Daring Tracker — +2/+2 until end of turn",
				thisGetsUntilEndOfTurn(2, 2, "Ruby, Daring Tracker — +2/+2")),
		},
	})
}
