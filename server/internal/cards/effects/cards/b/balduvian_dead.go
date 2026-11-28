package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Balduvian Dead — Creature — Zombie {3}{B} 2/3 (Alliances, S82 OLD batch
// 820):
//
//	"{2}{R}, Exile a creature card from your graveyard: Create a 3/1 black and
//	 red Graveborn creature token with haste. Sacrifice it at the beginning
//	 of the next end step."
//
// The token's sacrifice is a delayed trigger (CR 603.7) naming the token.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18d6b89a-3588-43d9-a247-4747a7d784fc",
		Name:         "Balduvian Dead",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Exile a creature card from your graveyard: Create a 3/1 black and red Graveborn creature token with haste. Sacrifice it at the beginning of the next end step.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.ExileFromGraveyard(1, "a creature card", effects.IsCreatureCard)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: ctx.Controller(),
					Source:     ctx.Source(),
					Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("3/1 black and red Graveborn with haste")), Count: 1}},
				}, ctx.ItemTail(balduvianDeadSacrifice))
			},
		}},
	})
}

var balduvianDeadSacrifice = effects.CardTail("balduvian-dead/sacrifice-the-token", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Controller(),
		Label:      "Balduvian Dead — sacrifice the Graveborn",
		Cards:      append([]uuid.UUID(nil), o.IDs...),
		Body:       effects.SacrificeListedCardsBody,
	}.Apply(ctx)
})
