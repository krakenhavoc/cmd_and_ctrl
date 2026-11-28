package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harvest Wurm — Creature — Wurm {1}{G}, 3/2 (Weatherlight):
//
//	"When this creature enters, sacrifice it unless you return a basic land card
//	 from your graveyard to your hand."
//
// The choice is made as the trigger resolves. With no basic land card in the
// graveyard the Wurm is sacrificed at once; otherwise the controller names one,
// or none, which declines and sacrifices it. The card named must still be in the
// graveyard when the answer comes in, and the Wurm stays only if the card really
// reached a hand. A Wurm that has left the battlefield, or come back as a new
// object, is not sacrificed (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd539212-ef7b-4f78-8d62-c42289aedc51",
		Name:         "Harvest Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Harvest Wurm — sacrifice it unless you return a basic land card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					lands := effects.GraveyardIDs(g, ctx.Controller(), harvestWurmBasicLand)
					if len(lands) == 0 {
						return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:    ctx.Controller(),
						FromPlayer: ctx.Controller(),
						Source:     item.SourceCardID,
						Question:   "Harvest Wurm — return a basic land card from your graveyard to your hand, or sacrifice the Wurm",
						Cards:      lands,
						Min:        0,
						Max:        1,
						LiveZone:   game.ZoneGraveyard,
						ThenKey:    harvestWurmPick,
						Asked:      game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

func harvestWurmBasicLand(c game.Card) bool { return c.IsLand() && c.HasSupertype("basic") }

// harvestWurmPick is the answer: the card goes to the hand, or the Wurm is
// sacrificed when none was named or the card is gone. An on-disk key.
var harvestWurmPick = effects.AskedPickRule("choose-cards/s82-old-b775-harvest-wurm-return",
	func(picked []game.Card) bool { return len(picked) == 1 },
	func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Picked) == 1 {
			if c, ok := ctx.Game.LookupCardForEffect(r.Picked[0]); ok && harvestWurmBasicLand(c) {
				if z := ctx.Game.FindCardZoneForEffect(c.InstanceID); z != nil && z.Kind == game.ZoneGraveyard {
					return harvestWurmReturn(ctx, c.InstanceID)
				}
			}
		}
		return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
	})

// harvestWurmReturn puts the card in its owner's hand and keeps the Wurm only
// if that happened.
func harvestWurmReturn(ctx *effects.Context, id uuid.UUID) error {
	if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
		return err
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
		return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
	}
	return nil
}
