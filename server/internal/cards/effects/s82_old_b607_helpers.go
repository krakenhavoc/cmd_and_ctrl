package effects

import (
	"slices"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b607_helpers.go — Mirrodin, batch 607 (#5606): Second Sunrise, Wrench
// Mind, Brown Ouphe, Override, Timesifter, Tangleroot, Lightning Coils, Myr
// Incubator and Hum of the Radix.

// OldB607SecondSunrise is Second Sunrise's body: every artifact, creature,
// enchantment and land card in every graveyard that was put there from the
// battlefield this turn returns to the battlefield under its owner's control
// (each player returns the cards in their own graveyard), all at once. The set
// is read before the first card moves; a token is not a card.
func OldB607SecondSunrise(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsToken() || !(c.IsArtifact() || c.IsCreature() || c.HasCardType("enchantment") || c.HasCardType("land")) {
				continue
			}
			if PutIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID) {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	return PutCardsFromZoneTogether(ctx, game.ZoneGraveyard, ids, uuid.Nil, false)
}

// OldB607WrenchMindDiscard is Wrench Mind: the target player discards two
// cards unless they discard an artifact card (Teferi Akosa's "two, unless one
// of this kind" set rule, chosen by the discarding player).
func OldB607WrenchMindDiscard(item *game.StackItem, ctx *Context) error {
	player := TargetedPlayerOf(ctx)
	if player == uuid.Nil {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	need := min(2, p.Hand.Size())
	if need <= 0 {
		return nil
	}
	then := oldB607WrenchMindDiscardTwoOrArtifact
	if need == 1 {
		then = oldB607WrenchMindDiscardTheOne
	}
	return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   item.SourceCardID,
		N:        need,
		Min:      1,
		Question: "Wrench Mind — discard two cards, unless you discard an artifact card",
	}, then, nil)
}

var oldB607WrenchMindDiscardTwoOrArtifact = game.RegisterDiscardThenRule("choose-cards/old-b607-wrench-mind-two-or-artifact",
	func(picked []game.Card) bool {
		if len(picked) == 2 {
			return true
		}
		return len(picked) == 1 && picked[0].IsArtifact()
	},
	func(*game.Game, game.DiscardsDone) error { return nil })

var oldB607WrenchMindDiscardTheOne = game.RegisterDiscardThen("choose-cards/old-b607-wrench-mind-the-one",
	func(*game.Game, game.DiscardsDone) error { return nil })

// OldB607FromAnArtifactSource is "target activated ability from an artifact
// source" (Brown Ouphe): the ability's source is an artifact, read as the
// object it is, or last was.
func OldB607FromAnArtifactSource() StackItemPredicate {
	return func(g *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		if item.Kind != game.StackItemActivated {
			return false
		}
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		return ok && c.IsArtifact()
	}
}

// OldB607OverrideCounter is Override's "Counter target spell unless its
// controller pays {1} for each artifact you control", counted as it resolves.
func OldB607OverrideCounter(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	n := B03ArtifactsControlled(ctx.Game, ctx.Controller())
	if n <= 0 {
		return nil
	}
	cost := "{" + strconv.Itoa(n) + "}"
	return CounterUnlessPaid{
		StackID:  t.ID,
		Cost:     cost,
		Question: "Override — pay " + cost + " or your spell is countered",
	}.Apply(ctx)
}

// OldB607CastACreatureSpell is Tangleroot's condition: any player cast a
// creature spell (as it sits on the stack).
func OldB607CastACreatureSpell(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.IsCreature()
}

// OldB607CasterAddsG is "that player adds {G}": the caster named by the event
// gets the mana as the ability resolves, with the spell still on the stack.
func OldB607CasterAddsG(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
		return nil
	}
	return AddMana{Player: item.Trigger.Event.Actor, Produced: "{G}"}.Apply(NewContext(g, item))
}

// OldB607NontokenCreatureYouControlDied is Lightning Coils' first trigger: a
// nontoken creature its controller controlled died (this Coils is no creature,
// so "another" is never at issue).
func OldB607NontokenCreatureYouControlDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !ACreatureYouControlDied(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// OldB607CoilsHasFiveCharge is the intervening "if" of Lightning Coils'
// upkeep trigger: five or more charge counters on it.
func OldB607CoilsHasFiveCharge(source *game.Card) bool {
	return source != nil && source.Counters[game.CounterCharge] >= 5
}

// OldB607CoilsUpkeep is Lightning Coils' second trigger: "At the beginning of
// your upkeep, if this artifact has five or more charge counters on it, remove
// all of them from it and create that many 3/1 red Elemental creature tokens
// with haste. Exile them at the beginning of the next end step." The condition
// is asked as the upkeep begins and again as the ability resolves (CR 603.4).
func OldB607CoilsUpkeep(label string) game.TriggeredAbility {
	t := AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || !OnBattlefield(g, item.SourceCardID) || !OldB607CoilsHasFiveCharge(&c) {
			return nil
		}
		n := c.Counters[game.CounterCharge]
		ctx := NewContext(g, item)
		if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: -n}).Apply(ctx); err != nil {
			return err
		}
		return g.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(),
			Source:     ctx.Source(),
			Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("3/1 red Elemental with haste")), Count: n}},
		}, ctx.ItemTail(oldB607ExileTheElementals))
	})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return OldB607CoilsHasFiveCharge(source) && (when == nil || when(ev, source, lki, g))
	}
	return t
}

// oldB607ExileTheElementals is "exile them at the beginning of the next end
// step". An on-disk key.
var oldB607ExileTheElementals = SocB8ExileTokensAtNextEndStep("old-b607/lightning-coils-exile-the-elementals",
	"Lightning Coils — exile the Elementals")

// OldB607MyrIncubatorSearch is Myr Incubator's ability: search your library for
// any number of artifact cards, exile them, then create that many 1/1
// colorless Myr artifact creature tokens, then shuffle.
func OldB607MyrIncubatorSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return c.IsArtifact() },
		Dest:      game.ZoneExile,
		Unbounded: true,
		Shuffle:   true,
		Reason:    "Myr Incubator — search your library for any number of artifact cards",
		ThenTail:  ctx.ItemTail(oldB607MyrIncubatorTokens),
	}.Apply(ctx)
}

// oldB607MyrIncubatorTokens: one Myr for each card the search exiled. An
// on-disk key.
var oldB607MyrIncubatorTokens = CardTail("then/old-b607-myr-incubator", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 colorless Myr artifact"), N: len(o.IDs)}.Apply(ctx)
})

// OldB607ArtifactsYouControl is Hum of the Radix's amount: the artifacts the
// spell's caster controls, counted as the cost is worked out (CR 601.2f).
func OldB607ArtifactsYouControl(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return B03ArtifactsControlled(q.Game, q.Controller)
}

// OldB607TimesifterUpkeep is Timesifter's trigger: every player exiles the top
// card of their library (Triple Triad's chain), then the judge reads the cards.
func OldB607TimesifterUpkeep(g *game.Game, item *game.StackItem) error {
	var seats []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			seats = append(seats, p.ID)
		}
	}
	return oldB607TimesifterRound(g, item.Controller, item.SourceCardID, seats)
}

// oldB607TimesifterRound has each of `players`, in APNAP order, exile the top
// card of their library, then hands the cards to the judge.
func oldB607TimesifterRound(g *game.Game, controller, source uuid.UUID, players []uuid.UUID) error {
	judge := game.KeyedTail{Ref: oldB607TimesifterJudge, Player: controller, Source: source}
	step := game.KeyedTail{Ref: tripleTriadExileStep, Player: controller, Source: source, Next: &judge}
	return tripleTriadExileFrom(g, step, ApnapOrder(g, players), nil)
}

// oldB607TimesifterJudge: the player who exiled the card with the greatest mana
// value takes an extra turn after this one; players whose cards tie repeat the
// process until the tie is broken. A player with no card to exile has no card
// to compare (CR 701.17 has no card to exile), so cannot win; with no cards at
// all nobody takes the turn. An on-disk key.
var oldB607TimesifterJudge game.TailRef

func init() {
	oldB607TimesifterJudge = game.RegisterTail("then/old-b607-timesifter-judge", oldB607TimesifterJudged)
}

func oldB607TimesifterJudged(g *game.Game, o game.TailOutcome) error {
	best := -1
	var winners []uuid.UUID
	for _, id := range o.IDs {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		mv := c.ManaValue()
		switch {
		case mv > best:
			best, winners = mv, []uuid.UUID{c.Owner}
		case mv == best && !slices.Contains(winners, c.Owner):
			winners = append(winners, c.Owner)
		}
	}
	switch len(winners) {
	case 0:
		return nil
	case 1:
		g.TakeExtraTurnsForEffect(winners[0], o.Tail.Source, 1)
		return nil
	}
	return oldB607TimesifterRound(g, o.Tail.Player, o.Tail.Source, winners)
}
