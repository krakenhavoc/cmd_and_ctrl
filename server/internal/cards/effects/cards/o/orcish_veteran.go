package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orcish Veteran — Creature — Orc {2}{R}, 2/2 (Fallen Empires, S82 OLD batch 852,
// #5637):
//
//	"This creature can't block white creatures with power 2 or greater.
//	 {R}: This creature gains first strike until end of turn."
//
// Cyclops Tyrant's block rule over the attacker, read as the block is declared
// (CR 509.1b): a white creature (a multicolour one too) whose power, layers
// included, is 2 or more. The first strike is Icatian Infantry's grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3b66180-ee08-4bef-90aa-7859dc77cb10",
		Name:         "Orcish Veteran",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.OfColor("W"), effects.PowerGE(2))),
				"it can't block white creatures with power 2 or greater"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Orcish Veteran — first strike until end of turn"),
		}},
	})
}
