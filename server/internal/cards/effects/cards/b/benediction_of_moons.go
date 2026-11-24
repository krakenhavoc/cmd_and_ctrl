package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Benediction of Moons — Sorcery {W} (Guildpact):
//
//	"You gain 1 life for each player.
//	 Haunt (When this spell card is put into a graveyard after resolving,
//	 exile it haunting target creature.)
//	 When the creature this card haunts dies, you gain 1 life for each
//	 player."
//
// Haunt on a spell is CR 702.55a's second sentence (#6083): the card
// haunts only if it resolved, and from exile its owner controls the second
// trigger (CR 108.4). Every player still in the game counts, you among
// them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ff2fc3c-0871-4ba7-9b7c-8634ec770d6b",
		Name:         "Benediction of Moons",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.GainOneLifeForEachPlayer(ctx.Game, ctx.Item)
		})),
		Triggered: []game.TriggeredAbility{
			effects.HauntSpell("Benediction of Moons"),
			effects.WhenTheCreatureThisCardHauntsDies("Benediction of Moons — you gain 1 life for each player", effects.GainOneLifeForEachPlayer),
		},
	})
}
