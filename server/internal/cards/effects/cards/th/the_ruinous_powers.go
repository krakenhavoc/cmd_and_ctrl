package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Ruinous Powers — Enchantment {2}{B}{R} (Warhammer 40,000 Commander):
//
//	"At the beginning of your upkeep, choose an opponent at random. Exile
//	 the top card of that player's library. Until end of turn, you may play
//	 that card and you may spend mana as though it were mana of any color
//	 to cast it. When you cast a spell this way, its owner loses life equal
//	 to its mana value."
//
// The opponent is drawn from the keyed random stream as the trigger
// resolves and announced in the log (Raving Dead's pick). The exile is
// LINKED (CR 607.2a) to this object and this ability, with a play
// permission for the controller until end of turn that spends mana as
// though it were any color (ExileWithPermission). The second trigger is
// the other half of that pair, as Prowl's is: a CAST from exile of a card
// this Ruinous Powers exiled with that ability. The spell's owner and its
// mana value (the X it was cast with included, CR 202.3e) are read as the
// trigger goes on the stack, so the loss is the cast spell's even if it
// is countered first (CR 608.2h). A land played this way is not cast and
// costs its owner nothing.
//
// No simplification.
func init() {
	const upkeepLabel = "The Ruinous Powers — choose an opponent at random, exile the top card of their library; you may play it this turn"
	const castLabel = "The Ruinous Powers — its owner loses life equal to its mana value"
	effects.Register(effects.Spec{
		OracleID:     "283054ed-5fbd-4202-bc02-294079aaff03",
		Name:         "The Ruinous Powers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep(upkeepLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				pick := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ctx.Opponents(), 1)
				if len(pick) != 1 {
					return nil
				}
				name := "an opponent"
				if p := g.PlayerByIDForEffect(pick[0]); p != nil && p.Name != "" {
					name = p.Name
				}
				g.EmitEvent(game.Event{
					Kind:   game.EventOptionChosen,
					Actor:  item.Controller,
					CardID: item.SourceCardID,
					Label:  "The Ruinous Powers: " + name + " (at random)",
				})
				p := g.PlayerByIDForEffect(pick[0])
				if p == nil || p.Library == nil {
					return nil
				}
				top, err := p.Library.Top()
				if err != nil {
					return nil
				}
				return effects.ExileWithPermission{Target: top.InstanceID, GrantTo: item.Controller, AnyColor: true, Linked: true}.Apply(ctx)
			}),
			{
				Watches: []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
						game.PlayedCardExiledWithRow(ev, game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch},
							game.AbilitySlotTriggered, upkeepLabel)
				},
				Key: castLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, castLabel)
					if spell, ok := g.LookupCardForEffect(ev.CardID); ok {
						item.Params.Player = spell.Owner
						item.Params.Amount, _ = g.ManaValueForEffect(spell)
					}
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Params.Player == uuid.Nil || item.Params.Amount <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Params.Player, -item.Params.Amount)
				},
			},
		},
	})
}
