package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// unpreventable_damage_test.go — ADR 0107 §5 (#1853): CR 615.12's gate
// and its four sources, and the redirection ban beside it.

const unpreventableOracle = "test-unpreventable-static"

// withUnpreventableStatics stubs the catalog so a battlefield permanent
// with unpreventableOracle has exactly these statics.
func withUnpreventableStatics(t *testing.T, statics ...UnpreventableDamageStatic) {
	t.Helper()
	prev := CatalogUnpreventableDamage
	CatalogUnpreventableDamage = func(key string) []UnpreventableDamageStatic {
		if key == unpreventableOracle {
			return statics
		}
		return nil
	}
	t.Cleanup(func() { CatalogUnpreventableDamage = prev })
}

// pushUnpreventableCreature puts a red creature carrying the stubbed
// statics onto the battlefield.
func pushUnpreventableCreature(g *Game, owner *Player) uuid.UUID {
	c := NewCard("Unpreventable", owner.ID)
	c.TypeLine = "Creature — Avatar"
	c.Power, c.Toughness = 4, 4
	c.Colors = []string{"R"}
	c.OracleID = unpreventableOracle
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// pushRedSpell puts a red instant on the stack and returns it, to be a
// damage source.
func pushRedSpell(g *Game, owner *Player) uuid.UUID {
	bolt := NewCard("Test Bolt", owner.ID)
	bolt.TypeLine = "Instant"
	bolt.Colors = []string{"R"}
	g.WithWriteLock(func() { g.Stack.PushTop(bolt) })
	return bolt.InstanceID
}

// A spell's own "the damage can't be prevented" gets through protection
// (CR 702.16e is a prevention effect, so CR 615.12 applies), and the
// same spell's unmarked damage does not.
func TestMarkedDamageGetsThroughProtection(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushColouredCreature(g, opp, "Pro-Red Bear", []string{"W"}, "protection from red")
	bolt := pushRedSpell(g, me)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(bolt, victim, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DealMarkedDamageForEffect(bolt, nil, victim, 3, DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, victim); got != 3 {
		t.Errorf("protected creature has %d damage, want 3 (the unmarked 1 prevented, the marked 3 dealt)", got)
	}
}

// CR 615.12: "Existing damage prevention shields won't be reduced by
// damage that can't be prevented." The shield is applied, prevents
// nothing, and keeps its whole charge for the next, preventable event.
func TestUnpreventableDamageDoesNotSpendAShield(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushColouredCreature(g, opp, "Shielded Bear", []string{"G"})
	bolt := pushRedSpell(g, me)

	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, victim, 3, false, "shield") {
			t.Fatal("no shield")
		}
		if err := g.DealMarkedDamageForEffect(bolt, nil, victim, 2, DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, victim); got != 2 {
		t.Fatalf("victim has %d damage, want 2 — the shield must not prevent unpreventable damage", got)
	}
	g.WithWriteLock(func() {
		if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Mods[0].Amount != 3 {
			t.Fatalf("shield after unpreventable damage = %+v, want one record with all 3 charges", g.ScopedEffects)
		}
		if err := g.DealDamageToCreatureForEffect(bolt, victim, 3); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, victim); got != 2 {
		t.Errorf("victim has %d damage after a preventable 3, want still 2 — the untouched shield takes it", got)
	}
}

// "Damage can't be prevented this turn" (Skullcrack) is a rule grant:
// it covers combat damage under a Fog and damage from a source that did
// not exist when it began (CR 611.2c), and it ends at cleanup.
func TestDamageCantBePreventedThisTurnBeatsFogUntilCleanup(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocker := pushColouredCreature(g, opp, "Blocker", []string{"G"})

	g.WithWriteLock(func() {
		g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog")
		g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Skullcrack")
	})
	attacker := pushColouredCreature(g, me, "Late Attacker", []string{"R"})
	g.WithWriteLock(func() { g.markCombatDamageOnCardLocked(blocker, 2, attacker, CombatStepRegular) })
	if got := damageOn(g, blocker); got != 2 {
		t.Fatalf("blocker has %d damage under Fog + Skullcrack, want 2", got)
	}
	g.WithWriteLock(func() {
		g.sweepScopedEffectsLocked(true)
		if len(g.DamageCantBePreventedThisTurnLabels()) != 0 {
			t.Fatal("the turn grant outlived cleanup (CR 514.2)")
		}
	})
}

// The battlefield statics, each against the damage it does and does not
// cover.
func TestUnpreventableStaticsCoverWhatTheyName(t *testing.T) {
	cases := []struct {
		name            string
		scope           UnpreventableDamageScope
		ownCombat       bool // my creature's combat damage gets through
		oppCombat       bool // my opponent's creature's combat damage gets through
		spellNoncombat  bool // a spell's damage gets through
		staticNoncombat bool // the static's own permanent's noncombat damage gets through
	}{
		{"all damage (Leyline of Punishment)", UnpreventableAll, true, true, true, true},
		{"combat damage (Frenzied Baloth)", UnpreventableCombat, true, true, false, false},
		{"combat damage by your creatures (Questing Beast)", UnpreventableCombatByYourCreatures, true, false, false, false},
		{"damage by this creature (Excruciator)", UnpreventableByThis, false, false, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withUnpreventableStatics(t, UnpreventableDamageStatic{Label: tc.name, Scope: tc.scope})
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			static := pushUnpreventableCreature(g, me)
			mine := pushColouredCreature(g, me, "My Red", []string{"R"})
			theirs := pushColouredCreature(g, opp, "Their Red", []string{"R"})
			mineVictim := pushColouredCreature(g, opp, "Pro-Red A", []string{"W"}, "protection from red")
			theirVictim := pushColouredCreature(g, me, "Pro-Red B", []string{"W"}, "protection from red")
			spellVictim := pushColouredCreature(g, opp, "Pro-Red C", []string{"W"}, "protection from red")
			staticVictim := pushColouredCreature(g, opp, "Pro-Red D", []string{"W"}, "protection from red")
			bolt := pushRedSpell(g, me)

			g.WithWriteLock(func() {
				g.markCombatDamageOnCardLocked(mineVictim, 1, mine, CombatStepRegular)
				g.markCombatDamageOnCardLocked(theirVictim, 1, theirs, CombatStepRegular)
				if err := g.DealDamageToCreatureForEffect(bolt, spellVictim, 1); err != nil {
					t.Fatal(err)
				}
				if err := g.DealDamageToCreatureForEffect(static, staticVictim, 1); err != nil {
					t.Fatal(err)
				}
			})
			check := func(what string, victim uuid.UUID, want bool) {
				t.Helper()
				if got := damageOn(g, victim) == 1; got != want {
					t.Errorf("%s got through protection = %v, want %v", what, got, want)
				}
			}
			check("my creature's combat damage", mineVictim, tc.ownCombat)
			check("an opponent's creature's combat damage", theirVictim, tc.oppCombat)
			check("a spell's damage", spellVictim, tc.spellNoncombat)
			check("the static's own permanent's noncombat damage", staticVictim, tc.staticNoncombat)
		})
	}
}

// "Damage that would be dealt by this creature" is read from last-known
// information once the source has left (CR 608.2h): a dies trigger's
// damage, or a paused event's.
func TestSourceStaticIsReadFromLastKnownInformation(t *testing.T) {
	withUnpreventableStatics(t, UnpreventableDamageStatic{Scope: UnpreventableByThis})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushUnpreventableCreature(g, me)
	victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(src); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(src, victim, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, victim); got != 2 {
		t.Errorf("victim has %d damage from the departed source, want 2 — its static is last-known information", got)
	}
}

// The pinned grant covers damage dealt TO that creature and nothing
// else.
func TestPinnedGrantCoversOnlyThatCreature(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pinned := pushColouredCreature(g, opp, "Pinned", []string{"W"}, "protection from red")
	other := pushColouredCreature(g, opp, "Other", []string{"W"}, "protection from red")
	bolt := pushRedSpell(g, me)
	g.WithWriteLock(func() {
		if !g.DamageToCantBePreventedThisTurnForEffect(uuid.Nil, pinned, true, "Whippoorwill") {
			t.Fatal("no grant")
		}
		_ = g.DealDamageToCreatureForEffect(bolt, pinned, 2)
		_ = g.DealDamageToCreatureForEffect(bolt, other, 2)
	})
	if got := damageOn(g, pinned); got != 2 {
		t.Errorf("pinned creature has %d damage, want 2", got)
	}
	if got := damageOn(g, other); got != 0 {
		t.Errorf("other creature has %d damage, want 0 — the grant names one creature", got)
	}
}

// A replacement that is not a prevention effect is untouched by "can't
// be prevented", and an inert shield is never an ordering question: a
// doubler and a Fog under Skullcrack ask nothing and double.
func TestUnpreventableDamageStillDoublesWithoutAPrompt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocker := pushColouredCreature(g, opp, "Blocker", []string{"G"})
	attacker := pushColouredCreature(g, me, "Attacker", []string{"R"})
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventDealDamage},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventDamage && ev.DamageAmount > 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Label: "doubler",
		})
		g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog")
		g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Skullcrack")
		g.markCombatDamageOnCardLocked(blocker, 1, attacker, CombatStepRegular)
	})
	if n := len(g.PendingChoices); n != 0 {
		t.Fatalf("%d pending choices, want none — an inert Fog is not an ordering question", n)
	}
	if got := damageOn(g, blocker); got != 2 {
		t.Errorf("blocker has %d damage, want 2 (doubled, not prevented)", got)
	}
}

// A redirection under "can't be dealt instead to another permanent or
// player" does nothing; the same redirection applies to ordinary damage.
func TestRedirectionBanStopsARedirection(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushColouredCreature(g, opp, "Victim", []string{"G"})
	decoy := pushColouredCreature(g, opp, "Decoy", []string{"G"})
	bolt := pushRedSpell(g, me)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:         []EventKind{EventDealDamage},
			RedirectsDamage: true,
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventDamage && ev.DamageTarget == victim
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.DamageTarget = decoy
				return nil
			},
			Label: "redirect to the decoy",
		})
		_ = g.DealMarkedDamageForEffect(bolt, nil, victim, 2, DamageMarks{CantBePrevented: true, CantBeRedirected: true})
	})
	if got := damageOn(g, victim); got != 2 {
		t.Errorf("victim has %d damage, want 2 — the redirection is banned", got)
	}
	if got := damageOn(g, decoy); got != 0 {
		t.Errorf("decoy has %d damage, want 0", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bolt, victim, 1) })
	if got := damageOn(g, decoy); got != 1 {
		t.Errorf("decoy has %d damage after ordinary damage, want 1 — the redirection works when not banned", got)
	}
}

// Both turn grants are data: a table holding one is a restore point, and
// the restored game still honours them.
func TestDamageAndLifeGrantsAreRestorePoints(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	g.WithWriteLock(func() {
		g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Skullcrack")
		g.PlayersCantGainLifeForEffect(uuid.Nil, me.ID, CantGainLifeEveryone, g.UntilEndOfTurnDuration(), "Skullcrack")
		g.PlayerCantGainLifeForEffect(uuid.Nil, opp.ID, IndefiniteDuration(), "Screaming Nemesis")
	})
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	bolt := pushRedSpell(restored, restored.Seats[0])
	restored.WithWriteLock(func() {
		_ = restored.DealDamageToCreatureForEffect(bolt, victim, 2)
		_ = restored.ChangePlayerLifeForEffect(uuid.Nil, restored.Seats[0].ID, 5)
		restored.sweepScopedEffectsLocked(true)
		_ = restored.ChangePlayerLifeForEffect(uuid.Nil, restored.Seats[1].ID, 5)
	})
	if got := damageOn(restored, victim); got != 2 {
		t.Errorf("restored game: victim has %d damage, want 2", got)
	}
	if got := restored.Seats[0].Life; got != me.Life {
		t.Errorf("restored game: player gained life under Skullcrack, life %d want %d", got, me.Life)
	}
	if got := restored.Seats[1].Life; got != opp.Life {
		t.Errorf("restored game: the rest-of-the-game grant did not survive cleanup, life %d want %d", got, opp.Life)
	}
}
