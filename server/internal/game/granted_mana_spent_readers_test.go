package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// granted_mana_spent_readers_test.go — ADR 0109 §11 (#1552), the engine
// half: sunburst as a keyword read off the resolving spell, a keyword
// given to a spell keeping its duration on the permanent, the spend
// record read inside the entry window, the spend rider that gives a
// spell a keyword, the counted entry rider, and the restore refusal of
// a rider this binary cannot read. The printed cards are pinned in
// cards/effects/granted_mana_spent_readers_test.go.

// seedSunburst puts a 0/0 sunburst permanent of `typeLine` costing {4}
// into p's hand. The keyword rides Card.Keywords, the deck importer's
// road, so the engine needs no catalog entry to honour it.
func seedSunburst(t *testing.T, g *Game, p *Player, typeLine string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Sunburst", p.ID)
	c.TypeLine = typeLine
	c.ManaCost = "{4}"
	c.Controller = p.ID
	c.Keywords = []string{KeywordSunburst}
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

// CR 702.44a: a +1/+1 counter per colour of mana spent for an object
// entering as a creature, a charge counter otherwise. Colourless mana is
// not a colour.
func TestSunburstIsAKeywordCountingColoursSpent(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, kind string
	}{
		{"a creature", "Artifact Creature — Golem", CounterPlusOne},
		{"a noncreature artifact", "Artifact", CounterCharge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := seedSunburst(t, g, me, tc.typeLine)
			floatMana(me, "WUBC")
			if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			resolveTop(t, g)
			if got := counters(t, g, id, tc.kind); got != 3 {
				t.Errorf("%s counters = %d, want 3 — W, U and B, and {C} is no colour", tc.kind, got)
			}
		})
	}
}

// CR 702.44d and CR 400.7a: a spell GIVEN sunburst on the stack (Lux
// Artillery) has it as it enters, and a second instance counts the
// colours again.
func TestSunburstGivenToTheSpellCountsEachInstance(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSunburst(t, g, me, "Artifact Creature — Golem")
	floatMana(me, "WUGG")
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if !g.GrantKeywordsToSpellForEffect(uuid.New(), id, []string{KeywordSunburst}, IndefiniteDuration(), "test — it gains sunburst") {
			t.Fatal("the grant was refused")
		}
	})
	resolveTop(t, g)
	if got := counters(t, g, id, CounterPlusOne); got != 6 {
		t.Errorf("+1/+1 counters = %d, want 6 — two instances, three colours each", got)
	}
}

// CR 702.44b and ADR 0068 §3: a payment the engine waived claims no
// colours, so sunburst adds nothing.
func TestSunburstOnAWaivedPaymentAddsNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedSunburst(t, g, me, "Artifact")
	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	if got := counters(t, g, id, CounterCharge); got != 0 {
		t.Errorf("charge counters = %d on an unrecorded payment, want 0", got)
	}
}

// advancePastThisTurn walks the cursor until a new turn has begun, so
// the cleanup step's "until end of turn" sweep has run.
func advancePastThisTurn(t *testing.T, g *Game) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	for i := 0; g.Turn.ActiveSeat == seat; i++ {
		if i > 64 {
			t.Fatal("the turn never ended")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// ADR 0109 §11 decision 3, CR 400.7a and 611.2a: "it gains haste until
// end of turn" given to a creature spell is haste on the creature until
// end of turn, and no longer. Before this change the re-pin to the
// permanent made every spell grant indefinite.
func TestASpellGrantKeepsItsDurationOnThePermanent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	item := castSpellForControlTest(t, g, me, "Creature — Elf")
	g.WithWriteLock(func() {
		if !g.GrantKeywordsToSpellForEffect(uuid.New(), item.ID, []string{"haste"}, g.UntilEndOfTurnDuration(), "test — it gains haste until end of turn") {
			t.Fatal("GrantKeywordsToSpellForEffect refused a creature spell")
		}
	})
	resolveTop(t, g)
	if c := layeredBattlefieldCard(t, g, item.ID); !HasKeyword(&c, "haste") {
		t.Fatal("the creature does not have the haste its spell was given")
	}
	advancePastThisTurn(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := layeredBattlefieldCard(t, g, item.ID); HasKeyword(&c, "haste") {
		t.Error("the creature still has haste after the turn ended")
	}
}

// EntrySpentForEffect: the entering spell's whole spend, and the known
// zero for an entry that was not a cast.
func TestEntrySpentForEffectReadsTheEnteringSpell(t *testing.T) {
	g := newActiveGame(t)
	paid := &StackItem{Paid: PaidCost{Mana: []ManaToken{
		treasureMana("R"), {Color: "G", Source: uuid.New(), SourceKinds: ManaSourceArtifact},
		{Color: "G", Source: uuid.New(), SourceKinds: ManaSourceLand},
	}}}
	waived := &StackItem{Paid: PaidCost{OnPaper: true}}
	g.WithWriteLock(func() {
		spent := g.EntrySpentForEffect(&ReplacementEvent{stackItem: paid})
		if spent.Total() != 3 || spent.CountFrom(ManaSourceArtifact) != 2 || spent.CountFrom(ManaSourceTreasure) != 1 {
			t.Errorf("cast: total %d, artifact %d, treasure %d; want 3, 2, 1",
				spent.Total(), spent.CountFrom(ManaSourceArtifact), spent.CountFrom(ManaSourceTreasure))
		}
		if none := g.EntrySpentForEffect(&ReplacementEvent{}); !none.Known() || !none.None() {
			t.Error("an entry that was not a cast should read as a known nothing")
		}
		if g.EntrySpentForEffect(nil).Total() != 0 {
			t.Error("a nil event should read as nothing")
		}
		if w := g.EntrySpentForEffect(&ReplacementEvent{stackItem: waived}); w.Known() || w.None() {
			t.Error("a waived payment should read as unknown, never as no mana spent")
		}
	})
}

// stubRiderCount registers a counted-rider key once per process.
func stubRiderCount(key string, c ManaRiderCount) {
	if _, ok := ManaRiderCountFor(key); !ok {
		RegisterManaRiderCount(key, c)
	}
}

// riderToken is one mana carrying `r`, stamped with a production.
func riderToken(color string, production uuid.UUID, r ManaSpendRider) ManaToken {
	r.Production = production
	return ManaToken{Color: color, Source: uuid.New(), Riders: []ManaSpendRider{r}}
}

// ManaRiderSpellGains: "if that mana is spent on a creature spell, it
// gains haste until end of turn" (Generator Servant). The spell gains
// it at the spend, the creature has it on entering, and it is gone once
// the turn ends. Mana spent on a noncreature spell gives nothing.
func TestSpellGainsRiderGivesTheSpellAKeywordForItsDuration(t *testing.T) {
	rider := ManaSpendRider{
		Kind: ManaRiderSpellGains, When: []string{ManaRestrictCast, "type:Creature"},
		Keywords: []string{"haste"}, UntilEndOfTurn: true,
	}
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, StepPrecombatMain)
	production := uuid.New()
	g.WithWriteLock(func() {
		me.ManaPool.AddMana(riderToken("C", production, rider))
		me.ManaPool.AddMana(riderToken("C", production, rider))
	})
	c := NewCard("Test Elf", me.ID)
	c.TypeLine, c.ManaCost, c.Controller, c.Power, c.Toughness = "Creature — Elf", "{2}", me.ID, 2, 2
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	grants := 0
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			for _, a := range e.Affected {
				if a.OnStack && a.ID == c.InstanceID {
					grants++
				}
			}
		}
	})
	if grants != 1 {
		t.Errorf("%d grants pinned to the spell, want 1 — one production is one \"that mana\"", grants)
	}
	resolveTop(t, g)
	if got := layeredBattlefieldCard(t, g, c.InstanceID); !HasKeyword(&got, "haste") {
		t.Fatal("the creature entered without haste")
	}
	advancePastThisTurn(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := layeredBattlefieldCard(t, g, c.InstanceID); HasKeyword(&got, "haste") {
		t.Error("the haste outlived the turn")
	}

	// The same mana on an artifact spell: the filter refuses it.
	g2 := newActiveGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	advanceTo(t, g2, StepPrecombatMain)
	g2.WithWriteLock(func() { me2.ManaPool.AddMana(riderToken("C", uuid.New(), rider)) })
	rock := NewCard("Test Rock", me2.ID)
	rock.TypeLine, rock.ManaCost, rock.Controller = "Artifact", "{1}", me2.ID
	g2.WithWriteLock(func() { me2.Hand.PushTop(rock) })
	if err := g2.CastSpell(me2.ID, rock.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if n := len(g2.ScopedEffects); n != 0 {
		t.Errorf("an artifact spell was given %d grants by mana that only hastes creatures", n)
	}
}

// A counted entry rider (Opal Palace): its Condition decides at the
// spend whether it applies, and its Count is read as the permanent
// enters.
func TestCountedEntryRiderCountsAsThePermanentEnters(t *testing.T) {
	const key = "test-1552-counted-entry-rider"
	stubRiderCount(key, ManaRiderCount{
		Condition: func(_ *Game, controller uuid.UUID, paidFor Card) bool {
			return paidFor.IsCommander && paidFor.Owner == controller
		},
		Count: func(g *Game, controller uuid.UUID, entering Card) int {
			if p := g.playerByIDLocked(controller); p != nil {
				return p.CommanderCasts[entering.InstanceID]
			}
			return 0
		},
	})
	rider := ManaSpendRider{Kind: ManaRiderEntersWithCounters, CounterKind: CounterPlusOne, Count: key,
		When: []string{ManaRestrictCast}}
	for _, tc := range []struct {
		name      string
		commander bool
		want      int
	}{
		{"the commander, cast from the command zone for the third time", true, 3},
		{"a creature that is not a commander", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceTo(t, g, StepPrecombatMain)
			c := NewCard("Test Commander", me.ID)
			c.TypeLine, c.ManaCost, c.Controller, c.Power, c.Toughness = "Legendary Creature — Elf", "{1}", me.ID, 2, 2
			c.IsCommander = tc.commander
			g.WithWriteLock(func() {
				me.Hand.PushTop(c)
				// Two earlier casts from the command zone; this one,
				// from hand, does not bump the tally, so the count is
				// the two earlier casts plus one we add by hand to stand
				// for a third.
				me.CommanderCasts[c.InstanceID] = 3
				me.ManaPool.AddMana(riderToken("G", uuid.New(), rider))
			})
			if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			resolveTop(t, g)
			if got := counters(t, g, c.InstanceID, CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// ADR 0109 §11 snapshot impact: a rider kind, or a count key, this
// binary does not read was written by a newer one, and the file is
// refused rather than restored as mana whose rider silently does
// nothing.
func TestAnUnknownManaRiderIsRefusedOnRestore(t *testing.T) {
	const key = "test-1552-restorable-count"
	stubRiderCount(key, ManaRiderCount{Count: func(*Game, uuid.UUID, Card) int { return 1 }})
	for _, tc := range []struct {
		name   string
		rider  ManaSpendRider
		refuse bool
	}{
		{"a spell-gains rider", ManaSpendRider{Kind: ManaRiderSpellGains, Keywords: []string{"haste"}, UntilEndOfTurn: true}, false},
		{"a counted rider this binary registers", ManaSpendRider{Kind: ManaRiderEntersWithCounters, CounterKind: CounterPlusOne, Count: key}, false},
		{"a rider kind from a newer binary", ManaSpendRider{Kind: "gains_wings"}, true},
		{"a count key from a newer binary", ManaSpendRider{Kind: ManaRiderEntersWithCounters, CounterKind: CounterPlusOne, Count: "test-1552-no-such-count"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newRestorableGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			g.WithWriteLock(func() { me.ManaPool.AddMana(riderToken("C", uuid.New(), tc.rider)) })
			_, err := g.CaptureSnapshot().Restore()
			if refused := errors.Is(err, ErrUnknownEffectKey); refused != tc.refuse {
				t.Errorf("Restore: err = %v, want refused = %v", err, tc.refuse)
			}
		})
	}
}

// ADR 0109 owner decision 1, CR 614.12: sunburst is read off the permanent
// as it would exist on the battlefield. A creature entering under an
// ability-removing static ("Creatures lose all abilities") has no sunburst
// and enters with no counters; a noncreature artifact the static does not
// reach still gets its charge counters. Reading the stack card instead gave
// the creature three.
func TestSunburstIsReadOffThePermanentAsItWouldEnter(t *testing.T) {
	withRiotStatics(t)
	for _, tc := range []struct {
		name, typeLine, kind string
		want                 int
	}{
		{"a creature loses it", "Artifact Creature — Golem", CounterPlusOne, 0},
		{"a noncreature keeps it", "Artifact", CounterCharge, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			pushStaticSource(g, me.ID, testRiotRemovalOracle)
			id := seedSunburst(t, g, me, tc.typeLine)
			floatMana(me, "WUBC")
			if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			resolveTop(t, g)
			if got := counters(t, g, id, tc.kind); got != tc.want {
				t.Errorf("%s counters = %d, want %d", tc.kind, got, tc.want)
			}
		})
	}
}
