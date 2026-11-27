package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infiltrator's Magemark — Enchantment — Aura {2}{U} (Guildpact):
//
//	"Enchant creature
//	 Creatures you control that are enchanted get +1/+1 and can't be blocked
//	 except by creatures with defender."
//
// Beastmaster's Magemark's anthem, with a block rule over the same set of
// creatures: a blocker without defender is refused (CR 509.1b), read on
// every check.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df82b67f-f685-4278-a108-f3689a3a121b",
		Name:         "Infiltrator's Magemark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.OldB570MagemarkPump()},
		BlockRules: []game.BlockRule{effects.CantBeBlockedExceptBy(
			effects.OnMatching(effects.OldB570YourEnchantedCreaturePredicate),
			effects.HasKeyword("defender"),
			"creatures with defender")},
	})
}
