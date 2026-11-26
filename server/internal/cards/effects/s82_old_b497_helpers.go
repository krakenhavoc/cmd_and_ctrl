package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b497_helpers.go — helpers for Zendikar, batch 497 (S82 older
// sets, #5583).

// OldB497ThisWouldDieToLibraryBottom is "If this creature would die, put it on
// the bottom of its owner's library instead" (Nissa's Chosen): a move of the
// source from the battlefield to a graveyard becomes a move to the bottom of
// its owner's library (CR 614.1a, CR 700.4).
func OldB497ThisWouldDieToLibraryBottom(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventMove && ev.CardID == src.InstanceID &&
				ev.OldZone == game.ZoneBattlefield && ev.NewZone == game.ZoneGraveyard
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneLibrary
			ev.NewZoneOwner = uuid.Nil
			ev.PutOnLibraryBottom = true
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label:    label + ": bottom of its owner's library instead of dying",
		RowLabel: "If this creature would die, put it on the bottom of its owner's library instead.",
	}
}

// OldB497HasSubtypeCard is a CardPredicate for "a <subtype> card".
func OldB497HasSubtypeCard(sub string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtype(sub) }
}

// OldB497TrapfindersTrick is "Target player reveals their hand and discards
// all Trap cards."
func OldB497TrapfindersTrick(_ *game.StackItem, ctx *Context) error {
	t, ok := FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(t.ID)
	if p == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(t.ID)
	var ids []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasSubtype("Trap") {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(t.ID, ids, ctx.Source(), game.KeyedTail{})
}

// OldB497WorldQuellerNone is the first, always-legal branch of World
// Queller's prompt: "you may choose a card type" declined.
const OldB497WorldQuellerNone = "Don't choose a card type"

// OldB497WorldQuellerTypes are the card types a player can name, in the
// order a type line lists them.
var OldB497WorldQuellerTypes = []string{"Artifact", "Battle", "Creature", "Enchantment", "Land", "Planeswalker"}

// OldB497WorldQuellerUpkeep is World Queller's trigger: its controller may
// choose a card type.
func OldB497WorldQuellerUpkeep(g *game.Game, item *game.StackItem) error {
	options := []game.ChoiceOption{{Label: OldB497WorldQuellerNone}}
	for _, t := range OldB497WorldQuellerTypes {
		options = append(options, game.ChoiceOption{Label: t})
	}
	g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "World Queller — you may choose a card type; each player sacrifices a permanent of that type",
		Options:  options,
		ThenKey:  OldB497WorldQuellerChosen,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// OldB497WorldQuellerChosen is "each player sacrifices a permanent of their
// choice of that type" once a type is chosen. An on-disk key: never renamed,
// never reused.
var OldB497WorldQuellerChosen = game.RegisterOptionPickThen("option-pick/s82-old-b497-world-queller-type", func(g *game.Game, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Label == OldB497WorldQuellerNone {
		return nil
	}
	want := strings.ToLower(r.Option.Label)
	return EachPlayerSacrifices{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasCardType(want) },
		Label: "a " + want + " permanent",
	}.Apply(NewContext(g, r.Asked))
})

// OldB497ChandraDiscard is Chandra Ablaze's +1: "Discard a card. If a red
// card is discarded this way, Chandra deals 4 damage to any target." The
// target is chosen as the ability goes on the stack (CR 601.2c); the discard
// is the first instruction.
func OldB497ChandraDiscard(g *game.Game, item *game.StackItem) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        1,
		Question: "Chandra Ablaze — discard a card",
		Asked:    game.NewAskedItem(item),
	}, OldB497ChandraDiscardThen, nil)
}

// OldB497ChandraDiscardThen deals the 4 damage if the card discarded is red.
// An on-disk key: never renamed, never reused.
var OldB497ChandraDiscardThen = game.RegisterDiscardThenRule("choose-cards/s82-old-b497-chandra-ablaze-discard", nil,
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Item == nil {
			return nil
		}
		for _, id := range r.Discarded.By(r.Player) {
			c, ok := g.LookupCardForEffect(id)
			ctx := NewContext(g, r.Item)
			if ok && c.HasColor(ctx.TextColor("R")) {
				return DamageToFirstTargetStep(4)(r.Item, ctx)
			}
		}
		return nil
	})

// OldB497ChandraDiscardHands is the first instruction of Chandra Ablaze's −2:
// each player discards their hand.
func OldB497ChandraDiscardHands(ctx *Context) error {
	for _, id := range TablePlayers(ctx) {
		if _, err := DiscardWholeHand(ctx.Game, id); err != nil {
			return err
		}
	}
	return nil
}

// OldB497ChandraDrawThree is its second: each player draws three cards. It is
// a step of its own so the draws wait for any prompt a discard asks (CR
// 608.2c).
func OldB497ChandraDrawThree(ctx *Context) error {
	for _, id := range TablePlayers(ctx) {
		if err := (DrawCards{Player: id, N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// OldB497ChandraCastReds is Chandra Ablaze's −7: "Cast any number of red
// instant and/or sorcery cards from your graveyard without paying their mana
// costs." Each red instant and sorcery card in the graveyard is granted its
// own free cast, offered during the resolution (CR 608.2g).
func OldB497ChandraCastReds(g *game.Game, item *game.StackItem) error {
	const name = "Chandra Ablaze"
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	var granted []uuid.UUID
	red := NewContext(g, item).TextColor("R")
	for _, c := range p.Graveyard.Cards {
		if !c.HasColor(red) || !(c.HasCardType("instant") || c.HasCardType("sorcery")) {
			continue
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      item.Controller,
			Zone:        game.ZoneGraveyard,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			AltCostKey:  game.AltCostKeyGrantedCost,
			Timing:      game.TimingFlash,
			CastOnly:    true,
			LapseOnPass: game.LapseStaysInHand,
			Source:      item.SourceCardID,
			SourceName:  name,
			Label:       name + " — cast " + c.Name + " without paying its mana cost",
		}, []game.Card{c})
		granted = append(granted, c.InstanceID)
	}
	if len(granted) == 0 {
		return nil
	}
	OfferResolutionCastOver(g, item.Controller, item.SourceCardID, name, game.MayCastKeywordFreeCast, DeclineLeaveInGraveyard, granted...)
	return nil
}
