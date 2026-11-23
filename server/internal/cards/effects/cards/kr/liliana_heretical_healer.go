package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana, Heretical Healer // Liliana, Defiant Necromancer — Magic Origins
// (Commander Collection: Black, #5948), a transforming legendary creature
// whose back face is a planeswalker.
//
// Front face, Legendary Creature — Human Cleric {1}{B}{B}, 2/3:
//
//	"Lifelink
//	 Whenever another nontoken creature you control dies, exile Liliana,
//	 Heretical Healer, then return her to the battlefield transformed under
//	 her owner's control. If you do, create a 2/2 black Zombie creature
//	 token."
//
// Back face, Legendary Planeswalker — Liliana, loyalty 3:
//
//	"+2: Each player discards a card.
//	 −X: Return target nonlegendary creature card with mana value X from your
//	     graveyard to the battlefield.
//	 −8: You get an emblem with "Whenever a creature dies, return it to the
//	     battlefield under your control at the beginning of the next end
//	     step.""
//
// The back face registers under "<oracle_id>#1" (game.CatalogKey). The flip is
// a plain trigger (every qualifying death triggers it; a Liliana that has
// already gone does nothing) reading the dead creature's last-known
// information (CR 603.10a). "If you do" is the exile-and-return's own tail
// (ExileAndReturnTransformedThenForEffect): the Zombie is made only once she
// has really come back, so a replacement that kept her in exile makes none.
// The −X is a loyalty cost of X with the target's mana value held to X, at
// announce and on resolution. The emblem's trigger schedules a CR 603.7
// delayed trigger carrying the dead creature, which returns it under the
// emblem owner's control at the next end step only if it is still in a
// graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        lilianaHereticalHealerOracle,
		Name:            "Liliana, Heretical Healer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B19AnotherNontokenCreatureYouControlDied(ev, source, g)
			},
			Key: "Liliana, Heretical Healer — exile her, then return her transformed; if you do, create a 2/2 black Zombie",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.ExileAndReturnTransformedThenForEffect(item.SourceCardID, uuid.Nil, ctx.ItemTail(lilianaHereticalHealerZombie))
			},
		}},
	})

	effects.Register(effects.Spec{
		OracleID:        lilianaHereticalHealerOracle + "#1",
		Name:            "Liliana, Defiant Necromancer",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Liliana, Defiant Necromancer emblem",
			Text:  "Whenever a creature dies, return it to the battlefield under your control at the beginning of the next end step.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := effects.DiedCreature(ev, g)
					return ok
				}, "Liliana, Defiant Necromancer emblem — return it to the battlefield under your control at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ScheduleDelayedTrigger{
							Label: "Liliana, Defiant Necromancer emblem — return it to the battlefield under your control",
							Cards: []uuid.UUID{item.Trigger.Event.CardID},
							Body:  effects.GapB45DesertWarfareReturnBody,
						}.Apply(effects.NewContext(g, item))
					}),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+2: Each player discards a card.",
				Cost:   effects.LoyaltyCost(2),
				Effect: effects.TspB1EachPlayerDiscardsACard("Liliana, Defiant Necromancer — discard a card"),
			},
			{
				Label: "−X: Return target nonlegendary creature card with mana value X from your graveyard to the battlefield.",
				Cost:  effects.LoyaltyMinusX(),
				Targets: effects.TargetCardInGraveyard("target nonlegendary creature card with mana value X from your graveyard",
					effects.Creature(), effects.YouOwn(), effects.Not(effects.Legendary())).WithManaValueEqualsX(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ids := effects.LegalTargetCardIDs(ctx)
					if len(ids) == 0 {
						return nil
					}
					return effects.ReturnFromGraveyardTogether{Targets: ids}.Apply(ctx)
				},
			},
			{
				Label:  "−8: You get an emblem with \"Whenever a creature dies, return it to the battlefield under your control at the beginning of the next end step.\"",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// lilianaHereticalHealerOracle is shared by both faces.
const lilianaHereticalHealerOracle = "b96a8ad2-3d86-459b-a34f-19c9dc07c3c4"

// lilianaHereticalHealerZombie is "if you do, create a 2/2 black Zombie
// creature token": it runs once she has returned, told her new object. An
// on-disk key: never renamed, never reused.
var lilianaHereticalHealerZombie = effects.CardTail("liliana-heretical-healer/zombie", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
})
