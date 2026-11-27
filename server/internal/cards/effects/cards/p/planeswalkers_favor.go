package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planeswalker's Favor — Enchantment {2}{G} (Planeshift, S82 OLD #5615):
//
//	"{3}{G}: Target opponent reveals a card at random from their hand. Target
//	 creature gets +X/+X until end of turn, where X is the revealed card's mana
//	 value."
//
// Friendly Fire's random reveal (the game's keyed random stream, shown to the
// table), then the pump with X read from the revealed card. An empty hand, or
// an opponent who is no longer a legal target, reveals nothing and X is 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1990753c-06a7-4e1f-bc0e-905a9ba6e293",
		Name:         "Planeswalker's Favor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Target opponent reveals a card at random from their hand. Target creature gets +X/+X until end of turn, where X is the revealed card's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: effects.ForTargets(game.TargetPurpose{Slot: 1, UntilEndOfTurn: game.TargetBenefit}).Targets},
			Cost:    effects.ManaCost("{3}{G}"),
			Targets: effects.Clauses(effects.TargetPlayer("target opponent", effects.Opponent()), effects.TargetCreature("target creature")),
			Effect:  effects.OldB639RevealThenBoostTarget("Planeswalker's Favor", 1),
		}},
	})
}
