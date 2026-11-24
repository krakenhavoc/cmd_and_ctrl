package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcane Artisan — Creature — Human Wizard {2}{U}, 0/3 (Kaldheim Commander,
// #5506):
//
//	"{2}{U}, {T}: Target player draws a card, then exiles a card from their
//	 hand. If a creature card is exiled this way, that player creates a
//	 token that's a copy of that card.
//	 When this creature leaves the battlefield, exile all tokens created
//	 with it at the beginning of the next end step."
//
// The target player draws, then CHOOSES the card from their own hand
// (a keyed card prompt, Fiend of the Shadows' shape); a creature card
// exiled that way is copied as a token under THAT player's control, each
// token stamped with the Artisan OBJECT that made it (Card.CreatedWith,
// CR 607.1d; Faerie Artisans). The leaves trigger schedules a delayed
// end-step trigger carrying that object; it exiles whatever tokens are
// linked to it when the end step begins, so a copy made by an ability that
// resolves after the Artisan left is exiled too, and a new Artisan is a
// different object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7dbf5f59-1246-4156-bb21-090cfddd9111",
		Name:         "Arcane Artisan",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Target player draws a card, then exiles a card from their hand. If a creature card is exiled this way, that player creates a token that's a copy of that card.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  arcaneArtisanActivate,
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Arcane Artisan — exile all tokens created with it at the beginning of the next end step",
				arcaneArtisanLeaves),
		},
	})
}

func arcaneArtisanActivate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := effects.TargetedPlayer(ctx)
	if player == uuid.Nil {
		return nil
	}
	if err := (effects.DrawCards{Player: player, N: 1}).Apply(ctx); err != nil {
		return err
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	cards := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		cards = append(cards, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     item.SourceCardID,
		Question:   "Arcane Artisan — exile a card from your hand",
		Cards:      cards,
		Min:        1,
		Max:        1,
		ThenKey:    arcaneArtisanExile,
		Carry:      []uuid.UUID{ref.ID},
		CarryNums:  []int{ref.Epoch},
		LiveZone:   game.ZoneHand,
	})
	return nil
}

// arcaneArtisanExile is the keyed answer: exile the pick, then (once it is
// really in exile) the copy. An on-disk key: never renamed, never reused.
var arcaneArtisanExile = game.RegisterChooseCardsThen("choose-cards/arcane-artisan-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 || len(r.Nums) == 0 || len(r.Picked) == 0 {
		return nil
	}
	return g.ExileCardKeyedForEffect(r.Picked[0], game.KeyedTail{
		Ref: arcaneArtisanCopy, Player: r.Chooser, Source: r.Source,
		Carry: []uuid.UUID{r.Carry[0]}, N: r.Nums[0],
	})
})

// arcaneArtisanCopy is "if a creature card is exiled this way, that player
// creates a token that's a copy of that card": it runs only for a card that
// reached exile, and the token is linked to the Artisan object. An on-disk
// key.
var arcaneArtisanCopy = effects.CardTail("arcane-artisan/copy-exiled-creature", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	g := ctx.Game
	c, ok := g.LookupCardForEffect(o.IDs[0])
	if !ok || !c.IsCreature() {
		return nil
	}
	tmpl, ok := effects.TokenCopyTemplate(g, o.IDs[0])
	if !ok {
		return nil
	}
	link := game.PermissionCardRef{ID: o.Tail.Carry[0], Epoch: o.Tail.N}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: o.Tail.Player,
		Source:     o.Tail.Source,
		Groups:     []game.TokenGroup{{Template: effects.CopiedToken(tmpl), Count: 1, CreatedWith: link}},
	}, game.KeyedTail{})
})

func arcaneArtisanLeaves(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Arcane Artisan — exile all tokens created with it",
		Body:   arcaneArtisanExileTokens,
		Params: game.EffectParams{Object: ref},
	}.Apply(ctx)
}

// arcaneArtisanExileTokens exiles every token linked to the Artisan object
// its params name, as the end step begins. An on-disk key.
var arcaneArtisanExileTokens = game.SimpleDelayedBody("arcane-artisan/exile-linked-tokens",
	func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, id := range g.TokensCreatedWithForEffect(game.PermissionCardRef(item.Params.Object)) {
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
