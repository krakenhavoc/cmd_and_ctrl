package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zirda, the Dawnwaker — Legendary Creature — Elemental Fox
// {1}{R/W}{R/W}, 3/3 (Ikoria; Multiverse Legends, #5908):
//
//	"Companion — Each permanent card in your starting deck has an
//	 activated ability. (If this card is your chosen companion, you may
//	 put it into your hand from outside the game for {3} as a sorcery.)
//	 Abilities you activate that aren't mana abilities cost {2} less to
//	 activate. This effect can't reduce the mana in that cost to less
//	 than one mana.
//	 {1}, {T}: Target creature can't block this turn."
//
// The reduction is Training Grounds' (KeepsOneMana over
// ActivationCostsLess) with Zirda's reach: every ability its controller
// activates (YouActivate), from any source and any zone (cycling from
// the hand, unearth from the graveyard, an opponent's permanent whose
// ability they may activate), except a mana ability (CR 605.1a). Only
// generic mana comes off, and a cost of one mana or less is left alone,
// so Zirda's own {1} stays {1}. A loyalty ability has no mana to take
// off, and the companion's {3} is a special action, not an ability.
//
// The can't-block ability is Bola Warrior's row with a {1}, {T} cost.
//
// Companion (ADR 0172) is not this file's business: the condition is
// internal/deck's registry entry keyed by this oracle ID
// (deck/companion_zirda.go, #5909), and the {3} special action is the
// engine's for any revealed companion.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b515c56-7b1e-42a8-82b0-1b30201fb162",
		Name:         "Zirda, the Dawnwaker",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			KeepsOneMana(ActivationCostsLess(2,
				"Abilities you activate that aren't mana abilities cost {2} less to activate. This effect can't reduce the mana in that cost to less than one mana.",
				YouActivate(), cmdr2B03NotAManaAbility())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target creature can't block this turn.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(CantBlockUntilEOT(0, 0, 0)),
			Effect:  targetCreatureCantBlockThisTurn("Zirda, the Dawnwaker — target creature can't block this turn"),
		}},
	})
}
