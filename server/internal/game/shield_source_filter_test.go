package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// shield_source_filter_test.go — #2026 (ADR 0108 §7, amendment of
// 2026-10-05): a preventFromSource shield whose sources are named by what
// they are not, by power, by combat status, by counters, by being
// colourless or by their controller, read as the damage would be dealt
// (CR 609.7b), and a restore point that keeps it.

// filterShield registers a preventFromSource shield for `controller`
// protecting `protect` (uuid.Nil: anything) from creatures passing `f`.
func filterShield(t *testing.T, g *Game, controller, protect uuid.UUID, f DamageSourceFilter, qs ...PermanentQuery) {
	t.Helper()
	g.WithWriteLock(func() {
		s := DamageShield{Controller: controller, ProtectPlayer: protect, Queries: qs, Filter: f, Label: "test filter shield"}
		if !g.PreventDamageFromSourceThisTurnForEffect(s) {
			t.Fatal("no shield registered")
		}
	})
}

// hits deals `n` from each source to `to`, one instruction each, and
// returns the life `to` lost.
func hits(t *testing.T, g *Game, to uuid.UUID, n int, sources ...uuid.UUID) int {
	t.Helper()
	before := lifeOf(g, to)
	g.WithWriteLock(func() {
		for _, s := range sources {
			if err := g.DealDamageToPlayerForEffect(s, to, n); err != nil {
				t.Fatal(err)
			}
		}
	})
	return before - lifeOf(g, to)
}

func freshLayers(g *Game) {
	g.WithWriteLock(func() {
		g.layerVersion.Add(1)
		g.RecomputeLayersIfStaleLocked()
	})
}

var creatureQuery = PermanentQuery{Types: []string{"creature"}}

// "Non-Spider creatures" (Arachnogenesis): a Spider and a changeling
// (every creature type, CR 702.73a) deal their damage; a Bear does not.
func TestSourceFilterExceptsASubtype(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spiderCard := NewCard("Spider", opp.ID)
	spiderCard.TypeLine = "Creature — Spider"
	spiderCard.Power, spiderCard.Toughness = 2, 2
	g.Battlefield.PushTop(spiderCard)
	spider := spiderCard.InstanceID
	shape := pushCombatant(t, g, opp, "Shapeshifter", 2, 2, "changeling")
	bear := pushCombatant(t, g, opp, "Bear", 2, 2)
	freshLayers(g)
	filterShield(t, g, me.ID, me.ID, DamageSourceFilter{Except: []PermanentQuery{{Subtypes: []string{"Spider"}}}}, creatureQuery)
	if lost := hits(t, g, me.ID, 1, spider, shape, bear); lost != 2 {
		t.Fatalf("lost %d, want 2: the Spider's and the changeling's dealt, the Bear's prevented", lost)
	}
}

// "With power 3 or less" (Fog of War) is read as the damage would be
// dealt, counters included (CR 613.4c): a 2-power creature's damage is
// prevented until two +1/+1 counters make it 4.
func TestSourceFilterPowerBoundCountsCounters(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := pushCombatant(t, g, opp, "Grower", 2, 2)
	filterShield(t, g, me.ID, me.ID, DamageSourceFilter{PowerBounded: true, PowerAtMost: 3}, creatureQuery)
	if lost := hits(t, g, me.ID, 2, c); lost != 0 {
		t.Fatalf("lost %d, want 0: a 2-power creature's damage is prevented", lost)
	}
	findBattlefieldCard(g, c).Counters = map[string]int{CounterPlusOne: 2}
	freshLayers(g)
	if lost := hits(t, g, me.ID, 2, c); lost != 2 {
		t.Fatalf("lost %d, want 2: at power 4 the creature is past the bound", lost)
	}
}

// "Colorless sources" (Lithomancer's Focus, CR 105.2c) and "with no
// +1/+1 counters" (Hindervines).
func TestSourceFilterColorlessAndNoCounters(t *testing.T) {
	t.Run("colorless", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		golem := pushColouredCreature(g, opp, "Golem", nil)
		red := pushColouredCreature(g, opp, "Red", []string{"R"})
		filterShield(t, g, me.ID, me.ID, DamageSourceFilter{Colorless: true})
		if lost := hits(t, g, me.ID, 2, golem, red); lost != 2 {
			t.Fatalf("lost %d, want 2: the colourless source prevented, the red one dealt", lost)
		}
	})
	t.Run("no counters", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		plain := pushCombatant(t, g, opp, "Plain", 2, 2)
		grown := pushCombatant(t, g, opp, "Grown", 2, 2)
		findBattlefieldCard(g, grown).Counters = map[string]int{CounterPlusOne: 1}
		freshLayers(g)
		filterShield(t, g, me.ID, me.ID, DamageSourceFilter{NoCounters: []string{CounterPlusOne}}, creatureQuery)
		if lost := hits(t, g, me.ID, 1, plain, grown); lost != 1 {
			t.Fatalf("lost %d, want 1: only the creature with a counter deals damage", lost)
		}
		findBattlefieldCard(g, plain).Counters = map[string]int{CounterPlusOne: 1}
		freshLayers(g)
		if lost := hits(t, g, me.ID, 1, plain); lost != 1 {
			t.Fatalf("lost %d, want 1: a counter put on after the shield is read too", lost)
		}
	})
}

// "Attacking creatures" (Heavy Fog) is read as the damage would be dealt:
// a creature removed from combat no longer matches, and one that has left
// the battlefield is read as it last existed there (CR 608.2h, Heavy
// Fog's ruling: "if it was an attacking creature at the time it left").
func TestSourceFilterAttackingIsReadAsTheDamageIsDealt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pushCombatant(t, g, opp, "Attacker", 2, 2)
	home := pushCombatant(t, g, opp, "Home", 2, 2)
	gone := pushCombatant(t, g, opp, "Gone", 2, 2)
	findBattlefieldCard(g, att).AttackingTarget = me.ID
	findBattlefieldCard(g, gone).AttackingTarget = me.ID
	filterShield(t, g, me.ID, me.ID, DamageSourceFilter{Combat: SourceCombatAttacking}, creatureQuery)

	if lost := hits(t, g, me.ID, 1, att, home); lost != 1 {
		t.Fatalf("lost %d, want 1: the attacker's damage prevented, the other's dealt", lost)
	}
	g.WithWriteLock(func() { g.removeFromCombatLocked(findBattlefieldCard(g, att)) })
	if lost := hits(t, g, me.ID, 1, att); lost != 1 {
		t.Fatalf("lost %d, want 1: a creature removed from combat is no longer attacking", lost)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(gone); err != nil {
			t.Fatal(err)
		}
		if err := g.DestroyPermanentForEffect(home); err != nil {
			t.Fatal(err)
		}
	})
	if lost := hits(t, g, me.ID, 1, gone); lost != 0 {
		t.Fatalf("lost %d, want 0: it was attacking when it left the battlefield", lost)
	}
	if lost := hits(t, g, me.ID, 1, home); lost != 1 {
		t.Fatalf("lost %d, want 1: it was not attacking when it left", lost)
	}
}

// "Unblocked creatures" (Snag, CR 509.1h) and "target blocked creature"
// (Benalish Missionary): the blocked record, not the live blockers.
func TestSourceFilterUnblockedAndTheBlockedReader(t *testing.T) {
	t.Run("blocked", func(t *testing.T) {
		g := newActiveGame(t)
		attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
		chump := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
		blockAfterLockIn(t, g, attacker, chump)
		var blocked, unblocked bool
		g.WithWriteLock(func() {
			blocked, unblocked = g.BlockedAttackerForEffect(attacker), g.UnblockedAttackerForEffect(attacker)
		})
		if !blocked || unblocked {
			t.Fatalf("blocked %v, unblocked %v: want a blocked attacker", blocked, unblocked)
		}
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(chump); err != nil {
				t.Fatal(err)
			}
			blocked = g.BlockedAttackerForEffect(attacker)
		})
		if !blocked {
			t.Fatal("an attacker whose blocker died is still blocked (CR 509.1h)")
		}
		filterShield(t, g, g.Seats[1].ID, uuid.Nil, DamageSourceFilter{Combat: SourceCombatUnblocked}, creatureQuery)
		if lost := hits(t, g, g.Seats[1].ID, 2, attacker); lost != 2 {
			t.Fatalf("lost %d, want 2: a blocked creature's damage is not an unblocked creature's", lost)
		}
	})
	t.Run("unblocked", func(t *testing.T) {
		g := newActiveGame(t)
		attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
		declareAttacks(t, g, attacker)
		var blocked bool
		g.WithWriteLock(func() {
			blocked = g.BlockedAttackerForEffect(attacker)
			g.completeAllBlockDeclarationsLocked()
			g.commitBlockDeclarationLocked()
		})
		if blocked {
			t.Fatal("an attacker before the defender has declared is neither blocked nor unblocked")
		}
		filterShield(t, g, g.Seats[1].ID, uuid.Nil, DamageSourceFilter{Combat: SourceCombatUnblocked}, creatureQuery)
		if lost := hits(t, g, g.Seats[1].ID, 2, attacker); lost != 0 {
			t.Fatalf("lost %d, want 0: the unblocked attacker's damage is prevented", lost)
		}
	})
}

// The controller tests, relative to the shield's controller: "sources you
// don't control" (Channel Harm), "your opponents control" (Thwart the
// Enemy) and "target opponent controls" (Encircling Fissure), each read
// as the damage would be dealt.
func TestSourceFilterController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := pushColouredCreature(g, me, "Mine", []string{"G"})
	theirs := pushColouredCreature(g, a, "Theirs", []string{"G"})
	others := pushColouredCreature(g, b, "Others", []string{"G"})
	filterShield(t, g, me.ID, me.ID, DamageSourceFilter{Controller: SourceControllerOpponents})
	if lost := hits(t, g, me.ID, 1, mine, theirs, others); lost != 1 {
		t.Fatalf("lost %d, want 1: only my own source's damage is dealt", lost)
	}

	g2 := newFourPlayerActiveGame(t)
	me, a, b = g2.Seats[0], g2.Seats[1], g2.Seats[2]
	theirs = pushColouredCreature(g2, a, "Theirs", []string{"G"})
	others = pushColouredCreature(g2, b, "Others", []string{"G"})
	filterShield(t, g2, me.ID, me.ID, DamageSourceFilter{Controller: SourceControllerPlayer, ControllerPlayer: a.ID})
	if lost := hits(t, g2, me.ID, 1, theirs, others); lost != 1 {
		t.Fatalf("lost %d, want 1: only the named player's source is stopped", lost)
	}
	findBattlefieldCard(g2, others).Controller = a.ID
	findBattlefieldCard(g2, others).effective = nil
	if lost := hits(t, g2, me.ID, 1, others); lost != 0 {
		t.Fatalf("lost %d, want 0: the controller is read as the damage is dealt", lost)
	}
}

// Pinned exceptions (Haze Frog's "other creatures", Terrifying
// Presence's "other than target creature") name one object (CR 400.7):
// the same card as a new object is "other".
func TestSourceFilterExceptObjectsPinTheObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frog := pushColouredCreature(g, opp, "Frog", []string{"G"})
	bear := pushColouredCreature(g, opp, "Bear", []string{"G"})
	ref := ObjectRef{ID: frog, Epoch: findBattlefieldCard(g, frog).ObjectEpoch}
	filterShield(t, g, me.ID, me.ID, DamageSourceFilter{ExceptObjects: []ObjectRef{ref}}, creatureQuery)
	if lost := hits(t, g, me.ID, 1, frog, bear); lost != 1 {
		t.Fatalf("lost %d, want 1: the pinned Frog's damage dealt, the Bear's prevented", lost)
	}
	findBattlefieldCard(g, frog).ObjectEpoch++
	if lost := hits(t, g, me.ID, 1, frog); lost != 0 {
		t.Fatalf("lost %d, want 0: a new object is another creature", lost)
	}
}

// Inspire Awe's "except … enchanted creatures and enchantment
// creatures": Enchanted is read off the battlefield.
func TestSourceFilterExceptEnchanted(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := pushColouredCreature(g, opp, "Plain", []string{"G"})
	worn := pushColouredCreature(g, opp, "Worn", []string{"G"})
	aura := NewCard("Aura", opp.ID)
	aura.TypeLine = "Enchantment — Aura"
	aura.AttachedTo = TargetRef{Kind: TargetCard, ID: worn}
	g.Battlefield.PushTop(aura)
	filterShield(t, g, me.ID, me.ID, DamageSourceFilter{Except: []PermanentQuery{{Enchanted: true}, {Types: []string{"enchantment"}}}})
	if lost := hits(t, g, me.ID, 1, plain, worn); lost != 1 {
		t.Fatalf("lost %d, want 1: the enchanted creature's damage dealt, the other's prevented", lost)
	}
}

// What registration and restore refuse.
func TestSourceFilterRefusals(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Src", []string{"R"})
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		bad := []DamageShield{
			{Controller: me.ID, Source: ref, SourceZone: zone, Filter: DamageSourceFilter{Colorless: true}},
			{Controller: me.ID, ProtectPermanent: src, AndDealtBy: true, Filter: DamageSourceFilter{Colorless: true}},
			{Controller: me.ID, Filter: DamageSourceFilter{Combat: "blocking"}},
			{Controller: me.ID, Filter: DamageSourceFilter{Controller: SourceControllerPlayer}},
			{Controller: me.ID, Filter: DamageSourceFilter{PowerAtMost: 3}},
		}
		for i, s := range bad {
			if g.PreventDamageFromSourceThisTurnForEffect(s) {
				t.Errorf("shield %d registered: %+v", i, s.Filter)
			}
		}
	})
	f := []DamageSourceFilter{{Colorless: true}}
	for _, m := range []Mod{
		{Kind: ModPreventNextFromSource, Queries: []PermanentQuery{creatureQuery}, SourceFilter: f},
		{Kind: ModPreventDamage, Amount: 1, SourceFilter: f},
		{Kind: ModPreventFromSource, SourceFilter: []DamageSourceFilter{{}}},
		{Kind: ModPreventFromSource, SourceFilter: []DamageSourceFilter{{Colorless: true}, {Colorless: true}}},
		{Kind: ModPreventFromSource, SourceFilter: []DamageSourceFilter{{Controller: "teammates"}}},
	} {
		if nextFromSourceModProblem(m) == "" {
			t.Errorf("mod %+v passed the check", m)
		}
	}
	// A per-source follow-up is a source shield's, with a follow-up.
	for _, m := range []Mod{
		{Kind: ModPreventDamage, Amount: 1, Then: "prevention/gain-life-equal", ThenPer: ThenPerSource},
		{Kind: ModPreventFromSource, ThenPer: ThenPerSource},
		{Kind: ModPreventFromSource, Then: "prevention/gain-life-equal", ThenPer: ThenPerRecipient},
	} {
		if followUpModProblem(m) == "" {
			t.Errorf("mod %+v passed the follow-up check", m)
		}
	}
	if p := followUpModProblem(Mod{Kind: ModPreventFromSource, Then: "prevention/gain-life-equal", ThenPer: ThenPerSource}); p != "" {
		t.Errorf("a per-source follow-up on a source shield was refused: %s", p)
	}
}

// A filtered shield is a restore point: it comes back and still filters,
// and a key inside the filter that this binary does not know is refused
// (ADR 0041 P4), not dropped.
func TestSourceFilterSurvivesARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := pushColouredCreature(g, opp, "Golem", nil)
	red := pushColouredCreature(g, opp, "Red", []string{"R"})
	pinned := ObjectRef{ID: red, Epoch: findBattlefieldCard(g, red).ObjectEpoch}
	filterShield(t, g, me.ID, me.ID, DamageSourceFilter{
		Except:        []PermanentQuery{{Subtypes: []string{"Spider"}}},
		ExceptObjects: []ObjectRef{pinned},
		Combat:        SourceCombatAny,
		Controller:    SourceControllerOpponents,
		NoCounters:    []string{CounterPlusOne},
	})
	_, restored := roundTrip(t, g)
	if lost := hits(t, restored, me.ID, 1, golem, red); lost != 1 {
		t.Fatalf("lost %d, want 1 after the restore: the Golem prevented, the pinned creature dealt", lost)
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	effects := doc["scopedEffects"].([]any)
	mod := effects[0].(map[string]any)["mods"].([]any)[0].(map[string]any)
	mod["sourceFilter"].([]any)[0].(map[string]any)["teammates"] = true
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore error %v, want ErrUnknownEffectKey for an unknown source-filter key", err)
	}
}
