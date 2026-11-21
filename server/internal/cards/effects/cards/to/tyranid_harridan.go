package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyranid Harridan — Creature — Tyranid {4}{G}{U}, 4/4 (Warhammer 40,000
// Commander, S82 GAP batch 193, #5948):
//
//	"Flying, ward {4}
//	 Shrieking Gargoyles — Whenever this creature or another Tyranid you
//	 control deals combat damage to a player, create a 1/1 blue Tyranid
//	 Gargoyle creature token with flying."
//
// One trigger over every Tyranid you control (the Harridan included, whatever
// its types become is read as it dealt the damage): a Tyranid that connects
// makes one token per damage event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01432542-2ea7-45bc-bc90-151516faee0e",
		Name:            "Tyranid Harridan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{4}"), "Tyranid Harridan — ward {4}"),
			effects.On(game.EventDealDamage, effects.GapB193ThisOrAnotherTyranidDealtCombatDamageToAPlayer,
				"Tyranid Harridan — create a 1/1 blue Tyranid Gargoyle creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Tyranid Gargoyle with flying"), N: 1})),
		},
	})
}
