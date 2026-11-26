package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b447_helpers.go — shared bodies for the Innistrad batch 447 cards
// (S82 OLD, #5572).

// OldB447TargetExilesCardGainIfCreature is Graveyard Shovel's "Target player
// exiles a card from their graveyard. If it's a creature card, you gain 2
// life": the TARGET player chooses from their own graveyard as the ability
// resolves, and its controller gains the life when the card exiled was a
// creature card, read as it left the graveyard (CR 608.2h).
func OldB447TargetExilesCardGainIfCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := TargetedPlayerOf(ctx)
	if player == uuid.Nil {
		return nil
	}
	cards := GraveyardCardsOfPlayer(g, player)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   "Graveyard Shovel — exile a card from your graveyard",
		Cards:      cards,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    oldB447ShovelExiled,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// oldB447ShovelExiled is the target player's pick. An on-disk key: never
// renamed, never reused.
var oldB447ShovelExiled = AskedPick("choose-cards/old-b447-graveyard-shovel-exile", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		return ExileThenIfItWas{Target: id, Was: WasCreatureCard, Keyed: oldB447ShovelLife}.Apply(ctx)
	}
	return nil
})

// oldB447ShovelLife is "if it's a creature card, you gain 2 life": it runs only
// when the card reached exile and was a creature card. An on-disk key: never
// renamed, never reused.
var oldB447ShovelLife = IfItWasTail("if-it-was/old-b447-graveyard-shovel-life", func(ctx *Context) error {
	return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
}, nil)

// OldB447ReturnTwoAtRandom is "Return two cards at random from your graveyard to
// your hand" (Make a Wish): two distinct cards from a persisted random stream,
// or all of them when the graveyard holds fewer.
func OldB447ReturnTwoAtRandom(_ *game.StackItem, ctx *Context) error {
	ids := GraveyardCardsOfPlayer(ctx.Game, ctx.Controller())
	picks := ctx.Game.ChooseAtRandomForEffect(game.RandomDraw{Player: ctx.Controller(), Source: ctx.Source()}, ids, 2)
	for _, id := range picks {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
