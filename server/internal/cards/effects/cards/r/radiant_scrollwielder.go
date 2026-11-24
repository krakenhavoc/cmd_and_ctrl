package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Radiant Scrollwielder — Creature — Dwarf Cleric {2}{R}{W}, 2/4 (Strixhaven:
// School of Mages, #5502):
//
//	"Instant and sorcery spells you control have lifelink.
//	 At the beginning of your upkeep, exile an instant or sorcery card at random
//	 from your graveyard. You may cast it this turn. If a spell cast this way
//	 would be put into your graveyard, exile it instead."
//
// The static is Heartflame Duelist's spell-keyword grant: the damage an instant
// or sorcery spell of yours deals gains you life while the Scrollwielder is on
// the battlefield. The upkeep trigger picks uniformly at random among your
// instant and sorcery cards in the graveyard (the engine's keyed random
// stream), exiles that card, and then gives you a permission to cast it from
// exile this turn, paying its mana cost, at its own timing; the spell is exiled
// instead of going to a graveyard. With none in the graveyard nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f5ef838-839a-4eb0-8d65-c8c8def0a233",
		Name:         "Radiant Scrollwielder",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.SpellsYouControlHave(effects.Or(effects.Instant(), effects.Sorcery()), "lifelink"),
				"Instant and sorcery spells you control have lifelink."),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Radiant Scrollwielder — exile an instant or sorcery card at random from your graveyard; you may cast it this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ids := effects.GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
					if len(ids) == 0 {
						return nil
					}
					picked := effects.RandomPick(ctx, ids, 1)
					if len(picked) == 0 {
						return nil
					}
					return effects.ExileTarget{Target: picked[0], ThenTail: ctx.ItemTail(radiantScrollwielderExiled)}.Apply(ctx)
				}),
		},
	})
}

// radiantScrollwielderExiled is "you may cast it this turn", once the card has
// reached exile: a permission over that card, at its mana cost, that exiles the
// spell instead of putting it into a graveyard. An on-disk key.
var radiantScrollwielderExiled = effects.CardTail("old-b044/radiant-scrollwielder-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:                  ctx.Controller(),
			Zone:                    game.ZoneExile,
			Duration:                ctx.Game.UntilEndOfTurnDuration(),
			AltCostKey:              game.AltCostKeyGrantedCost,
			CastOnly:                true,
			ExileInsteadOfGraveyard: true,
			Source:                  ctx.Source(),
			SourceName:              "Radiant Scrollwielder",
			Label:                   "Radiant Scrollwielder — cast " + c.Name + " from exile this turn",
		}, []game.Card{c})
	}
	return nil
})

var _ = uuid.Nil
