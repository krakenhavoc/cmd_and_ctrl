package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// landwalk_test.go — #705 and the ADR 0045 addendum's PR 1 surface:
// the game-aware pair check (BlockPairRefusalLocked), landwalk read
// from effective characteristics, the refusal error DeclareBlocker
// returns, and the read-only contract the enumerator depends on.

// pushWalker puts an untapped, unsick creature with printed keywords on
// the battlefield. Keywords ride Card.Keywords, the deck importer's
// road, so the layer engine keeps them through every recompute.
func pushWalker(g *Game, owner uuid.UUID, name string, keywords ...string) uuid.UUID {
	id := pushTypedTestCard(g, Card{
		Name: name, TypeLine: "Creature — Test", Power: 2, Toughness: 2,
		Owner: owner, Controller: owner, Keywords: keywords,
	})
	if c := findCard(g, id); c != nil {
		c.SummonedThisTurn = false
	}
	return id
}

func pushLand(g *Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
}

// pairRefusal runs the pair check the way the enumerator does: inside
// ReadSnapshot, with layers fresh.
func pairRefusal(g *Game, attacker, blocker uuid.UUID) BlockRefusal {
	var r BlockRefusal
	g.ReadSnapshot(func() {
		r = g.BlockPairRefusalLocked(findBattlefieldCard(g, attacker), findBattlefieldCard(g, blocker))
	})
	return r
}

// attackAndStepToBlocks declares every attacker at `target` and moves
// to declare blockers.
func attackAndStepToBlocks(t *testing.T, g *Game, target uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, target); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
}

// TestLandwalkForEachBasicLandType — CR 702.14c for the five basic
// land types: blockable while the defending player controls no land
// of the type, unblockable once they control one. A land of the type
// under the ATTACKER's control, or a land of another type, changes
// nothing.
func TestLandwalkForEachBasicLandType(t *testing.T) {
	for _, tc := range []struct{ token, basic, other string }{
		{"plainswalk", "Plains", "Island"},
		{"islandwalk", "Island", "Swamp"},
		{"swampwalk", "Swamp", "Mountain"},
		{"mountainwalk", "Mountain", "Forest"},
		{"forestwalk", "Forest", "Plains"},
	} {
		t.Run(tc.token, func(t *testing.T) {
			g := newActiveGame(t)
			me, them := g.Seats[0].ID, g.Seats[1].ID
			walker := pushWalker(g, me, "Walker", tc.token)
			blocker := pushWalker(g, them, "Blocker")
			pushLand(g, me, tc.basic, "Basic Land — "+tc.basic)
			pushLand(g, them, tc.other, "Basic Land — "+tc.other)
			attackAndStepToBlocks(t, g, them, walker)

			if r := pairRefusal(g, walker, blocker); !r.Legal() {
				t.Fatalf("no %s under the defender: refused with %q", tc.basic, r.Reason)
			}

			pushLand(g, them, tc.basic, "Basic Land — "+tc.basic)
			r := pairRefusal(g, walker, blocker)
			if r.Reason != BlockReasonLandwalk || r.Source != walker {
				t.Fatalf("defender controls a %s: refusal = %+v, want landwalk sourced at the walker", tc.basic, r)
			}
		})
	}
}

// TestNonbasicLandwalkReadsTheBasicSupertype — CR 205.4c: a land
// without the basic supertype is nonbasic, whatever its land types.
// A basic Island does not switch nonbasic landwalk on; a Tropical
// Island (Land — Island Forest) does.
func TestNonbasicLandwalkReadsTheBasicSupertype(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	walker := pushWalker(g, me, "Booted", "nonbasic landwalk")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Island", "Basic Land — Island")
	pushLand(g, them, "Snow-Covered Forest", "Basic Snow Land — Forest")
	attackAndStepToBlocks(t, g, them, walker)

	if r := pairRefusal(g, walker, blocker); !r.Legal() {
		t.Fatalf("only basic lands: refused with %q", r.Reason)
	}
	pushLand(g, them, "Tropical Island", "Land — Island Forest")
	if r := pairRefusal(g, walker, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("a nonbasic land: reason = %q, want landwalk", r.Reason)
	}
}

// TestLandwalkAbilitiesAreCheckedSeparately — CR 702.14d: two landwalk
// abilities don't cancel, and a blocker's own landwalk answers
// nothing. The first ability the defender's lands switch on is the
// one reported.
func TestLandwalkAbilitiesAreCheckedSeparately(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	walker := pushWalker(g, me, "Double Walker", "islandwalk", "swampwalk")
	blocker := pushWalker(g, them, "Swampwalking Blocker", "swampwalk")
	pushLand(g, them, "Swamp", "Basic Land — Swamp")
	attackAndStepToBlocks(t, g, them, walker)

	if r := pairRefusal(g, walker, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("defender controls a Swamp: reason = %q, want landwalk", r.Reason)
	}
	var kw string
	g.ReadSnapshot(func() { kw, _ = g.landwalkBlockingLandLocked(findBattlefieldCard(g, walker)) })
	if kw != "swampwalk" {
		t.Errorf("reported keyword = %q, want swampwalk (no Island, so islandwalk is not the one biting)", kw)
	}
}

// TestLandwalkReadsEffectiveLandTypes — the layer-4 half of Decision
// 10. A static that makes every land a Swamp in addition to its other
// types (Urborg's shape) switches swampwalk on against a defender
// holding only a Forest, and the land's type is read through the same
// post-layer view the mana abilities use.
func TestLandwalkReadsEffectiveLandTypes(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != "test-every-land-a-swamp" {
			return nil
		}
		return []StaticAbility{{
			Layer:     Layer4Type,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.IsLand() },
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				if !typeListHas(c.Subtypes, "Swamp") {
					c.Subtypes = append(c.Subtypes, "Swamp")
				}
			},
		}}
	})
	me, them := g.Seats[0].ID, g.Seats[1].ID
	walker := pushWalker(g, me, "Bog Walker", "swampwalk")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Forest", "Basic Land — Forest")
	attackAndStepToBlocks(t, g, them, walker)

	if r := pairRefusal(g, walker, blocker); !r.Legal() {
		t.Fatalf("a plain Forest: refused with %q", r.Reason)
	}
	pushTypedTestCard(g, Card{
		Name: "Urborg-alike", TypeLine: "Legendary Land", OracleID: "test-every-land-a-swamp",
		Owner: me, Controller: me,
	})
	if r := pairRefusal(g, walker, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("the Forest is a Swamp too now: reason = %q, want landwalk", r.Reason)
	}
}

// TestLandwalkReadsTheDefendingPlayerOfEachAttack — CR 506.2 / 509.1a.
// Against a planeswalker the defending player is its controller;
// against a battle it is the protector, not the battle's controller.
// Another opponent's Island is irrelevant.
func TestLandwalkReadsTheDefendingPlayerOfEachAttack(t *testing.T) {
	t.Run("planeswalker", func(t *testing.T) {
		g := newFourPlayerActiveGame(t)
		active := g.Seats[g.Turn.ActiveSeat].ID
		walkerOwner := g.Seats[(g.Turn.ActiveSeat+1)%4].ID
		bystander := g.Seats[(g.Turn.ActiveSeat+2)%4].ID
		pw := pushPlaneswalkerForTest(g, walkerOwner, "Their Walker", 5)
		fish := pushWalker(g, active, "Fish", "islandwalk")
		blocker := pushWalker(g, walkerOwner, "Blocker")
		pushLand(g, bystander, "Island", "Basic Land — Island")
		attackAndStepToBlocks(t, g, pw, fish)

		if r := pairRefusal(g, fish, blocker); !r.Legal() {
			t.Fatalf("only a bystander controls an Island: refused with %q", r.Reason)
		}
		pushLand(g, walkerOwner, "Island", "Basic Land — Island")
		if r := pairRefusal(g, fish, blocker); r.Reason != BlockReasonLandwalk {
			t.Fatalf("the planeswalker's controller controls an Island: reason = %q, want landwalk", r.Reason)
		}
	})
	t.Run("battle", func(t *testing.T) {
		g := newFourPlayerActiveGame(t)
		active := g.Seats[g.Turn.ActiveSeat].ID
		controller := g.Seats[(g.Turn.ActiveSeat+1)%4].ID
		protector := g.Seats[(g.Turn.ActiveSeat+2)%4].ID
		battle := pushBattleForTest(g, controller, protector, "Siege", 5)
		fish := pushWalker(g, active, "Fish", "islandwalk")
		blocker := pushWalker(g, protector, "Blocker")
		pushLand(g, controller, "Island", "Basic Land — Island")
		attackAndStepToBlocks(t, g, battle, fish)

		if r := pairRefusal(g, fish, blocker); !r.Legal() {
			t.Fatalf("only the battle's controller controls an Island: refused with %q", r.Reason)
		}
		pushLand(g, protector, "Island", "Basic Land — Island")
		if r := pairRefusal(g, fish, blocker); r.Reason != BlockReasonLandwalk {
			t.Fatalf("the protector controls an Island: reason = %q, want landwalk", r.Reason)
		}
	})
}

// TestDeclareBlockerRefusesLandwalkWithAReason is the engine end of
// the wire: the refusal is a *BlockRefusedError that still satisfies
// errors.Is(ErrIllegalBlock), nothing is stored or announced, the
// #328 signal agrees there is no block to make, and the sentence is
// addressed to whoever reads it.
func TestDeclareBlockerRefusesLandwalkWithAReason(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	fish := pushWalker(g, me.ID, "Cold-Eyed Selkie", "islandwalk")
	blocker := pushWalker(g, them.ID, "Grizzly Bears")
	pushLand(g, them.ID, "Island", "Basic Land — Island")
	attackAndStepToBlocks(t, g, them.ID, fish)

	if g.SeatOwesBlockDecision(them.ID) {
		t.Error("the #328 signal says the defender owes a block, but its only creature can't block the islandwalker")
	}

	seqBefore := len(g.Events)
	err := g.DeclareBlocker(blocker, fish)
	if !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("DeclareBlocker = %v, want an ErrIllegalBlock refusal", err)
	}
	var br *BlockRefusedError
	if !errors.As(err, &br) {
		t.Fatalf("DeclareBlocker = %T, want *BlockRefusedError", err)
	}
	if br.Reason != BlockReasonLandwalk || br.Blocker != blocker || br.Attacker != fish || br.Keyword != "islandwalk" {
		t.Errorf("refusal = %+v", br)
	}
	if c := findCard(g, blocker); c.BlockingTarget != uuid.Nil {
		t.Error("a refused block was stored")
	}
	for _, ev := range g.Events[seqBefore:] {
		if ev.Kind == EventBlock {
			t.Error("a refused block emitted EventBlock")
		}
	}

	if got, want := br.Sentence(them.ID), "Cold-Eyed Selkie has islandwalk, and you control an Island (Island)."; got != want {
		t.Errorf("defender's sentence = %q, want %q", got, want)
	}
	if got := br.Sentence(uuid.Nil); !strings.Contains(got, them.Name+" controls an Island") {
		t.Errorf("third party's sentence = %q, want it to name %s", got, them.Name)
	}
}

// TestBlockRefusalSentencesAreNeverEmpty walks every reason the engine
// can return, so a new reason can't ship without a sentence.
func TestBlockRefusalSentencesAreNeverEmpty(t *testing.T) {
	for _, reason := range BlockReasons() {
		e := &BlockRefusedError{BlockRefusal: BlockRefusal{Reason: reason}}
		if s := e.Sentence(uuid.Nil); s == "" || strings.Contains(s, "  ") {
			t.Errorf("%s: sentence %q", reason, s)
		}
		if !errors.Is(e, ErrIllegalBlock) {
			t.Errorf("%s: does not wrap ErrIllegalBlock", reason)
		}
	}
	for spec, want := range map[string]string{
		"islandwalk": "an Island", "plainswalk": "a Plains", "nonbasic landwalk": "a nonbasic land",
	} {
		s, ok := landwalkRequirement(spec)
		if !ok || s.describe() != want {
			t.Errorf("%s describes as %q, want %q", spec, s.describe(), want)
		}
	}
}

// TestLandwalkTokensAreCanonical keeps the landwalk table and the
// closed keyword table in step: every token the pair check reads is
// one the importer keeps, and the bare family name is not.
func TestLandwalkTokensAreCanonical(t *testing.T) {
	for _, lw := range landwalkTokens {
		if got, ok := CanonicalKeyword(strings.ToUpper(lw.token[:1]) + lw.token[1:]); !ok || got != lw.token {
			t.Errorf("%q is read by the block check but not canonical (got %q, %v)", lw.token, got, ok)
		}
	}
	if _, ok := CanonicalKeyword("Landwalk"); ok {
		t.Error(`"Landwalk" is Scryfall's family name, not an ability; it must not import`)
	}
	for kw := range canonicalKeywords {
		if strings.HasSuffix(kw, "walk") {
			if _, ok := landwalkRequirement(kw); !ok {
				t.Errorf("%q is canonical but the block check doesn't read it", kw)
			}
		}
	}
}

// TestBlockLegalityDoesNotMutate is the read-only contract of Decision
// 7. The enumerator and the #328 signal call the pair check from
// inside ReadSnapshot, where a write is a data race; so on a busy board
// every pair is checked, every refusal's error is built, and the
// signal is asked for every seat, and the encoded game must not move.
func TestBlockLegalityDoesNotMutate(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	withRestrictionOn(t, CantBeBlocked, "Cloaked")
	active := g.Seats[g.Turn.ActiveSeat].ID
	def := g.Seats[(g.Turn.ActiveSeat+1)%4].ID
	protector := g.Seats[(g.Turn.ActiveSeat+2)%4].ID
	pw := pushPlaneswalkerForTest(g, def, "Walker", 4)
	battle := pushBattleForTest(g, def, protector, "Siege", 4)
	attackers := []uuid.UUID{
		pushWalker(g, active, "Flier", "flying"),
		pushWalker(g, active, "Fish", "islandwalk"),
		pushWalker(g, active, "Booted", "nonbasic landwalk"),
		pushWalker(g, active, "Fear", "fear"),
		pushWalker(g, active, "Intimidator", "intimidate"),
		pushWalker(g, active, "Shadow", "shadow"),
		pushWalker(g, active, "Horse", "horsemanship"),
		pushWalker(g, active, "Skulk", "skulk"),
		pushWalker(g, active, "Plain"),
		pushRestrictableCreature(g, g.playerByIDLocked(active), "Cloaked"),
	}
	pushRestrictor(g, g.playerByIDLocked(active))
	var blockers []uuid.UUID
	for _, seat := range []uuid.UUID{def, protector} {
		blockers = append(blockers,
			pushWalker(g, seat, "Reach", "reach"),
			pushWalker(g, seat, "Ground"),
			pushWalker(g, seat, "Flier", "flying"),
		)
		pushLand(g, seat, "Island", "Basic Land — Island")
		pushLand(g, seat, "Tropical Island", "Land — Island Forest")
	}
	advanceIntoStep(t, g, StepDeclareAttackers)
	for i, a := range attackers {
		target := []uuid.UUID{def, pw, battle, def, def, def, def, def, def, def}[i]
		if err := g.DeclareAttacker(a, target); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	encode := func() []byte {
		snap := g.captureSnapshotLocked()
		snap.TakenAt = time.Time{}
		raw, err := json.Marshal(snap)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return raw
	}
	refused := 0
	g.ReadSnapshot(func() {
		before := encode()
		layers, resolved, events := g.layerVersion.Load(), g.lastResolvedVersion.Load(), len(g.Events)
		for _, a := range attackers {
			for _, b := range blockers {
				ac, bc := findBattlefieldCard(g, a), findBattlefieldCard(g, b)
				r := g.BlockPairRefusalLocked(ac, bc)
				if r.Legal() != g.CanBlockLocked(ac, bc) {
					t.Errorf("BlockPairRefusalLocked and CanBlockLocked disagree on %s / %s", ac.Name, bc.Name)
				}
				if !r.Legal() {
					refused++
					_ = g.blockRefusedErrorLocked(ac, bc, r).Sentence(def)
				}
			}
		}
		for _, p := range g.Seats {
			_ = g.seatOwesBlockDecisionLocked(p.ID)
		}
		if !bytes.Equal(before, encode()) {
			t.Error("the block-legality check changed the encoded game")
		}
		if g.layerVersion.Load() != layers || g.lastResolvedVersion.Load() != resolved || len(g.Events) != events {
			t.Error("the block-legality check touched the layer version or the event log")
		}
	})
	if refused == 0 {
		t.Fatal("no pair was refused, so the refusal path was never exercised")
	}
}
