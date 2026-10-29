package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One search, with an injective assignment to the Zombie and Swamp
// clauses: a dual-qualified card cannot fill both printed slots.
var corpseHarvesterSearch = game.RegisterSearchRule("search/corpse-harvester-zombie-and-swamp", func(picked []game.Card, _ int) bool {
	if len(picked) > 2 {
		return false
	}
	if len(picked) == 0 {
		return true
	}
	if len(picked) == 1 {
		return picked[0].HasSubtypeWord("Zombie") || picked[0].HasSubtypeWord("Swamp")
	}
	return picked[0].HasSubtypeWord("Zombie") && picked[1].HasSubtypeWord("Swamp") || picked[1].HasSubtypeWord("Zombie") && picked[0].HasSubtypeWord("Swamp")
})

func init() {
	Register(Spec{OracleID: "f02ad6fa-7a82-471a-979e-0bb9ba866ef6", Name: "Corpse Harvester", Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, {T}, Sacrifice a creature: Search your library for a Zombie card and a Swamp card, reveal them, put them into your hand, then shuffle.",
			Cost:    Plus(ManaCost("{1}{B}"), TapCost(), SacrificeN(1, "a creature", Creature())),
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tutors: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return (SearchLibrary{Player: item.Controller, Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Zombie") || c.HasSubtypeWord("Swamp") },
					Dest: game.ZoneHand, Limit: 2, Reveal: true, Shuffle: true, Rule: corpseHarvesterSearch,
					Reason: "Corpse Harvester — a Zombie card and a Swamp card"}).Apply(NewContext(g, item))
			},
		}},
	})
}
