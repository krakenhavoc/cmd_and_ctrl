package game

import (
	"testing"

	"github.com/google/uuid"
)

// granted_triggers_test.go — ADR 0093 PR 3, the engine half: a GRANTED
// triggered ability fires from the host, belongs to the host's
// controller, reaches a creature entering under an existing grant
// (CR 603.6a), and a granted dies trigger fires from last-known
// information in a single death and in a simultaneous wipe (CR 603.10a).

const (
	gtGrantorOracle = "granted-trigger-fixture-grantor" // creatures you control have both bundles
	gtDiesBundle    = "granted-trigger-fixture/dies"
	gtEntersBundle  = "granted-trigger-fixture/enters"
	gtDiesLabel     = "granted fixture — when this creature dies"
	gtEntersLabel   = "granted fixture — when this creature enters"
)

func stubGrantedTriggers(t *testing.T) {
	t.Helper()
	noop := func(*Game, *StackItem) error { return nil }
	self := func(label string, kind EventKind, extra func(Event) bool) TriggeredAbility {
		return TriggeredAbility{
			Watches: []EventKind{kind},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID && extra(ev)
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, label, noop)
			},
		}
	}
	defs := map[string]*CardDef{
		gtGrantorOracle: {Static: []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			GrantAbilities: []string{gtDiesBundle, gtEntersBundle},
		}}},
		GrantKey(gtDiesBundle): {
			Triggered: []TriggeredAbility{self(gtDiesLabel, EventLTB, func(ev Event) bool { return ev.NewZone == ZoneGraveyard })},
			GrantText: "When this creature dies, …",
		},
		GrantKey(gtEntersBundle): {
			Triggered: []TriggeredAbility{self(gtEntersLabel, EventETB, func(Event) bool { return true })},
			GrantText: "When this creature enters, …",
		},
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
}

// triggeredFrom reports every queued or stacked trigger whose source is
// `host` and label `label`, with its controller.
func triggeredFrom(g *Game, host uuid.UUID, label string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.SourceCardID == host && it.Label == label {
				out = append(out, it.Controller)
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.SourceCardID == host && it.Label == label {
				out = append(out, it.Controller)
			}
		}
	})
	return out
}

func TestGrantedDiesTriggerFiresFromLastKnownInformation(t *testing.T) {
	stubGrantedTriggers(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := gaBear(g, me)
	gaPush(g, me, "Grantor", "Enchantment", gtGrantorOracle)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if got := triggeredFrom(g, bear, gtDiesLabel); len(got) != 1 || got[0] != me {
		t.Fatalf("dies trigger from the host = %v, want one under the host's controller", got)
	}
}

func TestGrantedDiesTriggerFiresForEveryCreatureInAWipe(t *testing.T) {
	stubGrantedTriggers(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	a := gaBear(g, me)
	b := gaBear(g, me)
	grantor := gaPush(g, me, "Grantor", "Creature — Wall", gtGrantorOracle)
	// The grantor dies with them: the batch copies were taken while its
	// grant still applied (CR 603.10a), so every creature's dies
	// trigger — the grantor's own included — still fires.
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{a, b, grantor}) })
	for _, id := range []uuid.UUID{a, b, grantor} {
		if got := triggeredFrom(g, id, gtDiesLabel); len(got) != 1 {
			t.Errorf("%s: %d dies triggers, want 1", id, len(got))
		}
	}
}

func TestGrantedEntersTriggerReachesACreatureEnteringUnderTheGrant(t *testing.T) {
	stubGrantedTriggers(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	gaPush(g, me, "Grantor", "Enchantment", gtGrantorOracle)
	p := g.playerByIDLocked(me)
	card := NewCard("Newcomer", me)
	card.TypeLine = "Creature — Bear"
	p.Hand.PushTop(card)
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me}, ZoneRef{Kind: ZoneBattlefield}, card.InstanceID); err != nil {
		t.Fatalf("enter: %v", err)
	}
	if got := triggeredFrom(g, card.InstanceID, gtEntersLabel); len(got) != 1 {
		t.Errorf("CR 603.6a: a creature entering under the grant has its granted ETB trigger, got %d", len(got))
	}
}

// After a control change the granted trigger belongs to the host's NEW
// controller — the grant is the host's ability, not the grantor's.
func TestGrantedTriggerBelongsToTheHostsController(t *testing.T) {
	stubGrantedTriggers(t)
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	// "Creatures you control" of THEIR grantor, on a creature THEY
	// control: the trigger is theirs, whoever owns the creature.
	borrowed := pushTypedTestCard(g, Card{Name: "Borrowed", TypeLine: "Creature — Bear", Owner: me, Controller: them, BaseController: them})
	gaPush(g, them, "Their Grantor", "Enchantment", gtGrantorOracle)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(borrowed) })
	if got := triggeredFrom(g, borrowed, gtDiesLabel); len(got) != 1 || got[0] != them {
		t.Errorf("the dies trigger's controller = %v, want the host's controller %s", got, them)
	}
}
