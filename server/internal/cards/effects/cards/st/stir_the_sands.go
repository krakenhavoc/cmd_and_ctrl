package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stir the Sands — {4}{B}{B} Sorcery (Amonkhet, #5535):
//
//	"Create three 2/2 black Zombie creature tokens.
//	 Cycling {3}{B} ({3}{B}, Discard this card: Draw a card.)
//	 When you cycle this card, create a 2/2 black Zombie creature token."
//
// The cycle trigger fires from the graveyard (CR 702.29c) and goes on the
// stack above the cycling ability, so the Zombie arrives before the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52fa7e9b-36d9-409c-8981-6f8dfd8bb031",
		Name:         "Stir the Sands",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 3}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
				"Stir the Sands — cycled: create a 2/2 black Zombie",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}))),
		},
	})
}
