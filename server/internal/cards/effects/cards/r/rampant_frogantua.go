package cardsr

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampant Frogantua — Creature — Frog {2}{G}, 3/3:
//
//	"Trample
//	 This creature gets +10/+10 for each player who has lost the game.
//	 Whenever this creature deals combat damage to a player, you may mill
//	 that many cards. Put any number of land cards from among them onto the
//	 battlefield tapped."
//
// The static is a layer 7c modifier of ten for each seat that has been
// eliminated from the game, read live. The trigger asks "you may mill that
// many cards", the amount the combat damage dealt (CR 120.3, read from the
// event as the trigger resolves, so a damage amount a replacement changed
// counts as dealt), mills them from the controller's own library (Ainok
// Wayfarer's mill-then-pick), and offers any number of the land cards this
// mill put into the graveyard: those picked enter the battlefield tapped
// together. A card that left the graveyard in between is not offered.
func init() {
	const name = "Rampant Frogantua"
	effects.Register(effects.Spec{
		OracleID:        "a08393b5-603b-41c4-aa03-2ae9c74884a7",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +10/+10 for each player who has lost the game.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := 0
				for _, p := range g.Seats {
					if p != nil && p.Eliminated {
						n++
					}
				}
				c.Power += 10 * n
				c.Toughness += 10 * n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(name+" — you may mill that many cards; put any number of land cards from among them onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					return effects.MayChoice{
						Question: name + " — mill " + strconv.Itoa(item.Trigger.Event.Amount) + " cards?",
						Then:     rampantFrogantuaMill,
						Word:     strconv.Itoa(item.Trigger.Event.Amount),
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// rampantFrogantuaMill mills the damage dealt. An on-disk key.
var rampantFrogantuaMill = effects.MayYes("confirm/rampant-frogantua-mill", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	n, err := strconv.Atoi(r.Word)
	if err != nil || n <= 0 {
		return nil
	}
	return effects.MillToZone{Player: ctx.Controller(), N: n, ThenTail: ctx.ItemTail(rampantFrogantuaMilled)}.Apply(ctx)
})

// rampantFrogantuaMilled offers any number of the land cards the mill put
// into the graveyard. An on-disk key.
var rampantFrogantuaMilled = effects.CardTail("rampant-frogantua/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	var lands []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !c.IsLand() {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		lands = append(lands, id)
	}
	if len(lands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Rampant Frogantua — put any number of land cards from among them onto the battlefield tapped",
		Cards:    lands,
		Min:      0,
		Max:      len(lands),
		LiveZone: game.ZoneGraveyard,
		ThenKey:  rampantFrogantuaPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// rampantFrogantuaPick puts the lands picked onto the battlefield tapped,
// together. An on-disk key.
var rampantFrogantuaPick = effects.AskedPick("choose-cards/rampant-frogantua-lands", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, r.Picked, game.ZoneEntryOptions{
		Controller: ctx.Controller(),
		Tapped:     true,
	}, game.TailRef{}, 0)
})
