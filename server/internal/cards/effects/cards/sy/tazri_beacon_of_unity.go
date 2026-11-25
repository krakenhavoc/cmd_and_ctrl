package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tazri, Beacon of Unity — Legendary Creature — Human Warrior {4}{W}, 4/6
// (Zendikar Rising, #5508):
//
//	"This spell costs {1} less to cast for each creature in your party.
//	 {2/U}{2/B}{2/R}{2/G}: Look at the top six cards of your library. You may
//	 reveal up to two Cleric, Rogue, Warrior, Wizard, and/or Ally cards from
//	 among them and put them into your hand. Put the rest on the bottom of
//	 your library in a random order."
//
// The party discount is counted as the spell is cast. The activated ability's
// cost is four hybrid symbols, each payable as {2} or its colour. Its effect is
// Sigarda's look-and-take with a limit of two and the five-subtype filter: the
// looking is private, the cards taken are revealed, and the rest go to the
// bottom in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "ce703e1d-b7b2-4b8b-a12c-714fc152d0a5",
		Name:              "Tazri, Beacon of Unity",
		Completeness:      effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{effects.S82OldB089PartyCostReduction()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2/U}{2/B}{2/R}{2/G}: Look at the top six cards of your library. You may reveal up to two Cleric, Rogue, Warrior, Wizard, and/or Ally cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2/U}{2/B}{2/R}{2/G}"),
			Effect: effects.LookAtTopThenMayTakeToHand(6,
				effects.AnySubtype("Cleric", "Rogue", "Warrior", "Wizard", "Ally"), 2,
				"Tazri, Beacon of Unity — you may reveal up to two Cleric, Rogue, Warrior, Wizard, and/or Ally cards from among them and put them into your hand"),
		}},
	})
}
