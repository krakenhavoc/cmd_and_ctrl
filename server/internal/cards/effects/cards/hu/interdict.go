package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Interdict — Instant {1}{U} (Tempest, S82 OLD batch 766, #5628):
//
//	"Counter target activated ability from an artifact, creature, enchantment,
//	 or land. That permanent's activated abilities can't be activated this turn.
//	 (Mana abilities can't be targeted.)
//	 Draw a card."
//
// Stifle's counter on an activated ability whose source is one of the four
// types, then a layer-free restriction on that permanent (every activated
// ability, mana abilities included, until end of turn, pinned to the object it
// was, CR 400.7), then the draw. If the ability is gone as Interdict resolves
// the target is illegal and nothing happens, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93771480-2672-4bc7-81a8-011d3b8878b0",
		Name:         "Interdict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets: effects.TargetSpellOrAbility("target activated ability from an artifact, creature, enchantment, or land",
			effects.AnActivatedAbilityItem(), effects.FromAnArtifactCreatureEnchantmentOrLandSource()),
		OnResolve: effects.InterdictResolves,
	})
}
