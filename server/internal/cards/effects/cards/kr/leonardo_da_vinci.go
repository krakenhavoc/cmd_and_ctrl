package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leonardo da Vinci — Legendary Creature — Human Artificer {2}{U}, 3/3
// (Assassin's Creed, #5948):
//
//	"{3}{U}{U}: Until end of turn, Thopters you control have base power and
//	 toughness X/X, where X is the number of cards in your hand.
//	 {2}{U}, {T}: Draw a card, then discard a card. If the discarded card was
//	 an artifact card, exile it from your graveyard. If you do, create a token
//	 that's a copy of it, except it's a 0/2 Thopter artifact creature with
//	 flying in addition to its other types."
//
// The first ability is a layer 7b set of base power and toughness on the
// Thopters you control as it resolves (CR 611.2c locks the set: a Thopter that
// arrives later is not affected), X being the cards in your hand as it
// resolves and fixed then. The second is Glacial Dragonhunt's loot (the
// discard is not optional). The discarded card is read as an artifact card as
// it was in the hand; the exile is a clause whose success the copy depends on
// ("If you do": a commander that goes to the command zone instead is not
// exiled and makes no copy), and the copy is of the exiled card's copiable values
// (CR 707.2) with the 0/2 Thopter exception (CR 707.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92a287a0-b3cc-4040-a530-fb71510cfc67",
		Name:         "Leonardo da Vinci",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{U}{U}: Until end of turn, Thopters you control have base power and toughness X/X, where X is the number of cards in your hand.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{3}{U}{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := 0
					if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
						x = p.Hand.Size()
					}
					return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.YouControl(), effects.HasSubtype("Thopter")),
						"Leonardo da Vinci — Thopters you control have base power and toughness X/X",
						game.SetBasePowerMod(x), game.SetBaseToughnessMod(x))
				},
			},
			{
				Label:   "{2}{U}, {T}: Draw a card, then discard a card. If the discarded card was an artifact card, exile it from your graveyard. If you do, create a token that's a copy of it, except it's a 0/2 Thopter artifact creature with flying in addition to its other types.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: leonardoDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				},
			},
		},
	})
}

var (
	// The loot's discard half, run once the draw has landed. An on-disk key.
	leonardoDiscard = game.RegisterDrawThen("leonardo-da-vinci/discard", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        1,
			Question: "Leonardo da Vinci — discard a card",
		}, leonardoDiscarded, nil)
	})

	// The discard's answer: an artifact card discarded is exiled from the
	// graveyard, and only an exile that happened makes the copy. An on-disk key.
	leonardoDiscarded = game.RegisterDiscardThen("choose-cards/leonardo-da-vinci-discarded", func(g *game.Game, r game.DiscardsDone) error {
		parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Leonardo da Vinci"}
		ctx := effects.NewContext(g, parent)
		for _, id := range r.Discarded.Cards() {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !c.IsArtifact() {
				continue
			}
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			return effects.ExileTarget{Target: id, ThenTail: effects.WithCarry(ctx.ItemTail(leonardoExiled), id)}.Apply(ctx)
		}
		return nil
	})

	// "If you do, create a token that's a copy of it": the exile happened.
	// An on-disk key.
	leonardoExiled = effects.CardTail("leonardo-da-vinci/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
		if len(o.IDs) != 1 {
			return nil
		}
		return effects.CreateTokenCopy{
			Controller: ctx.Controller(),
			Copy:       effects.TailCarry0(o),
			N:          1,
			Except:     effects.GapB39ThopterCopyException(ctx.TextSubtype("Thopter")),
		}.Apply(ctx)
	})
)
