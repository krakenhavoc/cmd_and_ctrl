package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB632DiscipleAnyColorMana is the Disciples' "<mana>, {T}: Add one mana of any
// color" (Helionaut's {1}, Necra Disciple's {G}): a mana ability with a mana cost
// (Rootrider Faun's filter shape), summoning sickness applying (CR 302.6).
func OldB632DiscipleAnyColorMana(mana string) ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true, Mana: mana},
		Produced: "{W|U|B|R|G}",
		Label:    mana + ", {T}: Add one mana of any color",
	}
}

// OldB632PreventOneRow is "{<c>}, {T}: Prevent the next 1 damage that would be
// dealt to any target this turn" (Raka Disciple, Necra Disciple): Jhessian
// Balmgiver's one-damage shield, a mana cost added to the {T}.
func OldB632PreventOneRow(name, mana string) ActivatedAbility {
	return ActivatedAbility{
		Label:   mana + ", {T}: Prevent the next 1 damage that would be dealt to any target this turn.",
		Purpose: game.Purpose{Answers: game.AnswerPrevent},
		Cost:    Plus(ManaCost(mana), TapCost()),
		Targets: TargetAny(),
		Effect:  PreventNextOneDamageToTheTarget(name),
	}
}
