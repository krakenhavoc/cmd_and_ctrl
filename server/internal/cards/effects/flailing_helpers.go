package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// flailing_helpers.go — the Flailing creatures' two abilities (ADR 0106
// §1, #1793). Flailing Manticore, Flailing Ogre and Flailing Soldier
// print the same pair word for word:
//
//	"{1}: This creature gets +1/+1 until end of turn. Any player may
//	 activate this ability.
//	 {1}: This creature gets -1/-1 until end of turn. Any player may
//	 activate this ability."
//
// Both rows are any-player rows (CR 602.2, 602.1b): whoever activates
// one pays its {1} (CR 602.1a), and the creature itself is what grows
// or shrinks, whoever controls it. Neither declares a Purpose, so the
// bot never pumps or shrinks a creature it does not control (ADR 0106
// owner decision 2) — it has no way to know which one helps.
//
// Append-only.

// flailingAbilities is the pair, with `name` in the effect labels.
func flailingAbilities(name string) []ActivatedAbility {
	return []ActivatedAbility{
		{
			Label:     "{1}: This creature gets +1/+1 until end of turn. Any player may activate this ability.",
			Cost:      ManaCost("{1}"),
			AnyPlayer: true,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Effect:    thisGetsUntilEndOfTurn(1, 1, name+" — +1/+1 until end of turn"),
		},
		{
			Label:     "{1}: This creature gets -1/-1 until end of turn. Any player may activate this ability.",
			Cost:      ManaCost("{1}"),
			AnyPlayer: true,
			// ADR 0142 sweep rulings: a shrink effect is remove.
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  thisGetsUntilEndOfTurn(-1, -1, name+" — -1/-1 until end of turn"),
		},
	}
}
