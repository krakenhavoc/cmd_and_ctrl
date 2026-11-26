package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b524_helpers.go — Eventide, graveyard recursion (S82 older sets,
// batch 524, #5590).

// S82OldB524NoOpponentHasCardsInHand is the test of "if an opponent has no
// cards in hand" (Rekindled Flame) for `you`: some opponent still in the game
// holds an empty hand.
func S82OldB524NoOpponentHasCardsInHand(g *game.Game, you uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == you || p.Eliminated || p.Hand == nil {
			continue
		}
		if p.Hand.Size() == 0 {
			return true
		}
	}
	return false
}

// S82OldB524EvershrikeReturn is Evershrike's graveyard ability: "Return this
// card from your graveyard to the battlefield. You may put an Aura card with
// mana value X or less from your hand onto the battlefield attached to it. If
// you don't, exile this creature." X is the announced X. The return is one
// entry; the Aura offer and the exile are its registered tail, run once the
// card has entered.
func S82OldB524EvershrikeReturn(g *game.Game, item *game.StackItem, x int) error {
	ctx := NewContext(g, item)
	return ReturnFromGraveyardTogether{
		Targets:  []uuid.UUID{item.SourceCardID},
		ThenTail: s82OldB524EvershrikeEntered,
		ThenN:    x,
	}.Apply(ctx)
}

// s82OldB524EvershrikeEntered offers the Aura (mana value at most N, from the
// controller's hand) once Evershrike is back, or exiles it when there is none
// to offer. A card that did not enter (it left the graveyard in response) does
// nothing. An on-disk key: never renamed, never reused.
var s82OldB524EvershrikeEntered = TogetherTail("then/s82-old-b524-evershrike-entered", func(ctx *Context, entered []uuid.UUID, n int) error {
	if len(entered) != 1 {
		return nil
	}
	host, g, you := entered[0], ctx.Game, ctx.Controller()
	var cands []uuid.UUID
	if p := g.PlayerByIDForEffect(you); p != nil && !p.Eliminated && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			if c.IsAura() && !c.IsToken() && c.ManaValue() <= n {
				cands = append(cands, c.InstanceID)
			}
		}
	}
	if len(cands) == 0 {
		return s82OldB524ExileHost(g, host)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefield,
		Chooser:     you,
		Source:      ctx.Source(),
		Question:    "Evershrike — you may put an Aura card with mana value " + strconv.Itoa(n) + " or less from your hand onto the battlefield attached to it (or exile it)",
		Cards:       cands,
		Min:         0,
		Max:         1,
		ThenKey:     s82OldB524EvershrikeAura,
		Carry:       []uuid.UUID{host},
	})
	return nil
})

// s82OldB524EvershrikeAura puts the picked Aura onto the battlefield attached
// to Evershrike; with none picked, or an Aura that could not enchant it and so
// stayed in hand (CR 303.4i), Evershrike is exiled ("If you don't").
var s82OldB524EvershrikeAura = game.RegisterChooseCardsThen("choose-cards/s82-old-b524-evershrike-aura",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Carry) == 0 {
			return nil
		}
		host := r.Carry[0]
		if len(r.Picked) == 0 {
			return s82OldB524ExileHost(g, host)
		}
		aura := r.Picked[0]
		if err := g.PutOntoBattlefieldTogetherKeyedForEffect(
			[]game.BatchEntry{{CardID: aura, From: game.ZoneHand}},
			game.ZoneEntryOptions{Controller: r.Chooser, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}},
			game.KeyedTail{}); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(aura); z != nil && z.Kind == game.ZoneBattlefield {
			return nil
		}
		return s82OldB524ExileHost(g, host)
	})

// s82OldB524ExileHost exiles Evershrike from the battlefield, if it is still
// there (CR 608.2b).
func s82OldB524ExileHost(g *game.Game, host uuid.UUID) error {
	if z := g.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return g.ExileCardForEffect(host)
}

// S82OldB524TargetExilesOwnGraveyardCard is "Target player exiles a card from
// their graveyard" for the named card (Merrow Bonegnawer): the TARGET player
// chooses, from their own graveyard, as the ability resolves. It is
// OldB178TargetExilesOwnGraveyardCard with the prompt naming the card that asks.
func S82OldB524TargetExilesOwnGraveyardCard(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player, ok := FirstLegalPlayerTarget(ctx)
		if !ok {
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
			Question:   name + " — exile a card from your graveyard",
			Cards:      cards,
			Min:        1,
			Max:        1,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    oldB178ClawsExiled,
			Asked:      game.NewAskedItem(item),
		})
		return nil
	}
}
