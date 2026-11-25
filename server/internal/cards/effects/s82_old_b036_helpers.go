package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b036_helpers.go — shared pieces of S82 older-sets batch 36
// (Modern Horizons 2: suspend, damage).

// OldB036Glimpse is Glimpse of Tomorrow's resolution, first half: "Shuffle all
// permanents you own into your library." Every permanent the caster owns, whoever
// controls it, goes to the library as one simultaneous move; the rest of the
// spell runs once it has settled (the tuck can pause on a commander's
// CR 903.9 offer). Tokens are counted: they are shuffled into the library for a
// moment (CR 111.7) and are part of "that many".
func OldB036Glimpse(_ *game.StackItem, ctx *Context) error {
	g := ctx.Game
	player := ctx.Controller()
	var ids []uuid.UUID
	tokens := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Owner != player {
			continue
		}
		ids = append(ids, c.InstanceID)
		if IsToken(c) {
			tokens++
		}
	}
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{},
		game.KeyedTail{Ref: OldB036GlimpseShuffled, Player: player, Source: ctx.Source(), N: tokens})
}

// OldB036GlimpseShuffled is "then reveal that many cards from the top of your
// library", once the permanents are in the library and it is shuffled. The
// number is the permanents that reached the library, plus the tokens (which
// cease to exist there). Then the non-Aura permanent cards revealed enter
// together. An on-disk key: never renamed, never reused.
var OldB036GlimpseShuffled = game.RegisterTail("old-b036/glimpse-of-tomorrow-shuffled", func(g *game.Game, o game.TailOutcome) error {
	player := o.Tail.Player
	if err := g.ShuffleLibraryForEffect(player); err != nil {
		return err
	}
	n := len(o.IDs) + o.Tail.N
	revealed := g.RevealTopOfLibraryForEffect(player, o.Tail.Source, n,
		"Glimpse of Tomorrow — revealed from the top of the library")
	ctx := KeyedContext(g, player, o.Tail.Source)
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    revealed,
		Match:    Not(oldB036IsAura()),
		All:      true,
		Label:    "Glimpse of Tomorrow — put all non-Aura permanent cards revealed this way onto the battlefield",
		ThenTail: OldB036GlimpseAuras,
	}.Apply(ctx)
})

// OldB036GlimpseAuras is "then do the same for Aura cards": the Auras among
// the cards not put onto the battlefield enter, each attached to something it
// can enchant (an Aura with nothing to enchant stays in the library), then the
// rest go on the bottom in a random order.
var OldB036GlimpseAuras = PutFromLibraryTail("old-b036/glimpse-of-tomorrow-auras", func(ctx *Context, res PutFromLibraryResult) error {
	return PutFromLibraryOntoBattlefield{
		Player:   res.Player,
		Cards:    res.Rest,
		Match:    oldB036IsAura(),
		All:      true,
		Label:    "Glimpse of Tomorrow — put all Aura cards revealed this way onto the battlefield",
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
})

func oldB036IsAura() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsAura() }
}

// OldB036CardTypeCount is the number of CR 205.2a card types the card has.
func OldB036CardTypeCount(c game.Card) int {
	n := 0
	for _, t := range game.ChoosableCardTypes {
		if c.HasCardType(strings.ToLower(t)) {
			n++
		}
	}
	return n
}

// OldB036ManticoreBlast is Mount Velus Manticore's reflexive trigger: the
// Manticore deals X damage to any target, X fixed when the card was discarded.
var OldB036ManticoreBody = game.ReflexiveBody("old-b036/mount-velus-manticore-strike", BreechesBlastEffect, ConstTargets(TargetAny))

// OldB036ManticoreDiscarded is the optional discard answered: with a card
// really discarded, the reflexive trigger is made with X the number of card
// types it has. An on-disk key.
var OldB036ManticoreDiscarded = game.RegisterDiscardThen("choose-cards/old-b036-mount-velus-manticore-strike", func(g *game.Game, r game.DiscardsDone) error {
	cards := r.Discarded.Cards()
	if len(cards) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(cards[0])
	if !ok {
		return nil
	}
	x := OldB036CardTypeCount(c)
	if x <= 0 {
		return nil
	}
	t := WhenYouDo("Mount Velus Manticore — it deals damage to any target equal to the card types of the discarded card", OldB036ManticoreBody)
	t.Params = game.EffectParams{Amount: x}
	return t.Apply(KeyedContext(g, r.Player, r.Source))
})

// OldB036ManticoreAsk is "At the beginning of combat on your turn, you may
// discard a card."
func OldB036ManticoreAsk(g *game.Game, item *game.StackItem) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        1,
		UpTo:     true,
		Question: "Mount Velus Manticore — you may discard a card; when you do, it deals damage equal to the card types of the discarded card to any target",
	}, OldB036ManticoreDiscarded, nil)
}

// OldB036PiruUpkeep is "At the beginning of your upkeep, sacrifice Piru unless
// you pay {R}{W}{B}." A Piru that left, or came back as a new object, asks
// nothing.
func OldB036PiruUpkeep(g *game.Game, item *game.StackItem) error {
	if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	return UpkeepPayUnless{
		Chooser:  item.Controller,
		Cost:     "{R}{W}{B}",
		Question: "Piru, the Volatile — pay {R}{W}{B} or sacrifice Piru?",
		Else:     UpkeepSacrificeSource,
	}.Apply(NewContext(g, item))
}

// OldB036PiruDies is "When Piru dies, it deals 7 damage to each nonlegendary
// creature": Piru, as it last existed, is the source; each creature is judged by
// its own legendary status as the damage is dealt.
func OldB036PiruDies(g *game.Game, item *game.StackItem) error {
	return DamageEachMatching(NewContext(g, item), And(Creature(), Not(Legendary())), 7)
}
