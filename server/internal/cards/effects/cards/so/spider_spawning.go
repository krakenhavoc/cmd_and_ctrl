package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spider Spawning — Sorcery {4}{G} (Innistrad Remastered):
//
//	"Create a 1/2 green Spider creature token with reach for each creature
//	 card in your graveyard.
//	 Flashback {6}{B}"
//
// The creature cards are counted as the spell resolves; with none, no token is
// made. Flashback exiles the spell as it is cast, so it is never counted.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "da113cc9-bb2b-4af4-9c43-32c165b87363",
		Name:             "Spider Spawning",
		Purpose:          game.Purpose{Answers: game.AnswerMakesBlocker},
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{6}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.GapGraveyardCardsMatching(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
			if n == 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/2 green Spider with reach"), N: n}.Apply(ctx)
		},
	})
}
