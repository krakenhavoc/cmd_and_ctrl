package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liliana, Untouched by Death — Legendary Planeswalker — Liliana {2}{B}{B},
// loyalty 4 (Starter Commander Decks):
//
//	"+1: Mill three cards. If at least one Zombie card is milled this way,
//	 each opponent loses 2 life and you gain 2 life.
//	 −2: Target creature gets -X/-X until end of turn, where X is the number
//	 of Zombies you control.
//	 −3: You may cast Zombie spells from your graveyard this turn."
//
// The +1 is Seek Thrills' mill with a keyed tail told the cards that really
// reached the graveyard; one Zombie among them (read through the layers, so
// a changeling counts) is the drain, once. The −2 counts your Zombies as it
// resolves (CR 608.2h) and shrinks the target with a layer 7c effect until
// end of turn. The −3 is a standing permission to cast Zombie spells from
// your graveyard this turn only; it replaces nothing and ends with the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4f66489c-5a19-40ad-9126-461fa8231f1b",
		Name:            "Liliana, Untouched by Death",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Mill three cards. If at least one Zombie card is milled this way, each opponent loses 2 life and you gain 2 life.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(lilianaUntouchedMilled)}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Target creature gets -X/-X until end of turn, where X is the number of Zombies you control.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetHarm}),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					zombie := ctx.TextSubtype("Zombie")
					x := effects.CountControlled(g, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype(zombie) })
					return effects.BoostUntilEOT{Target: ts[0].ID, Power: -x, Toughness: -x, Label: "Liliana, Untouched by Death — -X/-X"}.Apply(ctx)
				},
			},
			{
				Label:  "−3: You may cast Zombie spells from your graveyard this turn.",
				Cost:   effects.LoyaltyCost(-3),
				Effect: effects.Do(effects.GapB183CastSubtypeSpellsFromGraveyardThisTurn{Subtype: "Zombie"}),
			},
		},
	})
}

// lilianaUntouchedMilled is "if at least one Zombie card is milled this way,
// each opponent loses 2 life and you gain 2 life", told the cards that were
// milled. An on-disk key.
var lilianaUntouchedMilled = effects.CardTail("liliana-untouched-by-death/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !c.HasSubtype(ctx.TextSubtype("Zombie")) {
			continue
		}
		if err := (effects.EachOpponentLosesLife{N: 2}).Apply(ctx); err != nil {
			return err
		}
		return effects.GainLife{Amount: 2}.Apply(ctx)
	}
	return nil
})
