package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch23_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 23 (#385, `edhrec_rank` 2436–2535). Own file per
// the #231 convention; every package-level name carries the b23
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "another creature you control enters" is
// b13AnotherCreatureYouControlEntered, "a permanent entered under
// your control" is enteredUnderYourControl, "an opponent cast a
// spell" is b15OpponentCastSpell, "this creature died" is cardDied,
// the X-counter-on-resolve shape is Goldvein Hydra's, the +1/+1
// doubling body is b08DoubleCountersOnEachCreatureYouControl, the
// counter LKI read is b13LastKnownCounters, the mass-damage body is
// damageEachMatching, the two-basics fetch is Explosive Vegetation's
// SearchLibrary, and the tribal keyword grant is TribalKeywordGrant.

// --- token templates ---------------------------------------------

// --- trigger conditions ------------------------------------------

// b23ArtifactPutIntoGraveyardFromBattlefield is Disciple of the
// Vault's condition: an artifact — anyone's, token or card, the
// source's controller's own included, as printed — was put into a
// graveyard from the battlefield. The card is read post-move; its
// printed type line survives the move, and a token is still
// findable until the state-based sweep removes it (the same read
// Agent of the Iron Throne makes for a Treasure).
func b23ArtifactPutIntoGraveyardFromBattlefield(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsArtifact()
}

// b23AnotherGreenCreatureYouControlEntered is Ivy Lane Denizen's
// condition: another creature entered under the source's
// controller's control, and it is green. Colours are the effective
// ones, so a creature something else painted green counts.
func b23AnotherGreenCreatureYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	if !b13AnotherCreatureYouControlEntered(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasColor("G")
}

// b23DragonYouControlEntered is Encroaching Dragonstorm's second
// trigger: a Dragon entered under the source's controller's control.
// Effective subtypes, so a changeling counts; the enchantment itself
// is never a Dragon, so no "another" guard is needed.
func b23DragonYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSubtype("Dragon")
}

// b23NoCreaturesOnBattlefield is Pyrohemia's intervening-if: no
// creature — anyone's — is on the battlefield. Post-layer type, so
// an animated land keeps the enchantment alive.
func b23NoCreaturesOnBattlefield(g *game.Game) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			return false
		}
	}
	return true
}

// b23IsYourTurn reports whether the source's controller is the
// active player — "during your turn".
func b23IsYourTurn(g *game.Game, controller uuid.UUID) bool {
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return false
	}
	active := g.Seats[g.Turn.ActiveSeat]
	return active != nil && active.ID == controller
}

// b23ElfYouControlBecameTappedFirstTimeThisTurn is the condition of
// the ability Dionus, Elvish Archdruid grants — "whenever this
// creature becomes tapped during your turn … This ability triggers
// only once each turn" — hung on Dionus itself and read against
// every Elf its controller controls (Dionus included; he is an Elf).
//
// Two event kinds feed it, for the reason b11DwarfYouControlBecameTapped
// gives: the engine taps an attacker without an EventTapCard, so an
// EventAttack whose creature is now tapped is "became tapped" and a
// vigilance Elf is not.
//
// "Only once each turn" is per Elf — each Elf carries its own copy
// of the granted ability. The engine's trigger tally
// (b11TriggeredThisTurn) is keyed by source and label, which cannot
// tell two Elf Warrior tokens apart, so the tally here is the event
// log itself: the ability fires on the FIRST tap or tapped-attack
// event for this Elf this turn and on no later one. Every earlier
// tap of the Elf this turn is walked, whether or not it triggered
// the ability then, which errs weaker — an Elf tapped before Dionus
// arrived does not fire later that turn — and never stronger: a
// trigger is impossible on any but the first event.
func b23ElfYouControlBecameTappedFirstTimeThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	switch ev.Kind {
	case game.EventTapCard, game.EventAttack:
	default:
		return false
	}
	if !b23IsYourTurn(g, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller || !c.IsCreature() || !c.HasSubtype("Elf") {
		return false
	}
	if ev.Kind == game.EventAttack && !c.Tapped {
		return false
	}
	return !b23TappedEarlierThisTurn(g, ev)
}

// b23TappedEarlierThisTurn walks the event log back to the start of
// the current turn looking for a tap or attack event naming the same
// card as `ev`, ignoring `ev` itself and anything after it. The turn
// boundary is the first EventStepBegan carrying an earlier turn
// number: the cursor stamps one on every step it enters, so every
// event above that mark belongs to this turn.
func b23TappedEarlierThisTurn(g *game.Game, ev game.Event) bool {
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if (prev.Kind == game.EventTapCard || prev.Kind == game.EventAttack) && prev.CardID == ev.CardID {
			return true
		}
	}
	return false
}

// --- effect bodies -----------------------------------------------

// b23DoublePlusOneCountersOn puts as many +1/+1 counters on `target`
// as it already has — Primordial Hydra's upkeep, and the second half
// of Fangs of Kalonia per creature. Placed through AddCounter, so a
// counter doubler applies to the doubling, as printed. A target that
// has left the battlefield, or has no +1/+1 counters, is left alone.
func b23DoublePlusOneCountersOn(ctx *Context, target uuid.UUID) error {
	if z := ctx.Game.FindCardZoneForEffect(target); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	n := c.Counters["+1/+1"]
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: target, Kind: "+1/+1", N: n}.Apply(ctx)
}

// b23GrowThenDoubleEach is Fangs of Kalonia's body for a set of
// creatures: a +1/+1 counter on each, THEN the +1/+1 counters on
// each creature that actually received one are doubled. The two
// passes are separate, as the printed "then" says, so every first
// counter — and whatever a doubler did to it — is on the board
// before any doubling is read. "Had a +1/+1 counter put on it this
// way" is measured, not assumed: a creature whose count did not
// rise (it left the battlefield, or something prevented the
// counter) is skipped by the second pass.
func b23GrowThenDoubleEach(ctx *Context, ids []uuid.UUID) error {
	var grew []uuid.UUID
	for _, id := range ids {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		before := 0
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			before = c.Counters["+1/+1"]
		}
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
			return err
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.Counters["+1/+1"] > before {
			grew = append(grew, id)
		}
	}
	for _, id := range grew {
		if err := b23DoublePlusOneCountersOn(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// b23CreaturesYouControl lists the controller's creatures, the set
// snapshotted before anything is placed (CR 608.2).
func b23CreaturesYouControl(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// b23DamageEachCreatureAndEachPlayer is Pyrohemia's activation: 1
// damage from the source to every creature and every seated player,
// the controller included. Creatures first, then players, each set
// snapshotted before the first point lands; a creature that took
// lethal dies at the state-based sweep, so the end-step sacrifice
// clause sees an empty board only after the last activation.
func b23DamageEachCreatureAndEachPlayer(ctx *Context, n int) error {
	if err := damageEachMatching(ctx, Creature(), n); err != nil {
		return err
	}
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := (DealDamage{Source: ctx.Source(), Target: p.ID, Amount: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b23ChargeThenDrawPerCharge is Insight Engine's activation: a
// charge counter on the source, then a card for each charge counter
// on it. An Engine that left the battlefield in response cannot take
// the counter, and the draw reads its last-known count (CR 113.7a),
// which the counter LKI on the event log supplies.
func b23ChargeThenDrawPerCharge(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := item.SourceCardID
	charges := 0
	if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneBattlefield {
		if err := (AddCounter{Target: id, Kind: "charge", N: 1}).Apply(ctx); err != nil {
			return err
		}
		if c, ok := g.LookupCardForEffect(id); ok {
			charges = c.Counters["charge"]
		}
	} else {
		charges = b13LastKnownCounters(g, id, "charge")
	}
	return DrawCards{Player: item.Controller, N: charges}.Apply(ctx)
}

// b23TotalManaValueAtMost is Protean Hulk's set constraint: the
// picked creature cards' mana values, printed and with X as zero
// (CR 202.3e), sum to `limit` or less.
func b23TotalManaValueAtMost(limit int) func([]game.Card) bool {
	return func(cards []game.Card) bool {
		total := 0
		for _, c := range cards {
			total += c.ManaValue()
		}
		return total <= limit
	}
}

// b23SearchCreaturesWithTotalManaValue is Protean Hulk's body: the
// controller searches for any number of creature cards with total
// mana value `limit` or less and puts them onto the battlefield,
// then shuffles. "Any number" is a limit of the whole library, so the
// chooser sees every creature card and takes as many as fit; the
// Validate clause is checked on the picked set, and an empty pick is
// the ordinary "fail to find". Each card enters through the search
// path, so its own enters-tapped clause and every ETB trigger fire.
func b23SearchCreaturesWithTotalManaValue(g *game.Game, item *game.StackItem, limit int, reason string) error {
	ctx := NewContext(g, item)
	p := ctx.PlayerByID(item.Controller)
	if p == nil || p.Library == nil {
		return nil
	}
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return c.IsCreature() },
		Dest:      game.ZoneBattlefield,
		Limit:     p.Library.Size(),
		Shuffle:   true,
		Validate:  b23TotalManaValueAtMost(limit),
		Reason:    reason,
	}.Apply(ctx)
}

// b23UntapAndGrow is the granted ability's body: untap the Elf and
// put a +1/+1 counter on it, if it is still on the battlefield.
func b23UntapAndGrow(g *game.Game, item *game.StackItem, elf uuid.UUID) error {
	if !onBattlefield(g, elf) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (UntapTarget{Target: elf}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: elf, Kind: "+1/+1", N: 1}.Apply(ctx)
}
