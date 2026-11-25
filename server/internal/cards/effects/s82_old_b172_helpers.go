package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b172_helpers.go — shared bodies for the S82 older-sets batch 172
// (War of the Spark, #5520).

// S82OldB172TargetPlayerDiscardsTwoAndLosesTwo is Davriel's Shadowfugue: the
// target player picks and discards two cards (CR 701.9a) and loses 2 life.
// Life loss is not damage.
func S82OldB172TargetPlayerDiscardsTwoAndLosesTwo(item *game.StackItem, ctx *Context) error {
	var who uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			who = t.ID
		}
	}
	if who == uuid.Nil {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: who, Source: item.SourceCardID, N: 2})
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), who, -2)
}

// S82OldB172TapAllOpponentCreaturesAndGrantLifelink is Bond of Discipline: tap
// every creature your opponents control, then creatures you control gain
// lifelink until end of turn (the set read as it resolves, CR 611.2c).
func S82OldB172TapAllOpponentCreaturesAndGrantLifelink(item *game.StackItem, ctx *Context) error {
	if err := TapAllCreaturesYourOpponentsControl(ctx.Game, item); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl()),
		Keywords: []string{"lifelink"},
		Label:    "Bond of Discipline — creatures you control gain lifelink",
	}.Apply(ctx)
}

// S82OldB172IsInstantOrSorcery is "instant and sorcery card".
func S82OldB172IsInstantOrSorcery(c game.Card) bool {
	return c.IsInstant() || c.IsSorcery()
}

// S82OldB172VivienTake is Vivien's Grizzly's look: the top card, if it is a
// creature or planeswalker card you may reveal it and put it into your hand;
// if you don't, it goes on the bottom of your library. The registered take's
// rest rider puts what is left on the bottom (one card, so the order is
// moot).
var S82OldB172VivienTake = RegisterTake("choose-cards/s82-old-b172-viviens-grizzly-take",
	TakeDef{Reveal: true, Then: TakeRestOnBottomInRandomOrder})

// S82OldB172VivienLook is Vivien's Grizzly's whole ability.
func S82OldB172VivienLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player:   player,
		Cards:    g.LookAtTopOfLibraryForEffect(player, 1),
		Match:    Or(Creature(), Planeswalker()),
		Max:      1,
		Optional: true,
		Label:    "Vivien's Grizzly — you may reveal the creature or planeswalker card and put it into your hand",
		Key:      S82OldB172VivienTake,
	}.Apply(ctx)
}

// S82OldB172ArkbowDig is Vivien's Arkbow's resolution: look at the top X
// cards, you may put a creature card with mana value X or less from among
// them onto the battlefield, the rest on the bottom in a random order. X is
// the {X} paid.
func S82OldB172ArkbowDig(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := ctx.X()
	player := ctx.Controller()
	cards := g.LookAtTopOfLibraryForEffect(player, x)
	if len(cards) == 0 {
		return nil
	}
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    cards,
		Match:    And(Creature(), ManaValueLE(x)),
		Max:      1,
		Optional: true,
		Label:    "Vivien's Arkbow — you may put a creature card with mana value X or less onto the battlefield",
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}
