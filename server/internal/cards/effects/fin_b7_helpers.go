package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b7_helpers.go — shared helpers for Final Fantasy batches 7 and 8
// (#4462).

// returnThisFromGraveyardTappedWithAFinalityCounter is "<cost>: Return this
// card from your graveyard to the battlefield tapped with a finality
// counter on it" (Relentless X-ATM092). It is
// returnThisFromGraveyardWithAFinalityCounter with the tapped entry: the
// card comes back tapped and with the counter as it enters, and a card that
// left the graveyard in response stays where it is.
func returnThisFromGraveyardTappedWithAFinalityCounter(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(item.SourceCardID, item.Controller, true,
		map[string]int{game.CounterFinality: 1}, game.KeyedTail{})
}

// nonGodCreature is "non-God creature".
func nonGodCreature(g *game.Game, caster uuid.UUID, c game.Card) bool {
	return Creature()(g, caster, c) && !HasSubtype("God")(g, caster, c)
}

// eachPlayerSacrificesHalfTheirNonGodCreatures is Zodiark's "each player
// sacrifices half the non-God creatures they control of their choice,
// rounded down": every seat's count is worked out first, from the board as
// the ability starts to resolve, then the seats choose in APNAP order and
// everything chosen is sacrificed together (CR 101.4, 608.2e).
func eachPlayerSacrificesHalfTheirNonGodCreatures(ctx *Context) error {
	var asks []uuid.UUID
	for _, seat := range tablePlayers(ctx) {
		if p := ctx.PlayerByID(seat); p == nil || p.Eliminated {
			continue
		}
		owed := len(permanentsControlledByMatching(ctx.Game, seat, nonGodCreature)) / 2
		for i := 0; i < owed; i++ {
			asks = append(asks, seat)
		}
	}
	if len(asks) == 0 {
		return nil
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), asks, sacrificeSpec("a non-God creature", nonGodCreature),
		"Zodiark — sacrifice half the non-God creatures you control")
	return nil
}

// youControlTheCreatureWithTheGreatestPower is "if you control the creature
// with the greatest power or tied for the greatest power": among every
// creature on the battlefield, the best power as the layers now stand
// (CR 613) is held by at least one creature `controller` controls.
func youControlTheCreatureWithTheGreatestPower(g *game.Game, controller uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	best, mine, any := 0, 0, false
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsCreature() {
			continue
		}
		p := c.CurrentPower()
		if !any || p > best {
			best, any = p, true
		}
		if c.Controller == controller && p > mine {
			mine = p
		}
	}
	return any && mine == best && controllerHasCreature(g, controller)
}

func controllerHasCreature(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() {
			return true
		}
	}
	return false
}

// tripleTriadExileStep is one step of Triple Triad's exile chain: each
// player's library in APNAP order is exiled from (no permission stamped),
// the step's Carry holding the libraries still to go and the cards exiled so
// far, and its Next, the grant, told every card once the last library is
// done. An on-disk key: never renamed, never reused.
var tripleTriadExileStep = game.RegisterTail("fin/triple-triad-exile-step", func(g *game.Game, o game.TailOutcome) error {
	lists := unpackIDLists(o.Tail.Carry)
	var remaining, landed []uuid.UUID
	if len(lists) > 0 {
		remaining = lists[0]
	}
	if len(lists) > 1 {
		landed = lists[1]
	}
	landed = append(append([]uuid.UUID(nil), landed...), o.IDs...)
	return tripleTriadExileFrom(g, o.Tail, remaining, landed)
})

func tripleTriadExileFrom(g *game.Game, step game.KeyedTail, remaining, landed []uuid.UUID) error {
	if len(remaining) == 0 {
		if step.Next == nil {
			return nil
		}
		return step.Next.Run(g, landed)
	}
	step.Carry = packIDLists(remaining[1:], landed)
	return g.ExileTopKeyedForEffect(remaining[0], remaining[0], 1, step)
}

// tripleTriadGrant is the second sentence once every exile has landed:
// "Until end of turn, you may play the card you own exiled this way and each
// other card exiled this way with lesser mana value than it without paying
// their mana costs." The controller's own card is playable; each other card
// is playable if its mana value (X is 0 in exile, CR 202.3e) is less than
// that card's. With no card of the controller's own among them, "it" is
// nothing and no other card is playable. Tail.Player is the controller.
var tripleTriadGrant = game.RegisterTail("fin/triple-triad-grant", func(g *game.Game, o game.TailOutcome) error {
	me := o.Tail.Player
	mine, have := uuid.Nil, false
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok && c.Owner == me {
			mine, have = id, true
			break
		}
	}
	if !have {
		return nil
	}
	own, _ := g.LookupCardForEffect(mine)
	limit := own.ManaValue()
	for _, id := range o.IDs {
		c, ok := g.LookupCardForEffect(id)
		if !ok || (id != mine && c.ManaValue() >= limit) {
			continue
		}
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player: me, Zone: game.ZoneExile, Cost: "{0}",
			SourceName: "Triple Triad",
		})
	}
	return nil
})

// tripleTriadUpkeep is Triple Triad's whole trigger: every player exiles the
// top card of their library, then the grant is worked out.
func tripleTriadUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var seats []uuid.UUID
	for _, p := range g.Seats {
		if p != nil {
			seats = append(seats, p.ID)
		}
	}
	grant := game.KeyedTail{Ref: tripleTriadGrant, Player: item.Controller, Source: item.SourceCardID}
	step := game.KeyedTail{Ref: tripleTriadExileStep, Player: item.Controller, Source: item.SourceCardID, Next: &grant}
	return tripleTriadExileFrom(ctx.Game, step, apnapOrder(ctx.Game, seats), nil)
}

// vanilleMilled is Vanille's "then": once the mill of two has settled, a
// permanent card from the controller's graveyard goes to their hand. The
// return is not targeted and not optional, so with a permanent card there
// one of them comes back (a lone candidate without a question, two or more
// ask for a pick), the two cards just milled included (CR 608.2c). An
// on-disk key: never renamed, never reused.
var vanilleMilled = CardTail("vanille-cheerful-lcie/after-mill", func(ctx *Context, _ game.TailOutcome) error {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if Permanent()(ctx.Game, ctx.Controller(), c) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 1 {
		return ReturnFromGraveyard{Target: candidates[0], Dest: game.ZoneHand}.Apply(ctx)
	}
	return takeOneIntoYourHand(ctx, candidates,
		"Vanille, Cheerful l'Cie — return a permanent card from your graveyard to your hand", 1)
})

// condYouOwnAndControlThisAndANamed is YouOwnAndControlThisAndANamed as an
// intervening "if" (CR 603.4): checked as the meld ability triggers and
// again as it resolves, before the "you may pay" is asked. It declares no
// footprint.
type condYouOwnAndControlThisAndANamed struct{ partner, partnerType string }

func (c condYouOwnAndControlThisAndANamed) Holds(ctx *Context) bool {
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return false
	}
	return YouOwnAndControlThisAndANamed(c.partner, c.partnerType)(ctx.Game, &src)
}

func (condYouOwnAndControlThisAndANamed) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// vanilleMeldPaid is the paid half of Vanille's meld ability: exile them,
// then meld them into Ragnarok, Divine Deliverance.
var vanilleMeldPaid = PayUnlessBranches("pay-unless/vanille-cheerful-lcie-meld",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return MeldWith{Partner: "Fang, Fearless l'Cie", PartnerType: "creature"}.Apply(ctx)
	}, nil)

// equipmentAttachedToCreatureControlledBy is "it's attached to a creature you
// control": the Equipment is on the battlefield, attached to a creature that
// is on the battlefield under `player`'s control.
func equipmentAttachedToCreatureControlledBy(g *game.Game, equipment, player uuid.UUID) bool {
	eq, ok := g.LookupCardForEffect(equipment)
	if !ok || !eq.IsAttached() || !onBattlefield(g, equipment) {
		return false
	}
	host, ok := g.LookupCardForEffect(eq.AttachedTo.ID)
	return ok && host.IsCreature() && host.Controller == player && onBattlefield(g, host.InstanceID)
}

// terraMagicalAdeptTake is Terra's "put up to one enchantment card milled
// this way into your hand", the pick over the cards that actually reached
// the graveyard (CR 400.7), up to one, so declining is an answer. An on-disk
// key: never renamed, never reused.
var terraMagicalAdeptTakePick = askedPick("choose-cards/terra-magical-adept-take-pick", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})

var terraMagicalAdeptMilled = CardTail("terra-magical-adept/after-mill", func(ctx *Context, o game.TailOutcome) error {
	var eligible []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && Enchantment()(ctx.Game, ctx.Controller(), c) {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Item.Controller,
		Source:   ctx.Item.SourceCardID,
		Question: "Terra, Magical Adept — put up to one enchantment card milled this way into your hand",
		Cards:    eligible,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  terraMagicalAdeptTakePick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// esperTerraCopy is Esper Terra's chapters I–III, "Create a token that's a
// copy of target nonlegendary enchantment you control. It gains haste. If
// it's a Saga, put up to three lore counters on it. Sacrifice it at the
// beginning of your next end step."
func esperTerraCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return CreateTokenCopiesTogether{
			Controller: item.Controller,
			Copies:     []uuid.UUID{t.ID},
			Then:       ctx.ItemTail(esperTerraCopied),
		}.Apply(ctx)
	}
	return nil
}

// esperTerraCopied is the rest of the sentence run over the token that was
// made: it gains haste (no duration, CR 611.2a), a Saga is offered up to
// three lore counters, and it is sacrificed at the beginning of its
// creator's next end step. An on-disk key: never renamed, never reused.
var esperTerraCopied = CardTail("esper-terra/copied", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	token := o.IDs[0]
	if err := (ScopedEffectFor{
		Target:   token,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: game.IndefiniteDuration(),
		Label:    "Esper Terra — the token gains haste",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (ScheduleDelayedTrigger{
		Label:              "Esper Terra — sacrifice the token",
		ControllerTurnOnly: true,
		Cards:              []uuid.UUID{token},
		Body:               sacrificeListedCardsBody,
	}).Apply(ctx); err != nil {
		return err
	}
	c, ok := ctx.Game.LookupCardForEffect(token)
	if !ok || !c.HasSubtype("Saga") {
		return nil
	}
	return ChooseNumber{
		Min: 0, Max: 3,
		Question: "Esper Terra — put up to three lore counters on the token",
		Then:     esperTerraLoreCounters,
		Carry:    []uuid.UUID{token},
	}.Apply(ctx)
})

// esperTerraLoreCounters puts the chosen number of lore counters on the
// token, still on the battlefield as the number is answered.
var esperTerraLoreCounters = NumberThen("esper-terra/lore-counters", func(ctx *Context, r game.NumberAnswer) error {
	if r.Number <= 0 || len(r.Carry) == 0 || !onBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return AddCounter{Target: r.Carry[0], Kind: game.CounterLore, N: r.Number}.Apply(ctx)
})

// esperTerraManaAndReturn is chapter IV: add {W}{W}, {U}{U}, {B}{B}, {R}{R},
// and {G}{G}, then exile Esper Terra and return it front face up.
func esperTerraManaAndReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddMana{Produced: "{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}"}).Apply(ctx); err != nil {
		return err
	}
	return Flicker{Target: item.SourceCardID, Controller: item.Controller}.Apply(ctx)
}

// phoenixFlamesOfRebirth is Phoenix's chapter III: the target creature cards
// still legal go onto the battlefield together under the controller's
// control, then the Saga is exiled and returned front face up.
func phoenixFlamesOfRebirth(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) > 0 {
		if err := putCardsFromZoneTogether(ctx, game.ZoneGraveyard, ids, item.Controller, false); err != nil {
			return err
		}
	}
	return Flicker{Target: item.SourceCardID, Controller: item.Controller}.Apply(ctx)
}

// plusAnAdditionalTokenReplacement is "If one or more tokens would be created
// under your control, those tokens plus <a token> are created instead"
// (Peregrin Took's Food, Quina's Frog): an APPEND to the creation's groups,
// once per creation instruction (CR 701.7b), not a multiply or a rewrite of
// what the instruction makes. It composes with the other creation
// replacements on the same event as the rules want. The extra token is made
// as the original ones are: what the instruction says about them (tapped, a
// sacrifice later) is not applied to it.
func plusAnAdditionalTokenReplacement(label string, token func() game.Card) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventTokenCreated},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventCreateTokens && ev.TokenCount() > 0 && ev.TokenController == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{Template: ruleDefinedToken(token()), Count: 1})
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// offerAnyGraveyardCastWithAnyMana is "you may cast target instant or
// sorcery card from a graveyard, and mana of any type can be spent to cast
// that spell. If that spell would be put into a graveyard, exile it instead"
// (Quistis Trepe): OfferGraveyardCast over a card in ANY player's graveyard,
// with AnyType on the permission so any mana pays for it, colourless
// included (CR 609.4b). The cast is made during the resolution (CR 608.2g),
// at the card's mana cost, flash-timed because the card's own timing is
// ignored, and the holder's next priority pass closes the offer with the
// card staying in its graveyard.
func offerAnyGraveyardCastWithAnyMana(ctx *Context, target uuid.UUID, sourceName string) error {
	if target == uuid.Nil {
		return nil
	}
	g := ctx.Game
	z := g.FindCardZoneForEffect(target)
	c, ok := g.LookupCardForEffect(target)
	if !ok || z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:                  ctx.Controller(),
		Zone:                    game.ZoneGraveyard,
		Duration:                g.UntilEndOfTurnDuration(),
		AltCostKey:              game.AltCostKeyGrantedCost,
		Timing:                  game.TimingFlash,
		CastOnly:                true,
		AnyColor:                true,
		AnyType:                 true,
		ExileInsteadOfGraveyard: true,
		LapseOnPass:             game.LapseStaysInHand,
		Source:                  ctx.Source(),
		SourceName:              sourceName,
		Label:                   sourceName + " — cast " + c.Name + " from a graveyard",
	}, []game.Card{c})
	offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), sourceName, game.MayCastKeywordPaidCast, declineLeaveInGraveyard, c.InstanceID)
	return nil
}

// anotherCreatureEvent is the condition "another creature" for an event
// that names a card (Zodiark's EventSacrifice, Rampaging Ferocidon's
// EventETB): the event's card is a creature other than the source.
func anotherCreatureEvent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}
