package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch13_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 13 (#306, `edhrec_rank` 1421–1524). Own file per
// the #231 convention; every package-level name carries the b13
// prefix.
//
// What is NOT here, because main already had it: "whenever you cast
// a noncreature spell" is b10NoncreatureSpellCastByYou, "N damage to
// each opponent" is damageToEachOpponent, "each opponent loses N
// life" is eachOpponentLosesLife, "another permanent you control
// entered" is enteredUnderYourControl, "this permanent enters" is
// b06SelfETB, the fetch-a-basic-tapped body is fetchBasicTapped, the
// return-all-lands body is b10ReturnAllLandCardsFromGraveyardTapped,
// the wheel's discard is discardWholeHand, the spell-side mana value
// is game.(*Game).ManaValueForEffect, "nonbasic land" is b10NonbasicLand, and the
// per-ability "one or more" dedup is OncePerBatch.

// --- token templates ---------------------------------------------

// --- trigger conditions ------------------------------------------

// b13AnotherCreatureYouControlEntered is "whenever another creature
// you control enters" — Corpse Knight, Witty Roastmaster, Prosperous
// Innkeeper. Post-layer type, so an animated land counts.
func b13AnotherCreatureYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsCreature()
}

// b13LandYouControlEntered is landfall — Hedron Crab. Tireless
// Provisioner's condition, named.
func b13LandYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsLand()
}

// b13CreatureDealtDamageToYou is No Mercy's condition: a creature —
// anyone's, the controller's own included, as printed — dealt damage
// to the source's controller, combat or not. The creature is looked
// up live; damage is dealt before SBAs run, so an attacker that
// traded with its blocker is still on the battlefield when its event
// fires. The damaging creature's ID is ev.Source itself, so callers
// that need it read that off the event rather than a second return.
func b13CreatureDealtDamageToYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target != source.Controller {
		return false
	}
	return damageSourceIsABattlefieldCreature(ev, g)
}

// damageSourceIsABattlefieldCreature reports whether an EventDealDamage's
// source is a creature on the battlefield — "whenever a creature deals
// damage …" (No Mercy, Liliana's Talent). Looked up live, which is right
// at harvest: damage is dealt before SBAs run.
func damageSourceIsABattlefieldCreature(ev game.Event, g *game.Game) bool {
	if z := g.FindCardZoneForEffect(ev.Source); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsCreature()
}

// destroyTheCreatureThatDealtTheDamage is "destroy it" / "destroy that
// creature" after a creature's damage (No Mercy, Liliana's Talent). The
// creature rides the trigger's event and is destroyed only if it is
// still on the battlefield. Not targeted, so hexproof does not save it;
// indestructible does.
func destroyTheCreatureThatDealtTheDamage(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.Source
	if z := g.FindCardZoneForEffect(attacker); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return DestroyTarget{Target: attacker}.Apply(NewContext(g, item))
}

// b13OtherCreatureYouControlLeftWithoutDying is Dour Port-Mage's
// condition: another creature the source's controller controlled
// left the battlefield for anywhere but a graveyard — a bounce, an
// exile, a library tuck. EventLTB carries the destination in
// NewZone, so "without dying" is the one comparison CR 700.4 makes.
//
// "You control" is the controller the creature had as it left
// (leftUnderControlOf, #1682), so a stolen creature bounced to its
// owner's hand counts for the thief, under whose control it left, and
// not for the owner. "Creature" is its type as it last existed
// (leftAsType, #1675), so a crewed Vehicle or an animated land that
// is bounced counts.
func b13OtherCreatureYouControlLeftWithoutDying(ev game.Event, source *game.Card, g *game.Game) bool {
	return ev.CardID != source.InstanceID && creatureYouControlLeftWithoutDying(ev, source, g)
}

// creatureYouControlLeftWithoutDying is the same condition without
// the word "other" — Aang, Airbending Master counts himself. When the
// departing creature IS the source, the harvester's CR 603.10a
// look-back hands this the card that left, so the controller test is
// trivially the source's own.
func creatureYouControlLeftWithoutDying(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone == game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftAsType(ev, c, "creature") && leftUnderControlOf(ev, c) == source.Controller
}

// --- power of a departed permanent ---------------------------------

// b13LastKnownPower is a dead creature's power as it last was on the
// battlefield: the departure record's power (layers applied, every P/T
// counter kind, CR 122.1a), floored at zero the way CurrentPower floors
// it. `lki` is the harvester's characteristic, used only when the card
// has no departure record (it never left through the battlefield exit).
func b13LastKnownPower(g *game.Game, cardID uuid.UUID, lki game.Characteristic) int {
	p := lki.Power
	if info, ok := g.LastKnownPermanentForEffect(cardID); ok {
		p = info.Power
	}
	if p < 0 {
		return 0
	}
	return p
}

// b13ResolutionInProgressBy is the controller of the spell or ability
// whose resolution is happening right now — the Actor of the most
// recent EventResolve — or uuid.Nil when an action that cannot happen
// mid-resolution has been logged since: a cast, a mana ability, an
// attack declaration, a fizzle, or a step beginning. It is
// b12ResolvingController read from the end of the log rather than
// from an event's position, for a replacement effect that has no
// event of its own to anchor on.
func b13ResolutionInProgressBy(g *game.Game) uuid.UUID {
	for i := len(g.Events) - 1; i >= 0; i-- {
		switch ev := g.Events[i]; ev.Kind {
		case game.EventResolve:
			return ev.Actor
		case game.EventCast, game.EventManaAbilityActivated, game.EventAttack, game.EventFizzle,
			game.EventBeginUpkeep, game.EventBeginPrecombatMain, game.EventBeginEndStep:
			return uuid.Nil
		}
	}
	return uuid.Nil
}

// --- predicates --------------------------------------------------

// b13WithoutPlusOneCounter is Wave Goodbye's "creature without a
// +1/+1 counter on it".
func b13WithoutPlusOneCounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Counters["+1/+1"] <= 0
	}
}

// b13PowerGreaterThan is Fell the Mighty's "creatures with power
// greater than N", current power so counters and anthems count.
func b13PowerGreaterThan(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.CurrentPower() > n
	}
}

// b13ManaValueAtMostControllersGraveyard is Drown in the Loch's
// clause, one predicate for both modes: the candidate's mana value
// is at most the number of cards in ITS CONTROLLER's graveyard. A
// spell on the stack reads its controller off the stack item and its
// mana value with X included (CR 202.3e); a creature on the
// battlefield reads its controller off the card and a mana value
// where X is zero (CR 202.3e). Re-run at resolution like every
// target predicate, so a graveyard that shrank in response can
// legally take the target out.
func b13ManaValueAtMostControllersGraveyard() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		controller := c.Controller
		if item := g.StackItemForEffect(c.InstanceID); item != nil {
			controller = item.Controller
		}
		mv, ok := g.ManaValueForEffect(c)
		if !ok {
			return false
		}
		p := g.PlayerByIDForEffect(controller)
		if p == nil || p.Graveyard == nil {
			return false
		}
		return mv <= p.Graveyard.Size()
	}
}

// --- board reads -------------------------------------------------

// b13AttackingCreaturesYouControl snapshots the creatures
// `controller` controls that are attacking right now — Drana's
// "each attacking creature you control". AttackingTarget is stamped
// at declaration and cleared as the end of combat step ENDS
// (CR 511.3), so a trigger that resolves during a combat damage step
// or the end of combat step still sees the attackers.
func b13AttackingCreaturesYouControl(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.AttackingTarget != uuid.Nil {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- effect bodies -----------------------------------------------

// b13PutCounterOnEach puts one +1/+1 counter on every listed
// permanent still on the battlefield, in the given order.
func b13PutCounterOnEach(ctx *Context, ids []uuid.UUID) error {
	return b13PutCounterOnEachThen(ctx, ids, nil)
}

// b13PutCounterOnEachThen is b13PutCounterOnEach with a continuation
// that runs once every counter in the list has SETTLED (#1290) —
// Finneas, Ace Archer's "then if creatures you control have total
// power 10 or greater, draw a card" has to read power AFTER all of
// the counters have landed, and any one of them can pause on a CR 616
// ordering prompt (a Doubling Season / Hardened Scales board). The
// placements are chained one at a time — the next one is only
// attempted from inside the previous one's own continuation — so
// `then` cannot run until the whole list, not just the first
// placement, has landed. Nil `then` is b13PutCounterOnEach.
func b13PutCounterOnEachThen(ctx *Context, ids []uuid.UUID, then func(g *game.Game) error) error {
	if len(ids) == 0 {
		if then == nil {
			return nil
		}
		return then(ctx.Game)
	}
	id, rest := ids[0], ids[1:]
	item := ctx.Item
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return b13PutCounterOnEachThen(ctx, rest, then)
	}
	return ctx.Game.AddCounterThenForEffect(id, "+1/+1", 1, func(g *game.Game, _ int) error {
		return b13PutCounterOnEachThen(NewContext(g, item), rest, then)
	})
}

// b13CreateTappedTreasures is "create N tapped Treasure tokens" —
// Goldvein Hydra's payout. Zero or fewer makes nothing.
func b13CreateTappedTreasures(ctx *Context, controller uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	return CreateTokenAdvanced{
		Controller: controller,
		Spec:       Token(TreasureToken()).EntersTapped(),
		N:          n,
	}.Apply(ctx)
}
