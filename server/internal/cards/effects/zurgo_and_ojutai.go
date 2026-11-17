package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zurgo and Ojutai — Legendary Creature — Orc Dragon {2}{U}{R}{W}, 4/4
// (March of the Machine Commander):
//
//	"Flying, haste
//	 Zurgo and Ojutai has hexproof as long as it entered this turn.
//	 Whenever one or more Dragons you control deal combat damage to a player
//	 or battle, look at the top three cards of your library. Put one of them
//	 into your hand and the rest on the bottom of your library in any order.
//	 You may return one of those Dragons to its owner's hand."
//
// The trigger is one per damaged player or battle in a combat damage step
// (CR 603.2c). "Those Dragons" are the Dragons that dealt that damage in the
// batch, filed with the resolution before the look so a prompt in between
// (the order of the rest) does not lose them; the return is asked after the
// library has been dealt with, and only offers a Dragon still on the
// battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "941c9acd-2773-456a-88d5-0537735f93c6",
		Name:            "Zurgo and Ojutai",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Purpose:         game.Purpose{Draws: 1},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && g.EnteredThisTurn(source.InstanceID)
			}, "hexproof"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventDealDamage,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return s82CombatDamageToPlayerOrBattleBy(ev, source.Controller, g, s82IsDragon)
				},
				"Zurgo and Ojutai — look at the top three cards, put one into your hand and the rest on the bottom, then you may return one of those Dragons to its owner's hand",
				Do(
					Instruction(zurgoOjutaiLook),
					Instruction(zurgoOjutaiOfferReturn),
				))),
		},
	})
}

func s82IsDragon(c game.Card) bool { return c.HasSubtypeWord("Dragon") }

const zurgoOjutaiDragons = "zurgo-and-ojutai/dragons"

// zurgoOjutaiLook files the Dragons that dealt the damage, then looks at the
// top three, takes one and puts the rest on the bottom in the order chosen.
func zurgoOjutaiLook(ctx *Context) error {
	ctx.RecordStep(zurgoOjutaiDragons, s82BatchDamageDealers(ctx.Game, ctx.Item.Trigger.Event, ctx.Controller(), s82IsDragon))
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
		Max:    1,
		Label:  "Zurgo and Ojutai — put one of them into your hand",
		Key:    takeRestBottomAny,
	}.Apply(ctx)
}

// zurgoOjutaiOfferReturn is "you may return one of those Dragons to its
// owner's hand", over the Dragons still on the battlefield.
func zurgoOjutaiOfferReturn(ctx *Context) error {
	var onBattlefield []uuid.UUID
	for _, id := range ctx.StepRecord(zurgoOjutaiDragons).IDs {
		if ctx.Game.Battlefield.Contains(id) {
			onBattlefield = append(onBattlefield, id)
		}
	}
	if len(onBattlefield) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Item.SourceCardID,
		Question: "Zurgo and Ojutai — you may return one of those Dragons to its owner's hand",
		Cards:    onBattlefield,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  zurgoOjutaiReturned,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// zurgoOjutaiReturned is the answer: the chosen Dragon goes to its owner's
// hand. An on-disk key: never renamed, never reused.
var zurgoOjutaiReturned = askedPick("choose-cards/zurgo-and-ojutai-return", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
