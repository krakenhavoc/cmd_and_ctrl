package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Bounty — Enchantment — Aura Curse {1}{G} (Commander 2017,
// #5533):
//
//	"Enchant player
//	 Whenever enchanted player is attacked, untap all nonland permanents
//	 you control. Each opponent attacking that player untaps all nonland
//	 permanents they control."
//
// Curse of Verbosity's shape (c17CurseWhenAttacked) with Unstoppable Plan's
// untap: one declaration is one "is attacked", and the attacking opponent's
// ability untaps the nonland permanents THEY control, since they control
// it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6689782-08d8-48e1-a05d-cd040dfe85bc",
		Name:         "Curse of Bounty",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: effects.C17CurseWhenAttacked("Curse of Bounty — untap all nonland permanents you control",
			"Curse of Bounty — the attacking opponent untaps all nonland permanents they control",
			effects.C17UntapAllNonlandYouControl),
	})
}
