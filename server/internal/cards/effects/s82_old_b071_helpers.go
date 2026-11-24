package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b071_helpers.go — shared bodies for Kaldheim batch 071.

// OldB071TheListedObjectDealtDamageCondition is the event condition of "whenever
// that creature deals damage this turn" (Niko Aris): an EventDealDamage of
// positive damage whose source is the listed object. The trigger's
// duration pins the object, so a flicker ends it (CR 400.7).
var OldB071TheListedObjectDealtDamageCondition = game.DelayedCondition("damage/old-b071-the-listed-object-dealt", oldB071TheListedObjectDealtDamage)

func oldB071TheListedObjectDealtDamage(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return ev.Kind == game.EventDealDamage && ev.Amount > 0 && len(dt.Cards) > 0 && ev.Source == dt.Cards[0]
}

// OldB071ReturnTheListedObjectToHandBody is "return it to its owner's hand":
// each listed card that is still on the battlefield goes back (CR 603.7c).
var OldB071ReturnTheListedObjectToHandBody = game.SimpleDelayedBody("damage/old-b071-return-the-listed-object", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && OnBattlefield(g, t.ID) {
			if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})

// --- Kaya the Inexorable ------------------------------------------------------

// OldB071KayaEmblemName heads the emblem's prompts and labels.
const OldB071KayaEmblemName = "Kaya the Inexorable emblem"

// OldB071KayaEmblemUpkeep is the emblem's "you may cast a legendary spell from
// your hand, from your graveyard, or from among cards you own in exile without
// paying its mana cost": the controller picks at most one legendary card of the
// three places (a card a legendary spell can be cast from: a land is never one,
// and a face-down exiled card can't be looked at), then casts it free during the
// resolution (CR 608.2g). Only the picked card's permission is stamped, so one
// spell is cast, not one for each place.
func OldB071KayaEmblemUpkeep(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	clause := FreeCastSpell{Filter: game.PermissionFilter{LegendaryOnly: true}}
	probe := clause.On(game.CastPermission{})
	var candidates []uuid.UUID
	add := func(c game.Card) {
		if probe.OffersASpell(c) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if p.Hand != nil {
		for _, c := range p.Hand.Cards {
			add(c)
		}
	}
	if p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			add(c)
		}
	}
	for _, c := range g.Exile.Cards {
		if c.Owner == player && !c.FaceDown {
			add(c)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: OldB071KayaEmblemName + " — you may cast a legendary spell from your hand, your graveyard, or among cards you own in exile without paying its mana cost",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		ThenKey:  OldB071KayaEmblemCast,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// OldB071KayaEmblemCast is the pick's continuation: a free, any-timing,
// cast-only permission over the one card, in the zone it is in, offered during
// the resolution. An on-disk key: never renamed, never reused.
var OldB071KayaEmblemCast = AskedPick("choose-cards/old-b071-kaya-emblem-cast", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	g := ctx.Game
	id := r.Picked[0]
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	var perm game.CastPermission
	switch z.Kind {
	case game.ZoneHand:
		perm = g.FreeCastFromHandPermission(r.Chooser, r.Source, OldB071KayaEmblemName)
	case game.ZoneGraveyard, game.ZoneExile:
		perm = game.CastPermission{
			Player:      r.Chooser,
			Zone:        z.Kind,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			AltCostKey:  game.AltCostKeyGrantedCost,
			Timing:      game.TimingFlash,
			CastOnly:    true,
			LapseOnPass: game.LapseStaysInHand,
			Source:      r.Source,
			SourceName:  OldB071KayaEmblemName,
			Label:       OldB071KayaEmblemName + " — cast it without paying its mana cost",
		}
	default:
		return nil
	}
	perm = FreeCastSpell{Filter: game.PermissionFilter{LegendaryOnly: true}}.On(perm)
	decline := "Leave it where it is"
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      r.Chooser,
		Source:       r.Source,
		Reason:       OldB071KayaEmblemName,
		Cards:        []uuid.UUID{id},
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: decline,
	}) {
		return nil
	}
	if c, ok := g.LookupCardForEffect(id); ok {
		g.GrantCastPermissionToCardsForEffect(perm, []game.Card{c})
	}
	return nil
})

// OldB071KayaGhostformBundle is the bundle key of the trigger the +1 grants.
const OldB071KayaGhostformBundle = "kaya-the-inexorable/ghostform-return"

// OldB071GhostformReturn is the granted trigger's body: "return it to its
// owner's hand and create a 1/1 white Spirit creature token with flying". The
// card returns only if it is still the object that left (CR 400.7) and sits in
// a graveyard or in exile; the token is made regardless, a separate
// instruction (CR 608.2c).
func OldB071GhostformReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if z := g.FindCardZoneForEffect(item.SourceCardID); z != nil && (z.Kind == game.ZoneGraveyard || z.Kind == game.ZoneExile) && SameGraveyardObject(z, item) {
		if err := (BounceToHand{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
	}
	return CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(ctx)
}

// --- Valki, God of Lies // Tibalt, Cosmic Impostor ---------------------------

// OldB071ValkiExileKey is the label (the trigger's Key) of Valki's exiling
// ability: the row the "exiled with Valki" link names (CR 607.2a).
const OldB071ValkiExileKey = "Valki, God of Lies — each opponent reveals their hand; for each opponent, exile a creature card they revealed this way until Valki leaves the battlefield"

// OldB071ValkiEnters is "each opponent reveals their hand. For each opponent,
// exile a creature card they revealed this way until Valki leaves the
// battlefield.": every opponent's hand is revealed to the table, then the
// controller picks one creature card from each hand that has one (a hand with
// none exiles nothing), one prompt for each opponent in turn order. The
// exile is each pick's own, linked to Valki (ExileUntil.Linked) and ended with
// Valki's object; a Valki that has gone by the time a pick is answered exiles
// nothing (CR 610.3b).
func OldB071ValkiEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Hand == nil {
			continue
		}
		g.RevealHandForEffect(opp)
		var creatures []uuid.UUID
		for _, c := range p.Hand.Cards {
			if c.IsCreature() {
				creatures = append(creatures, c.InstanceID)
			}
		}
		if len(creatures) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    item.Controller,
			FromPlayer: opp,
			Source:     item.SourceCardID,
			Question:   "Valki, God of Lies — exile a creature card from this revealed hand until Valki leaves the battlefield",
			Cards:      creatures,
			Min:        1,
			Max:        1,
			LiveZone:   game.ZoneHand,
			ThenKey:    OldB071ValkiExile,
			Asked:      game.NewAskedItem(item),
		})
	}
	return nil
}

// OldB071ValkiExile exiles the picked creature card until Valki leaves,
// linked to him. An on-disk key: never renamed, never reused.
var OldB071ValkiExile = AskedPick("choose-cards/old-b071-valki-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	return ExileUntil{
		Target:     r.Picked[0],
		ThisLeaves: true,
		Linked:     true,
		Label:      "Valki, God of Lies — the exiled card returns when Valki leaves the battlefield",
	}.Apply(ctx)
})

// OldB071ValkiCopy is "{X}: Choose a creature card exiled with Valki with mana
// value X. Valki becomes a copy of that card.": Assimilation Aegis's linked
// copy over the cards Valki's exiling ability holds, narrowed to creature cards
// of mana value X, with no stated duration (CR 611.2a).
func OldB071ValkiCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := ctx.X()
	return BecomeCopyOfExiledWith{
		Target:   item.SourceCardID,
		Row:      OldB071ValkiExileKey,
		Match:    func(c game.Card) bool { return c.IsCreature() && c.ManaValue() == x },
		Duration: CopyIndefinite,
		Question: "Valki, God of Lies — choose a creature card exiled with Valki; Valki becomes a copy of it",
		Label:    "Valki, God of Lies",
	}.Apply(ctx)
}

// OldB071TibaltName heads Tibalt's permission and its labels.
const OldB071TibaltName = "Tibalt, Cosmic Impostor"

// OldB071TibaltPermission is "you may play cards exiled with Tibalt, and you may
// spend mana as though it were mana of any color to cast those spells" for one
// exiled card: a stored permission of the emblem's owner (the player Tibalt
// recorded as he entered), for as long as the card stays in exile.
func OldB071TibaltPermission(owner, tibalt uuid.UUID) game.CastPermission {
	return game.CastPermission{
		Player:     owner,
		Zone:       game.ZoneExile,
		Duration:   game.WhileInZoneDuration(),
		AnyColor:   true,
		Source:     tibalt,
		SourceName: OldB071TibaltName,
		Label:      OldB071TibaltName + " — play it, spending mana as though it were mana of any color",
	}
}

// OldB071TibaltOwner is the player who got Tibalt's emblem: the controller as
// he entered, recorded then, falling back to his current controller.
func OldB071TibaltOwner(g *game.Game, item *game.StackItem) uuid.UUID {
	if owner := g.ChosenPlayerOf(item.SourceCardID); owner != uuid.Nil {
		return owner
	}
	return item.Controller
}

// OldB071TibaltGranted is the continuation of every exile Tibalt makes: the
// cards that reached exile (told as the tail's IDs) get the emblem's play
// permission, one stored permission over the set (CR 611.2c). An on-disk key:
// never renamed, never reused.
var OldB071TibaltGranted = game.RegisterTail("exile-then/old-b071-tibalt-grant", func(g *game.Game, o game.TailOutcome) error {
	if o.Tail.Permission == nil || len(o.IDs) == 0 {
		return nil
	}
	var cards []game.Card
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok {
			if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
				cards = append(cards, c)
			}
		}
	}
	g.GrantCastPermissionToCardsForEffect(*o.Tail.Permission, cards)
	if next := o.Tail.Next; next != nil {
		return next.Run(g, o.IDs)
	}
	return nil
})

// OldB071TibaltRestOfTheAbility is the −8's "Add {R}{R}{R}", run once the
// graveyards have been exiled. An on-disk key.
var OldB071TibaltAddRRR = CardTail("exile-then/old-b071-tibalt-add-rrr", func(ctx *Context, _ game.TailOutcome) error {
	return AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
})

// --- Cosima, God of the Voyage // The Omenkeel -----------------------------------

// OldB071CosimaUpkeepKey is the label of Cosima's upkeep ability: the row her
// own exile is "exiled with" (CR 607.2a), which is what makes the card in exile
// the one that gained the second ability.
const OldB071CosimaUpkeepKey = "Cosima, God of the Voyage — you may exile Cosima; if you do, it gains the voyage ability"

// OldB071CosimaIsGainedExile reports whether this card in exile is a Cosima
// that her own upkeep ability exiled: linked to her old object by that very
// row, and still the object that was exiled (the link is cleared as the card
// leaves exile, CR 400.7).
func OldB071CosimaIsGainedExile(c game.Card, g *game.Game) bool {
	if c.ExiledWith.ID != c.InstanceID || c.ExiledWithAbility.Name != OldB071CosimaUpkeepKey {
		return false
	}
	z := g.FindCardZoneForEffect(c.InstanceID)
	return z != nil && z.Kind == game.ZoneExile
}

const (
	oldB071CosimaCounterOption = "Put a voyage counter on Cosima"
	oldB071CosimaReturnOption  = "Don't: return Cosima to the battlefield with X +1/+1 counters and draw X cards"
)

// OldB071CosimaLandEntered is the exiled Cosima's trigger body: "you may put a
// voyage counter on it. If you don't, return Cosima to the battlefield with X
// +1/+1 counters on it and draw X cards", asked as one two-way question.
func OldB071CosimaLandEntered(g *game.Game, item *game.StackItem) error {
	g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Cosima, God of the Voyage — a land entered: put a voyage counter on Cosima, or return her with X +1/+1 counters and draw X cards?",
		Options:  []game.ChoiceOption{{Label: oldB071CosimaCounterOption}, {Label: oldB071CosimaReturnOption}},
		ThenKey:  OldB071CosimaChoice,
		Carry:    []uuid.UUID{item.SourceCardID},
	})
	return nil
}

// OldB071CosimaChoice answers that question for the card it names, if she is
// still in exile with the ability. An on-disk key: never renamed, never reused.
var OldB071CosimaChoice = game.RegisterOptionPickThen("option-pick/old-b071-cosima-voyage", func(g *game.Game, r game.OptionPicked) error {
	if len(r.Carry) != 1 || r.Option == nil {
		return nil
	}
	id := r.Carry[0]
	c, ok := g.LookupCardForEffect(id)
	if !ok || !OldB071CosimaIsGainedExile(c, g) {
		return nil
	}
	if r.Option.Label == oldB071CosimaCounterOption {
		// A counter on a card in exile (CR 122.1): the permanent-only counter
		// pipeline (Doubling Season and the rest read permanents) does not
		// reach it, so it is written on the exiled card itself.
		for i := range g.Exile.Cards {
			if g.Exile.Cards[i].InstanceID == id {
				if g.Exile.Cards[i].Counters == nil {
					g.Exile.Cards[i].Counters = map[string]int{}
				}
				g.Exile.Cards[i].Counters["voyage"]++
			}
		}
		return nil
	}
	x := c.Counters["voyage"]
	var counters map[string]int
	if x > 0 {
		counters = map[string]int{game.CounterPlusOne: x}
	}
	return g.ReturnFromExileWithCountersKeyedForEffect(id, c.Owner, false, counters,
		game.KeyedTail{Ref: OldB071CosimaDraw, Player: c.Owner, Source: id, N: x})
})

// OldB071CosimaDraw is "and draw X cards", told the card has returned. An
// on-disk key.
var OldB071CosimaDraw = game.RegisterTail("then/old-b071-cosima-draw", func(g *game.Game, o game.TailOutcome) error {
	if o.Tail.N <= 0 || len(o.IDs) == 0 {
		return nil
	}
	ctx := NewContext(g, &game.StackItem{Kind: game.StackItemTriggered, Controller: o.Tail.Player, Owner: o.Tail.Player, SourceCardID: o.Tail.Source})
	return DrawCards{Player: o.Tail.Player, N: o.Tail.N}.Apply(ctx)
})

// OldB071OmenkeelTriggered is "whenever a Vehicle you control deals combat
// damage to a player": a damage event of a Vehicle the source's controller
// controls to a player, read off the dealer as it is now.
func OldB071OmenkeelTriggered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !CombatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.HasSubtype(source.TextSubtype("Vehicle"))
}

// OldB071OmenkeelExile is "that player exiles that many cards from the top of
// their library. You may play lands from among those cards for as long as they
// remain exiled": the damaged player's top cards go to exile and the Omenkeel's
// controller gets a play permission over the lands among them, lasting while
// each stays exiled.
func OldB071OmenkeelExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
		return nil
	}
	ev := item.Trigger.Event
	perm := game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneExile,
		Duration:   game.WhileInZoneDuration(),
		Filter:     game.PermissionFilter{LandsOnly: true},
		Source:     item.SourceCardID,
		SourceName: "The Omenkeel",
		Label:      "The Omenkeel — play lands from among the cards exiled this way",
	}
	return g.ExileTopWithPermissionKeyedForEffect(ev.Target, item.Controller, ev.Amount, perm, game.KeyedTail{})
}

// OldB071FrostAugurTake is Frost Augur's "you may reveal it and put it into
// your hand": the one card looked at is revealed and taken, and a card that
// isn't taken stays where it is, on top (a clause with no "rest" to move). An
// on-disk key: never renamed, never reused.
var OldB071FrostAugurTake = RegisterTake("choose-cards/old-b071-frost-augur-take",
	TakeDef{Reveal: true, Then: func(*game.Game, TakeFromLibraryResult) error { return nil }})
