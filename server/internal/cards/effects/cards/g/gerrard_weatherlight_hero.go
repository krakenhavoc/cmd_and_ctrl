package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gerrard, Weatherlight Hero — Legendary Creature — Human Soldier
// {2}{R}{W}, 3/3 (Commander 2019, #5517):
//
//	"First strike
//	 When Gerrard dies, exile it and return to the battlefield all artifact
//	 and creature cards in your graveyard that were put there from the
//	 battlefield this turn."
//
// The dies trigger is Ancient Adamantoise's shape: Gerrard is exiled if it
// is still in the graveyard, and the return happens whether or not the
// exile did (a commander that went to the command zone, CR 903.9a). The
// return is Gerrard's Hourglass Pendant's: the set is read once as the
// trigger resolves, from your own graveyard, by each card's newest arrival
// there (a card that left and came back is a new object, CR 400.7), and the
// cards enter together, untapped, under their owner's control, which is
// yours. A token is not a card (CR 108.2), and Gerrard itself, exiled
// first, is not among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8a1e8435-16f2-475b-a8a9-7b9f0cf4594a",
		Name:            "Gerrard, Weatherlight Hero",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Gerrard, Weatherlight Hero — exile it and return the artifact and creature cards put into your graveyard from the battlefield this turn",
				c19GerrardDies),
		},
	})
}

func c19GerrardDies(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return c19GerrardReturn(ctx)
	}
	return effects.ExileTarget{
		Target:   item.SourceCardID,
		ThenTail: ctx.ItemTail(c19GerrardExiled),
	}.Apply(ctx)
}

// c19GerrardExiled is the clause after the exile. An on-disk key: never
// renamed, never reused.
var c19GerrardExiled = effects.CardTail("gerrard-weatherlight-hero/after-exile", func(ctx *effects.Context, _ game.TailOutcome) error {
	return c19GerrardReturn(ctx)
})

// c19GerrardReturn returns every artifact and creature card in the
// controller's graveyard that was put there from the battlefield this turn.
func c19GerrardReturn(ctx *effects.Context) error {
	g := ctx.Game
	ids := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
		if c.IsToken() || (!c.IsArtifact() && !c.IsCreature()) {
			return false
		}
		return effects.PutIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID)
	})
	if len(ids) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyardTogether{Targets: ids, Controller: uuid.Nil}.Apply(ctx)
}
