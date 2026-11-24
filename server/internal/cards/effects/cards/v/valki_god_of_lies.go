package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valki, God of Lies // Tibalt, Cosmic Impostor — a modal double-faced card
// (oracle 907ae517…), one file for both faces; the back is the "<oracle>#1" row.
//
// Front, Legendary Creature — God {1}{B}, 2/1:
//
//	"When Valki enters, each opponent reveals their hand. For each opponent,
//	 exile a creature card they revealed this way until Valki leaves the
//	 battlefield.
//	 {X}: Choose a creature card exiled with Valki with mana value X. Valki
//	 becomes a copy of that card."
//
// Back, Legendary Planeswalker — Tibalt {5}{B}{R}, loyalty 5:
//
//	"As Tibalt enters, you get an emblem with 'You may play cards exiled with
//	 Tibalt, Cosmic Impostor, and you may spend mana as though it were mana of
//	 any color to cast those spells.'
//	 +2: Exile the top card of each player's library.
//	 −3: Exile target artifact or creature.
//	 −8: Exile all graveyards. Add {R}{R}{R}."
//
// Valki's exile is an ability of the object (OldB071ValkiEnters): the hands are
// revealed to the table, the controller picks a creature card for each opponent
// that has one, and the card is exiled until that Valki leaves, linked to it
// (CR 607.2a). The {X} ability is Assimilation Aegis's copy of a linked card, a
// layer-1 copy effect (CR 613.2a) with no end (CR 611.2a): a pick among the creature
// cards of mana value X Valki holds, none being no copy.
//
// Tibalt's emblem is its one standing effect, "play the cards exiled with Tibalt
// as though any color": every card his three exiling abilities put into exile
// is given that permission at once, held by the player who got the emblem
// (Tibalt records them as he enters, CR 114.2) and lasting while the card stays
// in exile, so it outlives Tibalt and does not follow a change of his control.
// The permission is a play (lands too) with mana of any color spent on the
// spells. The −8 makes {R}{R}{R} after the graveyards are exiled. The emblem
// object itself is not created: its permission is all it does, and it is given
// by the exiling abilities (the owner's call whether an emblem object is wanted).
//
// No simplification.
func init() {
	const oracle = "907ae517-22d5-4ac7-bc3a-3f4d5eaeeb57"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Valki, God of Lies",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(effects.OldB071ValkiExileKey, effects.OldB071ValkiEnters),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}: Choose a creature card exiled with Valki with mana value X. Valki becomes a copy of that card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{X}"),
			Effect:  effects.OldB071ValkiCopy,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Tibalt, Cosmic Impostor",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			return ctx.Game.SetChosenPlayerForEffect(card.InstanceID, card.Controller)
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Exile the top card of each player's library.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					perm := effects.OldB071TibaltPermission(effects.OldB071TibaltOwner(g, item), item.SourceCardID)
					for _, seat := range g.Seats {
						if seat.Eliminated {
							continue
						}
						if err := g.ExileTopWithPermissionKeyedForEffect(seat.ID, perm.Player, 1, perm, game.KeyedTail{}); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−3: Exile target artifact or creature.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					perm := effects.OldB071TibaltPermission(effects.OldB071TibaltOwner(g, item), item.SourceCardID)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.ExileTarget{Target: t.ID,
								ThenTail: game.KeyedTail{Ref: effects.OldB071TibaltGranted, Player: item.Controller, Source: item.SourceCardID, Permission: &perm}}.Apply(ctx)
						}
					}
					return nil
				},
			},
			{
				Label: "−8: Exile all graveyards. Add {R}{R}{R}.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					perm := effects.OldB071TibaltPermission(effects.OldB071TibaltOwner(g, item), item.SourceCardID)
					ids := effects.AllGraveyardsCardIDs(ctx, func(game.Card) bool { return true })
					rrr := ctx.Tail(effects.OldB071TibaltAddRRR)
					tail := game.KeyedTail{Ref: effects.OldB071TibaltGranted, Player: item.Controller, Source: item.SourceCardID, Permission: &perm, Next: &rrr}
					if len(ids) == 0 {
						return effects.AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
					}
					return g.ExileCardsKeyedForEffect(ids, tail)
				},
			},
		},
	})
}
