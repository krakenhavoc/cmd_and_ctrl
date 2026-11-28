package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistmoon Griffin — Creature — Griffin {3}{W}, 2/2 (Weatherlight, S82 older sets batch 777, #5629):
//
//	"Flying
//	 When this creature dies, exile it, then return the top creature card of your graveyard to
//	 the battlefield."
//
// Flying rides PrintedKeywords. Rooting Kavu's exile of the card that died (only while it is still
// that card in the graveyard, CR 400.7), then, whether or not it reached exile, the topmost creature
// card of the controller's graveyard (the last card in the zone, CR 404.2) is read as that clause
// resolves, and returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7fe9d30f-c778-4769-a3ae-3cb48cd23e07",
		Name:            "Mistmoon Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Mistmoon Griffin — exile it, then return the top creature card of your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					dead := item.Trigger.Event.CardID
					if !effects.DiedCardStillInGraveyard(ctx, dead) {
						return mistmoonGriffinReturn(ctx)
					}
					return effects.ExileTarget{Target: dead, ThenTail: ctx.ItemTail(mistmoonGriffinExiled)}.Apply(ctx)
				}),
		},
	})
}

// mistmoonGriffinExiled is "then return the top creature card of your graveyard": it runs
// once the exile has been attempted, whatever its outcome. An on-disk key.
var mistmoonGriffinExiled = effects.CardTail("mistmoon-griffin/exiled", func(ctx *effects.Context, _ game.TailOutcome) error {
	return mistmoonGriffinReturn(ctx)
})

func mistmoonGriffinReturn(ctx *effects.Context) error {
	id, ok := effects.S82OldB777TopCreatureCardOfYourGraveyard(ctx.Game, ctx.Controller())
	if !ok {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
}
