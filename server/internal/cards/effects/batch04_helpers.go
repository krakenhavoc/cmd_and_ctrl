package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch04_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 04 (#297, `edhrec_rank` 472–577).
//
// Its own file rather than helpers.go, per the convention #231 set:
// concurrent card batches collide on shared helper files. Every
// package-level name here carries a b04 prefix for the same reason —
// batches 02 and 03 are being written at the same time.

// b04BounceLand is the Ravnica "bounce land" / karoo shape:
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {A}{B}."
//
// Azorius Chancery (Aang deck) was the first and taps its own entry
// in OnETB; that file's comment predates the entering-card block in
// gatherActiveReplacementsLocked, and the three lands here use the
// real CR 614 self-replacement instead, exactly as every conditional
// dual does. The bounce is modelled as a targeted trigger for the
// reason the Chancery file gives: the printed "return a land you
// control" is a choice, not a target, and the pick_target prompt is
// the one picker the engine has for a choice among permanents. It
// can never fizzle — the land itself is always a legal answer, and
// returning it is a real line.
func b04BounceLand(oracleID, name, produced string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You pick the land to return when the trigger goes on the stack rather than on resolution, so opponents can respond to the choice."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: produced,
			Label:    "Add " + produced,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("a land you control", And(Land(), YouControl())),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, name+" — return a land you control to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return BounceToHand{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
					})
			},
		}},
	}
}

// b04NontokenCreature is "a nontoken creature" — Accursed Marauder's
// edict clause, the one word that separates it from Fleshbag
// Marauder: a Treasure deck's Zombie tokens cannot be fed to it.
func b04NontokenCreature(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && !IsToken(c)
}

// b04LandNamesControlled counts the lands `controller` controls
// with DIFFERENT names — Field of the Dead's seven. Reads the printed
// name; two copies of the same land count once, as printed.
func b04LandNamesControlled(g *game.Game, controller uuid.UUID) int {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsLand() {
			seen[c.Name] = true
		}
	}
	return len(seen)
}

// b04CreaturesControlled counts the creatures `controller` controls
// on the printed type line — Adeline's characteristic-defining
// power. Counts Adeline herself, as printed ("creatures you
// control", not "other"). Walks the live slice: it runs inside a
// layer recompute.
func b04CreaturesControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsCreature() {
			n++
		}
	}
	return n
}

// b04CreatureIDsControlledBy snapshots the instance IDs of the
// creatures `controller` controls, for effects that mutate the
// battlefield while iterating it (Cathars' Crusade's counters).
func b04CreatureIDsControlledBy(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// b04OpponentLostLife reports whether ev is an opponent of
// `controller` losing life, and how much — Exquisite Blood's
// trigger. Two event kinds carry a life loss: EventChangeLife with a
// negative Amount (a drain, a "loses N life" effect, a Bond) and
// EventDealDamage whose Target is a player (a Bolt, combat damage),
// because the damage paths write the life total directly and emit
// no EventChangeLife of their own. Neither kind is emitted for the
// other's loss, so nothing is counted twice.
func b04OpponentLostLife(ev game.Event, controller uuid.UUID, g *game.Game) (int, bool) {
	if ev.Target == uuid.Nil || ev.Target == controller || g.PlayerByIDForEffect(ev.Target) == nil {
		return 0, false
	}
	switch ev.Kind {
	case game.EventChangeLife:
		if ev.Amount < 0 {
			return -ev.Amount, true
		}
	case game.EventDealDamage:
		if ev.Amount > 0 {
			return ev.Amount, true
		}
	}
	return 0, false
}

// b04TriggerPendingOrOnStack reports whether a triggered ability
// from `source` is already waiting on PendingTriggers OR already
// sitting on the stack — the "whenever you attack" dedup Adeline
// needs, and a wider net than batch 01's triggerAlreadyPendingFrom.
//
// The difference is where the events come from. Combat damage
// (Professional Face-Breaker) fires every creature's event inside
// one mutation, so the first trigger is still on PendingTriggers
// when the second event arrives. Attackers declared one at a time
// through DeclareAttacker are one mutation EACH, and each ends with
// runStateChecksLocked, which drains the queue onto the stack — so
// by the second declaration the first Adeline trigger is in
// StackMeta, not PendingTriggers, and a queue-only check fires her
// once per attacker. The batch DeclareAttackers path emits every
// event before its single drain, so it is covered by either check.
//
// The residual gap runs the weaker way: once the trigger has
// RESOLVED, a further attacker declared in the same step (the
// sandbox permits it; paper does not) fires it again.
func b04TriggerPendingOrOnStack(g *game.Game, source *game.Card) bool {
	if triggerAlreadyPendingFrom(g, source) {
		return true
	}
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source.InstanceID {
			return true
		}
	}
	return false
}

// b04ZombieToken is the 2/2 black Zombie Field of the Dead makes.
func b04ZombieToken() game.Card {
	return game.Card{
		Name:      "Zombie",
		TypeLine:  "Token Creature — Zombie",
		Power:     2,
		Toughness: 2,
	}
}
