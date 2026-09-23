package game

import (
	"testing"

	"github.com/google/uuid"
)

// player_protection_test.go — #1197, CR 702.11d and CR 702.16i: the
// RECEIVER of a protection-style keyword may be a player.
//
// Every test here is an assertion about one of the three choke points
// ADR 0072's 2026-09-22 amendment names, and they are deliberately
// written from the OTHER side of each: the opponent's spell that must
// be refused AND the player's own that must not, the damage that is
// prevented AND the cost payment that is not, the derived grant that
// composes AND the granted one that expires.

// withCatalogPlayerKeywords stubs the derived half of the reader — a
// battlefield permanent whose printed static gives its controller an
// ability. The game package cannot import cards/effects, which is why
// every catalog hook is a var and every test that needs one stubs it.
func withCatalogPlayerKeywords(t *testing.T, fn func(oracleID string) []string) {
	t.Helper()
	prev := CatalogPlayerKeywords
	CatalogPlayerKeywords = fn
	t.Cleanup(func() { CatalogPlayerKeywords = prev })
}

// anyTargetSpec is "any target" narrowed to players — Lightning
// Bolt's player half, which is the clause every targeting test below
// casts through.
func anyPlayerSpec() *TargetSpec {
	return &TargetSpec{
		Mode:    "player",
		Label:   "target player",
		Players: true,
		Min:     1, Max: 1,
	}
}

// pushLeyline seeds a battlefield permanent whose catalog key grants
// its controller `keywords`. Named for the card it stands in for.
func pushLeyline(t *testing.T, g *Game, controller *Player, oracle string) uuid.UUID {
	t.Helper()
	c := NewCard("Test Leyline", controller.ID)
	c.TypeLine = "Enchantment"
	c.OracleID = oracle
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: c.InstanceID, OldZone: ZoneHand, NewZone: ZoneBattlefield})
	})
	return c.InstanceID
}

// grantProtectionFromEverything is Teferi's Protection's clause,
// applied directly. Returns the duration it was stamped with.
func grantProtectionFromEverything(g *Game, p *Player) Duration {
	var d Duration
	g.WithWriteLock(func() {
		d = g.UntilYourNextTurnDuration(p.ID)
		g.GrantPlayerStaticForEffect(p.ID, "protection from everything",
			"Test — protection from everything", uuid.Nil, d)
	})
	return d
}

// castAtPlayer casts a one-shot instant at `ref`, with `spec` as its
// catalogued clause. Mirrors castColouredSpellAt in
// protection_debt_test.go; the colour is a parameter because
// protection reads the SOURCE's and hexproof does not care.
func castAtPlayer(t *testing.T, g *Game, caster *Player, oracle string, colors []string, target uuid.UUID) error {
	t.Helper()
	spell := NewCard("Test Bolt", caster.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	spell.Colors = colors
	caster.Hand.PushTop(spell)
	return g.CastSpell(caster.ID, spell.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: target}},
	})
}

// --- CR 702.11d: hexproof ------------------------------------------

// TestHexproofPlayerRefusesAnOpponentAndNotThemselves is the headline
// for Leyline of Sanctity. The asymmetry is the whole rule: hexproof
// asks WHO is casting, so your own spell still reaches you — which is
// load-bearing rather than a nicety, because a player who could not
// target themselves could not cast their own Sylvan Library.
func TestHexproofPlayerRefusesAnOpponentAndNotThemselves(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)

	const leyline = "test-player-hexproof-leyline"
	const bolt = "test-player-hexproof-bolt"
	withCatalogPlayerKeywords(t, func(id string) []string {
		if id == leyline {
			return []string{"hexproof"}
		}
		return nil
	})
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == bolt {
			return anyPlayerSpec()
		}
		return nil
	})

	pushLeyline(t, g, me, leyline)

	if err := castAtPlayer(t, g, opp, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("an opponent's spell at a hexproof player: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, me.ID); err != nil {
		t.Fatalf("your own spell at yourself must stay legal under hexproof: %v", err)
	}
	// And the Leyline protects only its controller: the opponent is
	// still an ordinary target for everybody.
	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, opp.ID); err != nil {
		t.Fatalf("a spell at the UNprotected seat: %v", err)
	}
}

// TestHexproofPlayerIsNotOfferedByTheEnumerator is the half that has
// to follow by construction: the bot's move list and the client's
// legal_targets both read legalTargetsLocked, so a seat the announce
// gate would refuse must never reach the picker in the first place.
// A refusal the UI cannot see coming is a bug even when the rule is
// right.
func TestHexproofPlayerIsNotOfferedByTheEnumerator(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	const leyline = "test-player-hexproof-enum"
	withCatalogPlayerKeywords(t, func(id string) []string {
		if id == leyline {
			return []string{"hexproof"}
		}
		return nil
	})
	pushLeyline(t, g, me, leyline)

	var mine, theirs LegalTargets
	g.ReadSnapshot(func() {
		mine = g.legalTargetsLocked(SourceChooser(me.ID), anyPlayerSpec())
		theirs = g.legalTargetsLocked(SourceChooser(opp.ID), anyPlayerSpec())
	})
	if !containsPlayer(mine.Players, me.ID) {
		t.Error("a hexproof player is missing from their OWN legal-target list")
	}
	if containsPlayer(theirs.Players, me.ID) {
		t.Error("a hexproof player is offered to an opponent's picker")
	}
	if !containsPlayer(theirs.Players, opp.ID) {
		t.Error("the enumeration lost the unprotected seat entirely")
	}
}

// TestHexproofGainedInResponseFizzlesTheSpell is CR 608.2b through the
// one function announce and resolution share. The player was a legal
// target when the spell was cast and is not when it resolves, so
// nothing else has to be taught the rule — the same property that
// makes a creature gaining hexproof mid-flight work.
func TestHexproofGainedInResponseFizzlesTheSpell(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)

	const leyline = "test-player-hexproof-recheck"
	const bolt = "test-player-hexproof-recheck-bolt"
	var leylineOut bool
	withCatalogPlayerKeywords(t, func(id string) []string {
		if id == leyline && leylineOut {
			return []string{"hexproof"}
		}
		return nil
	})
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == bolt {
			return anyPlayerSpec()
		}
		return nil
	})
	pushLeyline(t, g, opp, leyline)

	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, opp.ID); err != nil {
		t.Fatalf("cast at a bare player: %v", err)
	}
	// The Leyline "resolves" in response.
	leylineOut = true

	var stillLegal bool
	g.ReadSnapshot(func() {
		item := g.Stack.Cards[len(g.Stack.Cards)-1]
		stillLegal = g.targetLegalLocked(SourceObject(me.ID, &item), anyPlayerSpec(),
			TargetRef{Kind: TargetPlayer, ID: opp.ID})
	})
	if stillLegal {
		t.Error("CR 608.2b: a player who gained hexproof in response is still a legal target")
	}
}

// --- CR 702.16i: protection from everything ------------------------

// TestPlayerProtectionFromEverythingRefusesEveryone is the other
// asymmetry, and the one that separates protection from hexproof:
// protection asks WHAT is casting, not who, so "from everything"
// refuses the protected player's OWN spell too.
func TestPlayerProtectionFromEverythingRefusesEveryone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)

	const bolt = "test-player-protection-bolt"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == bolt {
			return anyPlayerSpec()
		}
		return nil
	})
	grantProtectionFromEverything(g, me)

	if err := castAtPlayer(t, g, opp, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("an opponent's spell at a protected player: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, me, bolt, []string{"W"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("your OWN spell at yourself under protection from everything: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, opp.ID); err != nil {
		t.Fatalf("a spell at the unprotected seat: %v", err)
	}
}

// TestPlayerProtectionPreventsNoncombatDamageFromAnySource — CR
// 702.16e through the effect entry point. "From everything" names no
// characteristic to look up (CR 702.16j), so the colour of the source
// is irrelevant and a colourless one is refused too; the unprotected
// seat is the control that fails if the check degenerates into
// "prevent everything".
func TestPlayerProtectionPreventsNoncombatDamageFromAnySource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantProtectionFromEverything(g, me)

	startMe, startOpp := me.Life, opp.Life
	bolt := NewCard("Lightning Bolt", opp.ID)
	bolt.TypeLine = "Instant"
	bolt.Colors = []string{"R"}
	rock := NewCard("Colourless Pinger", opp.ID)
	rock.TypeLine = "Artifact"
	g.WithWriteLock(func() {
		g.Stack.PushTop(bolt)
		g.Battlefield.PushTop(rock)
		if err := g.DealDamageToPlayerForEffect(bolt.InstanceID, me.ID, 3); err != nil {
			t.Fatalf("red spell at the protected player: %v", err)
		}
		if err := g.DealDamageToPlayerForEffect(rock.InstanceID, me.ID, 2); err != nil {
			t.Fatalf("colourless permanent's ability at the protected player: %v", err)
		}
		if err := g.DealDamageToPlayerForEffect(bolt.InstanceID, opp.ID, 3); err != nil {
			t.Fatalf("red spell at the unprotected player: %v", err)
		}
	})

	if me.Life != startMe {
		t.Errorf("protected player lost %d life; CR 702.16e prevents damage from EVERY source", startMe-me.Life)
	}
	if opp.Life != startOpp-3 {
		t.Errorf("unprotected player is at %d, want %d — the built-in is preventing damage it should not", opp.Life, startOpp-3)
	}
}

// TestPlayerProtectionPreventsCombatDamage is the same rule through
// the OTHER entry point. It matters that both are tested and it costs
// almost nothing that they are: every damage path in the engine
// routes through damageThroughReplacementsLocked, so if one works and
// the other does not, the branch was written in the wrong place.
func TestPlayerProtectionPreventsCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantProtectionFromEverything(g, me)

	attacker := pushColouredCreature(g, opp, "Big Attacker", []string{"G"})
	start := me.Life
	g.WithWriteLock(func() {
		g.markCombatDamageToPlayerLocked(me.ID, attacker, 8, "")
	})
	if me.Life != start {
		t.Errorf("protected player took %d combat damage; CR 702.16e prevents it", start-me.Life)
	}
}

// TestPlayerProtectionEndsAsYourNextTurnBegins pins the duration
// against the one function that decides when any continuous effect in
// the game is over. The grant must survive the opponent's whole turn
// and be gone the instant the granting player's next turn begins
// (CR 500.1) — the boundary ADR 0063 Decision 3 exists for.
func TestPlayerProtectionEndsAsYourNextTurnBegins(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	grantProtectionFromEverything(g, me)

	if !playerHasAbility(g, me, "protection from everything") {
		t.Fatal("the grant did not take at all")
	}
	advanceOneTurn(t, g) // the opponent's turn
	if !playerHasAbility(g, me, "protection from everything") {
		t.Fatal("the grant ended during the opponent's turn — that is the turn it exists to survive")
	}
	advanceOneTurn(t, g) // back to me
	if playerHasAbility(g, me, "protection from everything") {
		t.Error("the grant survived the beginning of its own player's next turn")
	}
	// The sweep, not just the reader: a grant the reader refuses but
	// the slice still holds would leak into every snapshot forever.
	if n := len(me.Statics); n != 0 {
		t.Errorf("Player.Statics still holds %d expired entries; the sweep did not run", n)
	}
}

// TestExpiredPlayerStaticIsRefusedBeforeTheSweepRuns is the reason the
// reader tests the duration as well as the sweep: the sweep runs at
// known moments and the reader has to be right BETWEEN them. Stamped
// by hand into the past rather than by advancing turns, which is what
// makes the window observable at all.
func TestExpiredPlayerStaticIsRefusedBeforeTheSweepRuns(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(me.ID, "hexproof", "stale", uuid.Nil,
			Duration{Kind: UntilYourNextTurn, Player: me.ID, ExpiresAtTurnsBegun: 1})
	})
	if len(me.Statics) != 1 {
		t.Fatal("the stale grant was not stored; the test is vacuous")
	}
	if playerHasAbility(g, me, "hexproof") {
		t.Error("an expired grant still answers; the reader is trusting the sweep")
	}
}

// --- CR 702.16c: the attachment half -------------------------------

// TestPlayerProtectionSendsAnEnchantPlayerAuraToTheGraveyard —
// CR 702.16c for the one attachment kind a player can have. Curse of
// Opulence is the card; the state-based action is the same
// CR 704.5m that drops an Aura off a protected creature.
func TestPlayerProtectionSendsAnEnchantPlayerAuraToTheGraveyard(t *testing.T) {
	// Four seats: eliminating or protecting one of two would end the
	// game, and stateBasedActionsLocked is a no-op once it has.
	g := newFourPlayerActiveGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	curse := pushAttachTestCard(g, me.ID, "Test Curse", "Enchantment — Aura Curse")

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(curse, TargetRef{Kind: TargetPlayer, ID: foe.ID}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		g.runStateChecksLocked()
	})
	if got, ok := battlefieldCardByID(g, curse); !ok || !got.IsAttached() {
		t.Fatalf("the Curse fell off before the protection existed (onBattlefield=%v)", ok)
	}

	grantProtectionFromEverything(g, foe)
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if _, ok := battlefieldCardByID(g, curse); ok {
		t.Error("CR 702.16c: an Aura on a player with protection from everything must fall off")
	}
}

// --- the non-targeting half ----------------------------------------

// TestChoosingAProtectedPlayerForACostIsStillLegal is the boundary
// ADR 0038 §4 drew and this seam inherits: paying a cost, or a
// "choose a player" that does not say "target", is not targeting
// (CR 601.2f/h). specCandidatesLocked must still offer the seat, or
// a hexproof player could not be chosen by their own Coercive
// Portal — and neither could anybody else be, by a spell that merely
// chooses.
func TestChoosingAProtectedPlayerForACostIsStillLegal(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantProtectionFromEverything(g, me)

	var chooseable LegalTargets
	g.ReadSnapshot(func() {
		chooseable = g.specCandidatesLocked(opp.ID, anyPlayerSpec())
	})
	if !containsPlayer(chooseable.Players, me.ID) {
		t.Error("a protected player is not CHOOSABLE; the keyword gate leaked into the non-targeting walk")
	}
}

// --- composition, and the two stores -------------------------------

// TestTwoDerivedGrantsComposeAndOneLeavingDoesNotRevoke is the whole
// argument for deriving rather than writing, in one test: two
// Leylines, one leaves, the player is still hexproof. A "set on
// enter, restore on leave" design gets exactly this case wrong, which
// is the case CatalogNoMaxHandSize's comment describes at length.
func TestTwoDerivedGrantsComposeAndOneLeavingDoesNotRevoke(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const leyline = "test-player-hexproof-compose"
	withCatalogPlayerKeywords(t, func(id string) []string {
		if id == leyline {
			return []string{"hexproof"}
		}
		return nil
	})
	first := pushLeyline(t, g, me, leyline)
	pushLeyline(t, g, me, leyline)

	if !playerHasAbility(g, me, "hexproof") {
		t.Fatal("two Leylines and no hexproof")
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, first); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if !playerHasAbility(g, me, "hexproof") {
		t.Error("one Leyline leaving revoked the other's grant")
	}
}

// TestGrantedPlayerStaticSurvivesCloneAndUndo — the granted half is
// state, so it has to rewind with everything else. A grant made and
// then undone must be gone; a grant made BEFORE the snapshot must
// come back.
func TestGrantedPlayerStaticSurvivesCloneAndUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	grantProtectionFromEverything(g, me)

	var snap *Game
	g.ReadSnapshot(func() { snap = g.cloneLocked() })

	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(me.ID, "hexproof", "second grant", uuid.Nil, IndefiniteDuration())
	})
	if n := len(me.Statics); n != 2 {
		t.Fatalf("Player.Statics has %d entries after the second grant, want 2", n)
	}
	// The clone was taken before the second grant and must not have
	// grown one: the sweep replaces the slice rather than compacting
	// it precisely so this holds.
	if n := len(snap.Seats[0].Statics); n != 1 {
		t.Errorf("the clone holds %d entries, want 1 — the backing array is shared", n)
	}
	if snap.Seats[0].Statics[0].Keyword != "protection from everything" {
		t.Errorf("the clone's entry is %q", snap.Seats[0].Statics[0].Keyword)
	}
}

// --- helpers -------------------------------------------------------

func containsPlayer(ids []uuid.UUID, want uuid.UUID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func playerHasAbility(g *Game, p *Player, keyword string) bool {
	var found bool
	g.ReadSnapshot(func() { found = g.PlayerHasKeywordLocked(p, keyword) })
	return found
}
