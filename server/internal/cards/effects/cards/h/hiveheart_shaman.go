package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hiveheart Shaman — {3}{G} Creature — Human Shaman 3/5 (Innistrad:
// Crimson Vow, S82):
//
//	"Whenever this creature attacks, you may search your library for a
//	 basic land card that doesn't share a land type with a land you
//	 control, put that card onto the battlefield, then shuffle.
//	 {5}{G}: Create a 1/1 green Insect creature token. Put X +1/+1
//	 counters on it, where X is the number of basic land types among
//	 lands you control. Activate only as a sorcery."
//
// The search predicate is built as the ability resolves: the basic land
// types among the lands its controller has now (read through the layered
// land, so a land that gained or lost a type counts as it is), and a
// candidate qualifies when it is a basic land card with none of them.
// Wastes has no land type, so it always qualifies. The land enters
// untapped. The activated ability makes the token and then puts the
// counters on it, X read when the ability resolves (Domain's count,
// DomainCount); Fractal Harness's keyed token-then-counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3780529-d47e-48dc-86f6-c1a282c3b656",
		Name:         "Hiveheart Shaman",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Lands: 1, LandsUntapped: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Hiveheart Shaman — you may search your library for a basic land card that doesn't share a land type with a land you control, put it onto the battlefield, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: vowBasicLandSharingNoTypeWithYourLands(g, item.Controller),
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Hiveheart Shaman — a basic land card that doesn't share a land type with a land you control, onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{5}{G}: Create a 1/1 green Insect creature token. Put X +1/+1 counters on it, " +
				"where X is the number of basic land types among lands you control. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{5}{G}"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: item.Controller,
					Source:     item.SourceCardID,
					Groups: []game.TokenGroup{{
						Template: ctx.TextToken(effects.TokenCard("1/1 green Insect")),
						Count:    1,
					}},
				}, effects.WithN(ctx.ItemTail(hiveheartShamanInsectCreated), effects.DomainCount(g, item.Controller)))
			},
		}},
	})
}

// hiveheartShamanInsectCreated puts the X counters on the Insect that
// arrived; X rides the tail as its N. An on-disk key.
var hiveheartShamanInsectCreated = effects.CardTail("hiveheart-shaman/insect-created", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, o.IDs[0], game.CounterPlusOne, o.Tail.N, game.KeyedTail{})
})

// vowBasicLandSharingNoTypeWithYourLands is "a basic land card that
// doesn't share a land type with a land you control": a basic land card
// carrying none of the basic land types among `controller`'s lands now.
func vowBasicLandSharingNoTypeWithYourLands(g *game.Game, controller uuid.UUID) func(game.Card) bool {
	var held []string
	lands := g.BattlefieldCardsForEffect()
	for _, t := range game.BasicLandTypes {
		for _, c := range lands {
			if c.Controller == controller && c.IsLand() && c.HasSubtype(t) {
				held = append(held, t)
				break
			}
		}
	}
	return func(c game.Card) bool {
		if !effects.IsBasicLand(c) {
			return false
		}
		for _, t := range held {
			if c.HasSubtype(t) {
				return false
			}
		}
		return true
	}
}
