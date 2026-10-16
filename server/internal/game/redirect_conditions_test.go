package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// redirect_conditions_test.go — ADR 0108 §9, amendment 2026-10-16
// (#2056): a redirection whose source is an unblocked creature as the
// damage would be dealt (#2026's source filter), whose protected set is
// read as the damage would be dealt (#2045's recipient filter), and whose
// "you may" its controller answers each time (ReplacementEffect.Optional).

var yourCreatures = DamageRecipientFilter{Permanents: true, Match: []PermanentQuery{creatureQuery}, Controller: RecipientControllerYou}

// "All combat damage that would be dealt to you by unblocked creatures"
// (Kjeldoran Royal Guard): a blocked attacker's damage stays where it was,
// an unblocked attacker's is dealt to the guard (CR 509.1h).
func TestRedirectFromUnblockedCreatures(t *testing.T) {
	t.Run("blocked", func(t *testing.T) {
		g := newActiveGame(t)
		def := g.Seats[1]
		attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
		chump := pushCombatant(t, g, def, "Chump", 1, 1)
		guard := pushCombatant(t, g, def, "Guard", 2, 5)
		blockAfterLockIn(t, g, attacker, chump)
		redirect(t, g, DamageRedirection{Controller: def.ID, ProtectPlayer: def.ID, Queries: []PermanentQuery{creatureQuery},
			SourceFilter: DamageSourceFilter{Combat: SourceCombatUnblocked}, To: guard}, uuid.Nil)
		if lost := hits(t, g, def.ID, 2, attacker); lost != 2 || markedOn(g, guard) != 0 {
			t.Fatalf("lost %d (want 2), guard %d (want 0): a blocked creature is not unblocked", lost, markedOn(g, guard))
		}
	})
	t.Run("unblocked", func(t *testing.T) {
		g := newActiveGame(t)
		def := g.Seats[1]
		attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
		other := pushColouredCreature(g, g.Seats[0], "Pinger", []string{"R"})
		guard := pushCombatant(t, g, def, "Guard", 2, 5)
		declareAttacks(t, g, attacker)
		g.WithWriteLock(func() {
			g.completeAllBlockDeclarationsLocked()
			g.commitBlockDeclarationLocked()
		})
		redirect(t, g, DamageRedirection{Controller: def.ID, ProtectPlayer: def.ID, Queries: []PermanentQuery{creatureQuery},
			SourceFilter: DamageSourceFilter{Combat: SourceCombatUnblocked}, To: guard}, uuid.Nil)
		if lost := hits(t, g, def.ID, 2, attacker); lost != 0 || markedOn(g, guard) != 2 {
			t.Fatalf("lost %d (want 0), guard %d (want 2): the unblocked attacker's damage is the guard's", lost, markedOn(g, guard))
		}
		if lost := hits(t, g, def.ID, 1, other); lost != 1 {
			t.Fatalf("lost %d, want 1: a creature that is not attacking is not unblocked", lost)
		}
		var lines []string
		g.WithWriteLock(func() { lines = g.DamageRedirectionLines() })
		if len(lines) != 1 || !strings.Contains(lines[0], "from unblocked creatures") {
			t.Fatalf("banner %q, want the unblocked source named", lines)
		}
	})
}

// "All damage that would be dealt to creatures you control is dealt to
// that creature instead" (Ascent of the Worthy): your creatures, one that
// arrives later included (CR 611.2c), never you, never an opponent's
// creature; damage to the chosen creature stays on it.
func TestRedirectFromCreaturesYouControlSparesYou(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushArtifactSource(g, opp, "Rod")
	chosen := pushCombatant(t, g, me, "Chosen", 2, 9)
	other := pushCombatant(t, g, me, "Other", 2, 5)
	theirs := pushCombatant(t, g, opp, "Theirs", 2, 5)
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectSet: yourCreatures, To: chosen, UntilYourNextTurn: true}, uuid.Nil)
	late := pushCombatant(t, g, me, "Late", 2, 5)
	freshLayers(g)
	if got := strike(t, g, src, other, late, theirs, chosen); got[0] != 0 || got[1] != 0 || got[2] != 1 || got[3] != 3 {
		t.Fatalf("marked %v, want [0 0 1 3]: your creatures' damage is the chosen one's", got)
	}
	if lost := hits(t, g, me.ID, 2, src); lost != 2 {
		t.Fatalf("lost %d, want 2: you are not one of your creatures", lost)
	}
	var lines []string
	g.WithWriteLock(func() { lines = g.DamageRedirectionLines() })
	if len(lines) != 1 || !strings.Contains(lines[0], "to creatures "+me.Name+" controls") {
		t.Fatalf("banner %q, want the set named", lines)
	}
}

// optionalPrompt is the open optional_replacement prompt, or nil.
func optionalPrompt(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceOptionalReplacement {
			return c
		}
	}
	return nil
}

// "If damage would be dealt to any creature, you may have that damage
// dealt to you instead" (Blood of the Martyr): asked of the effect's
// controller, not of the creature's, once per event; yes moves it, no
// leaves it.
func TestOptionalRedirectionAsksItsController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushArtifactSource(g, opp, "Rod")
	theirs := pushCombatant(t, g, opp, "Theirs", 2, 5)
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectSet: everyCreature, To: me.ID, Optional: true}, uuid.Nil)
	freshLayers(g)
	start := lifeOf(g, me.ID)

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, theirs, 2) })
	p := optionalPrompt(g)
	if p == nil || p.Chooser != me.ID {
		t.Fatalf("want a may prompt for the effect's controller, got %+v", p)
	}
	if !strings.Contains(p.Reason, "have the 2 damage from Rod to Theirs dealt to you instead") {
		t.Errorf("reason %q", p.Reason)
	}
	var amount int
	var from, to uuid.UUID
	g.WithWriteLock(func() { amount, from, to = g.DamageRedirectOfferLocked(p) })
	if amount != 2 || from != theirs || to != me.ID {
		t.Errorf("offer (%d, %v, %v), want (2, the creature, me)", amount, from, to)
	}
	if err := g.ResolveOptionalReplacement(p.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if lifeOf(g, me.ID) != start-2 || markedOn(g, theirs) != 0 {
		t.Fatalf("yes: life %d (want %d), creature %d (want 0)", lifeOf(g, me.ID), start-2, markedOn(g, theirs))
	}

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, theirs, 3) })
	p = optionalPrompt(g)
	if p == nil {
		t.Fatal("the may is asked again for the next event")
	}
	if err := g.ResolveOptionalReplacement(p.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if lifeOf(g, me.ID) != start-2 || markedOn(g, theirs) != 3 {
		t.Fatalf("no: life %d (want %d), creature %d (want 3)", lifeOf(g, me.ID), start-2, markedOn(g, theirs))
	}
	if hits(t, g, opp.ID, 1, src) != 1 || optionalPrompt(g) != nil {
		t.Fatal("damage to a player is not damage to a creature, and is not asked about")
	}
}

// The three fields come back from a restore point and still read the
// same.
func TestRedirectConditionsSurviveARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushArtifactSource(g, opp, "Rod")
	chosen := pushCombatant(t, g, me, "Chosen", 2, 9)
	other := pushCombatant(t, g, me, "Other", 2, 5)
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectSet: yourCreatures, To: chosen, Optional: true}, uuid.Nil)
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Queries: []PermanentQuery{creatureQuery},
		SourceFilter: DamageSourceFilter{Combat: SourceCombatUnblocked}, To: chosen}, uuid.Nil)
	_, restored := roundTrip(t, g)
	var mods []Mod
	for _, e := range restored.ScopedEffects {
		mods = append(mods, e.Mods...)
	}
	if len(mods) != 2 || !mods[0].Optional || len(mods[0].RecipientFilter) != 1 || len(mods[1].SourceFilter) != 1 {
		t.Fatalf("restored mods %+v", mods)
	}
	freshLayers(restored)
	restored.WithWriteLock(func() { _ = restored.DealDamageToCreatureForEffect(src, other, 2) })
	p := optionalPrompt(restored)
	if p == nil || p.Chooser != me.ID {
		t.Fatalf("restored: want the may prompt, got %+v", p)
	}
	if err := restored.ResolveOptionalReplacement(p.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if markedOn(restored, chosen) != 2 || markedOn(restored, other) != 0 {
		t.Fatalf("restored: chosen %d (want 2), other %d (want 0)", markedOn(restored, chosen), markedOn(restored, other))
	}
}

// What registration and restore refuse: optional off the kind, without a
// destination or with a charge; a source filter beside a pinned source;
// a recipient set beside a player, the recipients vocabulary or a charge.
func TestRedirectConditionModProblems(t *testing.T) {
	src := ObjectRef{ID: uuid.New(), Epoch: 1}
	to := []ObjectRef{{ID: uuid.New()}}
	unblocked := []DamageSourceFilter{{Combat: SourceCombatUnblocked}}
	set := []DamageRecipientFilter{everyCreature}
	check := func(m Mod) string {
		for _, f := range []func(Mod) string{redirectDamageModProblem, nextFromSourceModProblem} {
			if p := f(m); p != "" {
				return p
			}
		}
		return ""
	}
	for _, c := range []struct {
		m    Mod
		want bool
	}{
		{Mod{Kind: ModRedirectDamage, To: to, Optional: true, RecipientFilter: set}, false},
		{Mod{Kind: ModRedirectDamage, To: to, Queries: []PermanentQuery{creatureQuery}, SourceFilter: unblocked, CombatOnly: true}, false},
		{Mod{Kind: ModRedirectDamage, To: to, RecipientFilter: []DamageRecipientFilter{yourCreatures}}, false},
		{Mod{Kind: ModPreventFromSource, Optional: true}, true},
		{Mod{Kind: ModMultiplyDamage, Amount: 2, Optional: true}, true},
		{Mod{Kind: ModRedirectDamage, To: to, Optional: true, Amount: 2}, true},
		{Mod{Kind: ModRedirectDamage, To: to, Objects: []ObjectRef{src}, SourceZone: ZoneBattlefield, SourceFilter: unblocked}, true},
		{Mod{Kind: ModRedirectDamage, To: to, SourceFilter: []DamageSourceFilter{{}}}, true},
		{Mod{Kind: ModRedirectDamage, To: to, SourceFilter: []DamageSourceFilter{unblocked[0], unblocked[0]}}, true},
		{Mod{Kind: ModRedirectDamage, To: to, Player: uuid.New(), RecipientFilter: set}, true},
		{Mod{Kind: ModRedirectDamage, To: to, Recipients: DamageRecipientsOpponents, RecipientFilter: set}, true},
		{Mod{Kind: ModRedirectDamage, To: to, Amount: 2, RecipientFilter: set}, true},
		{Mod{Kind: ModRedirectDamage, To: to, RecipientFilter: []DamageRecipientFilter{{}}}, true},
	} {
		if got := check(c.m) != ""; got != c.want {
			t.Errorf("%+v: problem %v, want %v (%q)", c.m, got, c.want, check(c.m))
		}
	}
	if p := recipientFilterScopeProblem(ScopeNone, []Mod{{Kind: ModRedirectDamage, To: to, RecipientFilter: set}}); p == "" {
		t.Error("a pinned redirection with a recipient set passed the check")
	}
}

// A redirection pinned to a permanent refuses a recipient set beside it:
// nothing is registered.
func TestRedirectRefusesASetBesideAPinnedPermanent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := pushCombatant(t, g, me, "A", 2, 2)
	b := pushCombatant(t, g, me, "B", 2, 2)
	g.WithWriteLock(func() {
		if g.RedirectDamageThisTurnForEffect(DamageRedirection{Controller: me.ID, ProtectPermanent: a, ProtectSet: everyCreature, To: b}) {
			t.Fatal("registered")
		}
	})
}
