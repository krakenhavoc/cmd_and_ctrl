package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saruman, the White Hand — Legendary Creature — Avatar Wizard
// {1}{U}{B}{R}, 2/5:
//
//	"Whenever you cast a noncreature spell, amass Orcs X, where X is
//	 that spell's mana value.
//	 Goblins and Orcs you control have ward {2}."
//
// The cast trigger reads the spell's mana value as it resolves
// (Endrek Sahr's triggeringSpellManaValue, CR 608.2h), its announced X
// included while the spell is on the stack, and feeds it to amass. An
// amass of 0 still makes the Army, as the rules say.
//
// The ward line is Hexing Squelcher's WardGranted over a tribe filter:
// the grant covers every Goblin and Orc you control, Saruman himself
// excepted only because he is neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2bade11e-04e0-42a2-8861-9257c99a7c08",
		Name:         "Saruman, the White Hand",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Saruman — amass Orcs X, where X is that spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					return Amass{Subtype: "Orc", N: triggeringSpellManaValue(g, item)}.Apply(NewContext(g, item))
				}),
			WardGranted(WardMana("{2}"), "Saruman — Goblins and Orcs you control have ward {2}",
				TribeFilter{Tribes: []string{"Goblin", "Orc"}, YoursOnly: true}.Matches),
		},
	})
}
