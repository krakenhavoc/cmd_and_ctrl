package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB632TapTwoUntappedCreaturesCost is "Tap two untapped creatures you control"
// as an ability cost (Diversionary Tactics, S82 OLD batch 632, #5614). It is not
// {T}, so summoning sickness does not stop a creature paying it (CR 302.6).
func OldB632TapTwoUntappedCreaturesCost() game.AbilityCost {
	const label = "two untapped creatures you control"
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:  2,
		Filter: TargetCreature(label),
		Label:  label,
	}}
}

// OldB632DiscipleFlyingRow is "<mana>, {T}: Target creature gains flying until
// end of turn" (Raka Disciple). {T} is in the cost, so summoning sickness
// applies (CR 302.6); a target gone by resolution gets nothing (CR 608.2b).
func OldB632DiscipleFlyingRow(name, mana string) ActivatedAbility {
	return ActivatedAbility{
		Label:   mana + ", {T}: Target creature gains flying until end of turn.",
		Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
		Cost:    Plus(ManaCost(mana), TapCost()),
		Targets: TargetCreature("target creature"),
		Effect:  EbTargetCreatureGainsUntilEOT("flying", name+" — flying until end of turn"),
	}
}

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
