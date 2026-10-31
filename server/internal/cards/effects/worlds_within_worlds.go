package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worlds Within Worlds — Sorcery {5}{G}{U}:
//
//	"Exile all creatures. Each player may put any number of creature cards
//	 from their hand onto the battlefield. Then put all cards exiled this
//	 way into their owners' hands. Exile Worlds Within Worlds."
//
// Four instructions in the printed order (CR 608.2c), each waiting for the
// prompts the one before it asked:
//
//   - Exile all creatures: the sweep's tail is told the cards that really
//     reached exile (a commander its owner sent to the command zone is not
//     one), and the tokens among them are dropped: a token is not a card,
//     and ceases to exist in exile (CR 111.7), so there is nothing to put
//     into a hand.
//   - Each player may put any number of creature cards from their hand:
//     one pick per player in APNAP order (CR 101.4: the active player
//     chooses, then each other player in turn order, and the actions then
//     happen simultaneously), a player with no creature card in hand not
//     asked. The hand is read after the sweep, so a creature exiled this
//     way is not in it. The picks, the players still to ask and the exiled
//     cards ride the prompts as data, so a table waiting on any seat is a
//     restore point. CR 101.4a lets each pick stay face down; the picks
//     are hidden until they enter.
//   - Every pick enters together under its owner's control, one entry
//     (CR 603.6a), so each newcomer sees the others arrive.
//   - Then every card exiled this way that is still in exile goes to its
//     owner's hand: it is the card that was exiled (CR 400.7), not one
//     that left and came back, and not a commander that went to the
//     command zone.
//   - Exile Worlds Within Worlds: the spell exiles itself last, and
//     spellMovedItselfLocked stops the resolution frame from putting it in
//     the graveyard afterwards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e25062ba-7176-4d59-ac3e-26d19b26f02a",
		Name:         "Worlds Within Worlds",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ExileAllMatching{Match: Creature(), ThenTail: ctx.ItemTail(worldsWithinWorldsExiledTail)}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

// worldsWithinWorldsExiledTail starts the APNAP walk once every creature has
// finished moving, carrying the cards exiled this way (tokens excluded).
var worldsWithinWorldsExiledTail = MassTail("mass/worlds-within-worlds-exiled", func(ctx *Context, swept []game.Card, _ int) error {
	var exiled []uuid.UUID
	for _, c := range swept {
		if !c.IsToken() {
			exiled = append(exiled, c.InstanceID)
		}
	}
	return worldsAskNext(ctx, apnapPlayers(ctx.Game), nil, exiled)
})

// worldsAskNext asks the next player, in APNAP order, whose hand holds a
// creature card; once nobody is left it puts every pick onto the battlefield
// together and returns the exiled cards to their owners' hands. The walk's
// state is by value: the players still to ask, the picks so far and the
// exiled cards.
func worldsAskNext(ctx *Context, order, picks, exiled []uuid.UUID) error {
	g, item := ctx.Game, ctx.Item
	for i, player := range order {
		candidates := handCardsMatching(g, player, Creature())
		if len(candidates) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  player,
			Source:   item.SourceCardID,
			Question: "Worlds Within Worlds — put any number of creature cards from your hand onto the battlefield",
			Cards:    candidates,
			Min:      0,
			Max:      len(candidates),
			ThenKey:  worldsWithinWorldsPick,
			LiveZone: game.ZoneHand,
			Carry:    packIDLists(order[i+1:], picks, exiled),
			Asked:    game.NewAskedItem(item),
		})
		return nil
	}
	entries := make([]game.BatchEntry, 0, len(picks))
	for _, id := range picks {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneHand})
		}
	}
	if len(entries) == 0 {
		return worldsReturnExiled(ctx, exiled)
	}
	then := ctx.ItemTail(worldsWithinWorldsReturnTail)
	then.Carry = exiled
	// Each card enters under its owner's control, put by that player
	// (CR 110.2a).
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{PutByEnteringPlayer: true}, then)
}

// worldsWithinWorldsPick is one player's pick answered, keyed: the next
// player is asked, or the picks enter together. Registered in init because
// it asks the next player, which names the key again.
var worldsWithinWorldsPick game.ChooseCardsThen

func init() {
	worldsWithinWorldsPick = askedPick("choose-cards/worlds-within-worlds-pick", func(ctx *Context, r game.CardsChosen) error {
		lists := unpackIDLists(r.Carry)
		for len(lists) < 3 {
			lists = append(lists, nil)
		}
		rest, picks, exiled := lists[0], lists[1], lists[2]
		return worldsAskNext(ctx, rest, append(append([]uuid.UUID(nil), picks...), r.Picked...), exiled)
	})
}

// worldsWithinWorldsReturnTail is the entry's "then": the exiled cards go to
// their owners' hands. An on-disk key.
var worldsWithinWorldsReturnTail = CardTail("worlds-within-worlds/return", func(ctx *Context, o game.TailOutcome) error {
	return worldsReturnExiled(ctx, o.Tail.Carry)
})

// worldsReturnExiled puts each of the cards exiled this way that is still in
// exile into its owner's hand.
func worldsReturnExiled(ctx *Context, exiled []uuid.UUID) error {
	g := ctx.Game
	for _, id := range exiled {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if _, err := g.ReturnFromAnyZoneToHandForEffect(ctx.Controller(), ctx.Item.SourceCardID, id); err != nil {
			return err
		}
	}
	return nil
}
