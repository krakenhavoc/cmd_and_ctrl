package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cry of Contrition — Sorcery {B} (Guildpact):
//
//	"Target player discards a card.
//	 Haunt (When this spell card is put into a graveyard after resolving,
//	 exile it haunting target creature.)
//	 When the creature this card haunts dies, target player discards a
//	 card."
//
// Haunt on a spell is CR 702.55a's second sentence (#6083): the card
// haunts only if it resolved, and from exile its owner controls the second
// trigger (CR 108.4). The discarding player picks the card (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b455d6a-dfee-4828-8873-094caf85c584",
		Name:         "Cry of Contrition",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.TargetPlayerDiscards(1)(ctx.Game, ctx.Item)
		})),
		Triggered: []game.TriggeredAbility{
			effects.HauntSpell("Cry of Contrition"),
			effects.Targeting(
				effects.WhenTheCreatureThisCardHauntsDies("Cry of Contrition — target player discards a card", effects.TargetPlayerDiscards(1)),
				effects.TargetPlayer("target player"),
			),
		},
	})
}
