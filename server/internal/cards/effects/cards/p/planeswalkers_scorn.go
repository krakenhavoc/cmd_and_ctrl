package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planeswalker's Scorn — Enchantment {2}{B} (Planeshift, S82 OLD #5615):
//
//	"{3}{B}: Target opponent reveals a card at random from their hand. Target
//	 creature gets -X/-X until end of turn, where X is the revealed card's mana
//	 value. Activate only as a sorcery."
//
// Planeswalker's Favor's reveal with a shrink, at sorcery speed. An empty hand,
// or an opponent who is no longer a legal target, reveals nothing and X is 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52cebfb0-0302-40d9-9a3f-f875dbf43659",
		Name:         "Planeswalker's Scorn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{B}: Target opponent reveals a card at random from their hand. Target creature gets -X/-X until end of turn, where X is the revealed card's mana value. Activate only as a sorcery.",
			Purpose:      game.Purpose{Targets: effects.ForTargets(game.TargetPurpose{Slot: 1, UntilEndOfTurn: game.TargetHarm}).Targets},
			Cost:         effects.ManaCost("{3}{B}"),
			Targets:      effects.Clauses(effects.TargetPlayer("target opponent", effects.Opponent()), effects.TargetCreature("target creature")),
			SorcerySpeed: true,
			Effect:       effects.OldB639RevealThenBoostTarget("Planeswalker's Scorn", -1),
		}},
	})
}
