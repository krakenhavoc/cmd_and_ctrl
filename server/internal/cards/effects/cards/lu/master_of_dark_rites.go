package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master of Dark Rites — Creature — Vampire Cleric {B}, 1/1 (Commander,
// #5368):
//
//	"{T}, Sacrifice another creature: Add {B}{B}{B}. Spend this mana
//	 only to cast Vampire, Cleric, and/or Demon spells."
//
// A mana ability (CR 605.1a): no target, adds mana, not a loyalty
// ability, so it resolves at once (CR 605.3b). {T} and the sacrifice are
// both costs; the Master cannot sacrifice itself ("another"), and its
// own {T} needs it to have been under your control since the turn began.
// The restriction is one subtype tag admitting a spell with ANY of the
// three types (ManaRestrictAnyType's alternation, which works for
// subtypes), and only a cast (CR 106.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a65250f6-55e2-47ed-a355-d690047db6f2",
		Name:         "Master of Dark Rites",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				SacrificeOther: effects.SacrificeAnotherN(1, "another creature", effects.Creature()).SacrificeOther,
			},
			Produced:     "{B}{B}{B}",
			Restrictions: []string{game.ManaRestrictCast, "subtype:Vampire|Cleric|Demon"},
			Label:        "{T}, Sacrifice another creature: Add {B}{B}{B}. Spend this mana only to cast Vampire, Cleric, and/or Demon spells.",
			Answers:      game.AnswerSacOutlet,
		}},
	})
}
