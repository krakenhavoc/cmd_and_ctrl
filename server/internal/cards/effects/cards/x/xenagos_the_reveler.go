package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Xenagos, the Reveler — Legendary Planeswalker — Xenagos {2}{R}{G}, loyalty 3
// (Commander Legends: Battle for Baldur's Gate, S82 GAP #5948):
//
//	"+1: Add X mana in any combination of {R} and/or {G}, where X is the
//	     number of creatures you control.
//	 0: Create a 2/2 red and green Satyr creature token with haste.
//	 −6: Exile the top seven cards of your library. You may put any number of
//	     creature and/or land cards from among them onto the battlefield."
//
// The +1 is a loyalty ability, so it resolves on the stack and the mana is added
// as it resolves (not a mana ability, CR 605.1a), counted then: X independent
// picks of {R} or {G} (Selvala's "any combination"). The −6 exiles the seven
// (an exile, not a mill) and asks for any number of the creature and land
// cards among those that reached exile (Aminatou's Augury's pick); the rest
// stay exiled, as printed.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "942a318e-90eb-4f10-bb30-2e10f5cc922a",
		Name:            "Xenagos, the Reveler",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Add X mana in any combination of {R} and/or {G}, where X is the number of creatures you control.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
					if n <= 0 {
						return nil
					}
					return effects.AddMana{Amount: effects.ManaN(n, "R", "G")}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "0: Create a 2/2 red and green Satyr creature token with haste.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red and green Satyr with haste"), N: 1}),
			},
			{
				Label: "−6: Exile the top seven cards of your library. You may put any number of creature and/or land cards from among them onto the battlefield.",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{
						Player:   item.Controller,
						N:        7,
						To:       game.ZoneExile,
						ThenTail: ctx.ItemTail(xenagosTheRevelerExiled),
					}.Apply(ctx)
				},
			},
		},
	})
}

// xenagosTheRevelerExiled offers the creature and land cards among the cards
// that reached exile. An on-disk key.
var xenagosTheRevelerExiled = effects.CardTail("then/xenagos-the-reveler-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	var picks []uuid.UUID
	for _, id := range o.IDs {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile || c.FaceDown {
			continue
		}
		if c.IsCreature() || c.IsLand() {
			picks = append(picks, id)
		}
	}
	if len(picks) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Xenagos, the Reveler — you may put any number of creature and/or land cards from among them onto the battlefield",
		Cards:      picks,
		Min:        0,
		Max:        len(picks),
		ThenKey:    xenagosTheRevelerPut,
		LiveZone:   game.ZoneExile,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})

// xenagosTheRevelerPut puts the picked cards onto the battlefield together. An
// on-disk key.
var xenagosTheRevelerPut = effects.AskedPick("choose-cards/xenagos-the-reveler-put", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{Targets: r.Picked, Controller: ctx.Controller()}.Apply(ctx)
})
