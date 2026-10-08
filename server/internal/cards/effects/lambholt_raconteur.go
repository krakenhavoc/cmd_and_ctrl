package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lambholt Raconteur // Lambholt Ravager — {3}{R} Creature — Human
// Werewolf 2/4 // Creature — Werewolf 4/4 (#2586, ADR 0132):
//
//	Front: "Whenever you cast a noncreature spell, this creature deals 1
//	        damage to each opponent.
//	        Daybound"
//	Back:  "Whenever you cast a noncreature spell, this creature deals 2
//	        damage to each opponent.
//	        Nightbound"
//
// Damage, not life loss: each opponent is dealt it by the creature as
// the source, so prevention and lifelink see it. The trigger goes on the
// stack above the spell that caused it.
//
// No simplification.
func init() {
	const oracle = "f3c104e2-470b-429f-a047-a21edc2adb3b"
	ping := func(name string, n int) game.TriggeredAbility {
		return WheneverYouCast(Noncreature(), name+" — deal damage to each opponent",
			func(g *game.Game, item *game.StackItem) error { return damageToEachOpponent(g, item, n) })
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Lambholt Raconteur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered:       []game.TriggeredAbility{ping("Lambholt Raconteur", 1)},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Lambholt Ravager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered:       []game.TriggeredAbility{ping("Lambholt Ravager", 2)},
	})
}
