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
// fires. Returns the creature's ID for the trigger to destroy.
func b13CreatureDealtDamageToYou(ev game.Event, source *game.Card, g *game.Game) (uuid.UUID, bool) {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target != source.Controller {
		return uuid.Nil, false
	}
	if z := g.FindCardZoneForEffect(ev.Source); z == nil || z.Kind != game.ZoneBattlefield {
		return uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	if !ok || !c.IsCreature() {
		return uuid.Nil, false
	}
	return c.InstanceID, true
}

// b13OtherCreatureYouControlLeftWithoutDying is Dour Port-Mage's
// condition: another creature the source's controller controlled
// left the battlefield for anywhere but a graveyard — a bounce, an
// exile, a library tuck. EventLTB carries the destination in
// NewZone, so "without dying" is the one comparison CR 700.4 makes.
//
// The card is read post-move (diedCreature's posture): the printed
// type line and the Controller field survive the move, so "creature
// you control" is the creature that just left under the source's
// controller's control. A creature that was a creature only through
// a layer effect, or a stolen creature that went back to its owner's
// hand, is not counted — weaker than printed, never stronger.
func b13OtherCreatureYouControlLeftWithoutDying(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone == game.ZoneGraveyard || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller
}

// --- counters read back off the log ------------------------------

// b13LastKnownCounters is the number of `kind` counters `cardID` had
// the last time it was on the battlefield — read for a permanent
// that has just left it, which is what a dies trigger (Goldvein
// Hydra's "equal to its power") and a sacrifice-cost ability
// (Twitching Doll's "for each counter on this creature") need and
// what neither can get from the card itself: MoveCard clears
// Counters on the way out (CR 400.7), and the CR 603.10 LKI
// characteristic the harvester hands a dies trigger carries the
// layer-computed P/T with no counter math.
//
// EventCounterPlaced carries the post-change total, so the most
// recent one for the card and kind IS the count at the moment it
// left. The walk stops at the card's arrival on the battlefield —
// an earlier life of the same instance (a reanimated card keeps its
// InstanceID) is not this one — except that "enters with" counters
// are placed a beat BEFORE the arrival is logged, by every entry
// site, so the placements immediately preceding the arrival are
// still read (b13LastKnownCounterWalk). Zero when the kind was never
// placed, or was removed to nothing.
func b13LastKnownCounters(g *game.Game, cardID uuid.UUID, kind string) int {
	total := 0
	b13LastKnownCounterWalk(g, cardID, func(ev game.Event) bool {
		if ev.Label != kind {
			return true
		}
		total = ev.Amount
		return false
	})
	return total
}

// b13LastKnownCounterTotal is the total number of counters of every
// kind `cardID` had when it last left the battlefield — Twitching
// Doll counts nest counters and anything else alike. Same walk as
// b13LastKnownCounters, taking the most recent total per kind.
func b13LastKnownCounterTotal(g *game.Game, cardID uuid.UUID) int {
	seen := map[string]bool{}
	total := 0
	b13LastKnownCounterWalk(g, cardID, func(ev game.Event) bool {
		if !seen[ev.Label] {
			seen[ev.Label] = true
			total += ev.Amount
		}
		return true
	})
	return total
}

// b13LastKnownCounterWalk hands `visit` every EventCounterPlaced for
// `cardID` from the card's most recent life on the battlefield,
// newest first, until visit returns false or the life runs out.
//
// The life's start is the card's arrival — its EventETB, or its
// EventTokenCreated. Past that, only the events one entry site emits
// about the card between placing its "enters with" counters and
// logging the arrival are stepped over (the zone move onto the
// battlefield, the resolve of the spell it was, a copy applied); the
// first event that is none of those is the previous life, or the
// game before the card, and the walk ends.
func b13LastKnownCounterWalk(g *game.Game, cardID uuid.UUID, visit func(ev game.Event) bool) {
	arrived := false
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventCounterPlaced && ev.Target == cardID {
			if !visit(ev) {
				return
			}
			continue
		}
		if !arrived {
			if ev.CardID == cardID && (ev.Kind == game.EventETB || ev.Kind == game.EventTokenCreated) {
				arrived = true
			}
			continue
		}
		if ev.CardID != cardID {
			return
		}
		switch ev.Kind {
		case game.EventResolve, game.EventCopyApplied, game.EventTokenCreated:
			continue
		case game.EventZoneMove:
			if ev.NewZone == game.ZoneBattlefield {
				continue
			}
		}
		return
	}
}

// b13LastKnownPower is a dead creature's power as it last was on the
// battlefield: the harvester's LKI characteristic (layers applied)
// plus its +1/+1 counters, minus its -1/-1 counters, floored at zero
// the way CurrentPower floors it.
func b13LastKnownPower(g *game.Game, cardID uuid.UUID, lki game.Characteristic) int {
	p := lki.Power + b13LastKnownCounters(g, cardID, "+1/+1") - b13LastKnownCounters(g, cardID, "-1/-1")
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
	for _, id := range ids {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
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
