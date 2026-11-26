package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exuberant Firestoker — {2}{R} Creature — Human Druid Shaman 1/1 (Shards of
// Alara, S82 OLD b515):
//
//	"At the beginning of your end step, if you control a creature with power 5
//	 or greater, you may have this creature deal 2 damage to target player or
//	 planeswalker.
//	 {T}: Add {C}."
//
// The intervening "if" is checked as the end step begins and again as the
// trigger resolves (CR 603.4); power is read through the layers.
//
// No simplification.
func init() {
	const label = "Exuberant Firestoker — it deals 2 damage to target player or planeswalker"
	effects.Register(effects.Spec{
		OracleID:      "f8c0c555-c750-4842-8924-d634206bbb81",
		Name:          "Exuberant Firestoker",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.TapForColorless()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.InterveningIf(effects.CondPowerAtLeast{N: 5},
				effects.On(game.EventBeginEndStep, effects.ByYou, label, effects.SourceDealsDamageToEachLegalTarget(2))),
				"Exuberant Firestoker — have it deal 2 damage to the target?"),
				effects.TargetPlayerOrPlaneswalker()),
		},
	})
}
