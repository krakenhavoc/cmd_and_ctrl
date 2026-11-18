package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hungry for More — Sorcery {B}{R} (Midnight Hunt, #5497):
//
//	"Create a 3/1 black and red Vampire creature token with trample,
//	 lifelink, and haste. Sacrifice it at the beginning of the next end
//	 step.
//	 Flashback {1}{B}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The token is created through the keyed creation so the sacrifice can be
// scheduled on exactly the token that came out (a delayed trigger naming
// it, CR 603.7c), as Echoing Assault does. Flashback is the shared
// constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3eaf1e6b-bf4c-4ea8-83ea-6827470fd802",
		Name:             "Hungry for More",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{B}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     ctx.Source(),
				Groups: []game.TokenGroup{{
					Template: ctx.TextToken(TokenCard("3/1 black and red Vampire with trample, lifelink, and haste")),
					Count:    1,
				}},
			}, ctx.ItemTail(hungryForMoreSacrifice))
		},
	})
}

var hungryForMoreSacrifice = CardTail("hungry-for-more/sacrifice-at-next-end-step", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Hungry for More — sacrifice the Vampire",
		Cards: append([]uuid.UUID(nil), o.IDs...),
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
