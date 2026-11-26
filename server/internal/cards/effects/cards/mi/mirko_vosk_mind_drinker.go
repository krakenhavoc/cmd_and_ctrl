package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirko Vosk, Mind Drinker — Legendary Creature — Vampire {3}{U}{B}, 2/4
// (Dragon's Maze, S82 OLD b405, #5564):
//
//	"Flying
//	 Whenever Mirko Vosk deals combat damage to a player, that player reveals
//	 cards from the top of their library until they reveal four land cards,
//	 then puts those cards into their graveyard."
//
// The damaged player is the event's target (Shriekgeist's shape). The run is
// revealed as one announcement and then milled, so the fourth land goes too;
// a library short of four lands is revealed and milled whole.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "469da932-8186-4e54-a858-cfc72c4ded96",
		Name:            "Mirko Vosk, Mind Drinker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Mirko Vosk — that player reveals until four lands and mills them",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					ev := ctx.Item.Trigger.Event
					if ev.Amount <= 0 || ctx.PlayerByID(ev.Target) == nil {
						return nil
					}
					return effects.S82OldB405MillUntilFourLands(ctx, ev.Target)
				}))),
		},
	})
}
