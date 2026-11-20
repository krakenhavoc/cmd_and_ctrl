package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bola Warrior — Creature — Human Spellshaper Warrior {1}{R}, 1/1
// (Nemesis, #5618):
//
//	"{R}, {T}, Discard a card: Target creature can't block this turn."
//
// The spellshaper cost (nemSpellshaperCost) on Jamuraan Lion's body,
// targetCreatureCantBlockThisTurn: the restriction ends in the cleanup
// step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13f41253-364f-40a5-a7cd-84d8785b2460",
		Name:         "Bola Warrior",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}, Discard a card: Target creature can't block this turn.",
			Cost:    effects.NemSpellshaperCost("{R}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Bola Warrior — target creature can't block this turn"),
		}},
	})
}
