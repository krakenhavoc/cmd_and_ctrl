package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seam2027_this_combat_test.go — #2027, ADR 0108's amendment of
// 2026-10-07: effects that last "this combat", "until end of combat" or
// "for as long as this Saga remains on the battlefield".

const (
	s2027Glyph    = "cca68900-2093-4cae-b757-efc7dd807e18"
	s2027OldFat   = "e4402d21-bb7f-4777-bc8d-b24bf2faccb0"
	s2027Sewers   = "2482f98a-eadf-450d-9e73-fc441b572862"
	s2027Skyguard = "06c47664-2b4b-468e-a38d-04b62e6edd67"
)

// s2027Cast casts `name` as `who` where the game stands and settles the
// stack.
func s2027Cast(t *testing.T, g *game.Game, who uuid.UUID, name, typeLine, oracle string, targets ...uuid.UUID) {
	t.Helper()
	id := uuid.New()
	var seat *game.Player
	for _, p := range g.Seats {
		if p.ID == who {
			seat = p
		}
	}
	seat.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: who, Controller: who})
	if err := g.CastSpell(who, id, game.CastSpellParams{Targets: pr7bTargets(targets...)}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

func s2027Wall(g *game.Game, owner uuid.UUID, power int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: power, Toughness: 4})
}

func s2027ShieldCount(g *game.Game) int { return pr7bSourceShields(g) }

// Sewers of Estark on a blocker prevents the combat damage it and the
// attacker it blocks would deal — and only this combat.
func TestSewersOfEstarkBlockerShieldsBothForThisCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	blk := pr7Creature(g, opp.ID, "Blocker", 2, "G")
	pr7bAttack(t, g, opp.ID, att)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
	s2027Cast(t, g, me.ID, "Sewers of Estark", "Instant", s2027Sewers, blk)
	if n := s2027ShieldCount(g); n != 2 {
		t.Fatalf("%d shields, want one for the blocker and one for the attacker it blocks", n)
	}
	pr7bDamageStep(t, g)
	if pr6Marked(g, att)+pr6Marked(g, blk) != 0 {
		t.Fatalf("damage marked: attacker %d, blocker %d — want none", pr6Marked(g, att), pr6Marked(g, blk))
	}
	// The combat is over once the end of combat step ends: no shield left
	// to carry into the postcombat main phase of the same turn.
	advanceTo(t, g, game.StepPostcombatMain)
	if n := s2027ShieldCount(g); n != 0 {
		t.Fatalf("%d shields in the postcombat main phase, want them gone with the combat", n)
	}
	life := opp.Life
	pr7Hit(t, g, att, opp.ID, 2)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d: damage after combat is not prevented", opp.Life, life-2)
	}
}

// Only the creatures the spell names are shielded: another creature's
// combat damage is dealt.
func TestSewersOfEstarkShieldsOnlyTheBlockerAndWhatItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := pr7Creature(g, me.ID, "A1", 3, "R")
	a2 := pr7Creature(g, me.ID, "A2", 4, "R")
	blk := pr7Creature(g, opp.ID, "Blocker", 2, "G")
	pr7bAttack(t, g, opp.ID, a1, a2)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: a1})
	s2027Cast(t, g, me.ID, "Sewers of Estark", "Instant", s2027Sewers, blk)
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life-4 {
		t.Fatalf("life %d, want %d: only the unblocked creature's 4 gets through", opp.Life, life-4)
	}
	if pr6Marked(g, a1)+pr6Marked(g, blk) != 0 {
		t.Fatalf("damage marked on the shielded pair: %d, %d", pr6Marked(g, a1), pr6Marked(g, blk))
	}
}

// Before blockers, on an attacker, it can't be blocked.
func TestSewersOfEstarkOnAnAttackerCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	blk := pr7Creature(g, opp.ID, "Blocker", 2, "G")
	pr7bAttack(t, g, opp.ID, att)
	s2027Cast(t, g, me.ID, "Sewers of Estark", "Instant", s2027Sewers, att)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blk, att); err == nil {
		t.Fatal("the attacker was blocked after Sewers of Estark made it unblockable")
	}
	if n := s2027ShieldCount(g); n != 0 {
		t.Fatalf("%d combat shields on an attacker, want none", n)
	}
}

// Glyph of Destruction: +10/+0 while the combat lasts, damage to the Wall
// prevented all turn, the Wall destroyed at the next end step.
func TestGlyphOfDestructionPumpsForTheCombatThenDestroysTheWall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	wall := s2027Wall(g, opp.ID, 0)
	pr7bAttack(t, g, opp.ID, att)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: att})
	s2027Cast(t, g, opp.ID, "Glyph of Destruction", "Instant", s2027Glyph, wall)
	if p := effectivePower(t, g, wall); p != 10 {
		t.Fatalf("Wall power %d, want 10", p)
	}
	pr7bDamageStep(t, g)
	if pr6Marked(g, wall) != 0 {
		t.Fatalf("%d damage on the Wall, want it prevented", pr6Marked(g, wall))
	}
	if findBattlefieldCardForTest(g, att) != nil {
		t.Fatalf("the attacker survived the Wall's 10 damage, with %d marked", pr6Marked(g, att))
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if p := effectivePower(t, g, wall); p != 0 {
		t.Fatalf("Wall power %d after combat, want the +10/+0 gone with the combat", p)
	}
	if findBattlefieldCardForTest(g, wall) == nil {
		t.Fatal("the Wall was destroyed before the end step")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, wall) != nil {
		t.Fatal("the Wall survived the beginning of the next end step")
	}
}

// Only a blocking Wall you control is a legal target.
func TestGlyphOfDestructionTargetsOnlyYourBlockingWall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	wall := s2027Wall(g, opp.ID, 0)
	notWall := pr7Creature(g, opp.ID, "Bear", 2, "G")
	idleWall := s2027Wall(g, opp.ID, 0)
	pr7bAttack(t, g, opp.ID, att)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: att, notWall: att})
	for name, target := range map[string]uuid.UUID{"a blocking non-Wall": notWall, "a Wall that isn't blocking": idleWall, "an attacker": att} {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Glyph of Destruction", TypeLine: "Instant", OracleID: s2027Glyph, Owner: opp.ID, Controller: opp.ID})
		if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: pr7bTargets(target)}); err == nil {
			t.Errorf("Glyph of Destruction accepted %s as a target", name)
		}
	}
}

// Suppressor Skyguard: an attack at you, with another opponent of the
// attacker not being attacked, shields you for this combat only.
func TestSuppressorSkyguardShieldsYouForThisCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Suppressor Skyguard", "Creature — Human Knight", s2027Skyguard, false)
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	pr7bAttack(t, g, opp.ID, att)
	if n := s2027ShieldCount(g); n != 1 {
		t.Fatalf("%d shields after the attack, want 1", n)
	}
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life {
		t.Fatalf("life %d, want %d: combat damage is prevented", opp.Life, life)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if n := s2027ShieldCount(g); n != 0 {
		t.Fatalf("%d shields after the combat, want none", n)
	}
	pr7Hit(t, g, att, opp.ID, 2)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d: damage after combat is dealt", opp.Life, life-2)
	}
}

// With every other opponent attacked too, the intervening if is false and
// the trigger does nothing.
func TestSuppressorSkyguardNeedsAnotherOpponentWhoIsntBeingAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Suppressor Skyguard", "Creature — Human Knight", s2027Skyguard, false)
	a1 := pr7Creature(g, me.ID, "A1", 3, "R")
	advanceTo(t, g, game.StepDeclareAttackers)
	// Everyone but the attacker is attacked.
	for _, other := range g.Seats[1:] {
		a := a1
		if other.ID != opp.ID {
			a = pr7Creature(g, me.ID, "Extra", 1, "R")
		}
		if err := g.DeclareAttacker(a, other.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if n := s2027ShieldCount(g); n != 0 {
		t.Fatalf("%d shields, want none: nobody else was left unattacked", n)
	}
}

// Old Fat Spider Can't See Me: chapter I's hexproof lasts exactly as long
// as the Saga.
func TestOldFatSpiderChapterIHexproofEndsWithTheSaga(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pr7Creature(g, me.ID, "Mine", 2, "U")
	saga := castCatalogSpell(t, g, "Old Fat Spider Can't See Me", "Enchantment — Saga", s2027OldFat, nil)
	passPriorityAroundTable(t, g)
	s2027AnswerPick(t, g, mine)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, mine, "hexproof") {
		t.Fatal("chapter I did not give the creature hexproof")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(saga); err != nil {
			t.Fatal(err)
		}
	})
	if effectiveAbilitiesContain(t, g, mine, "hexproof") {
		t.Fatal("the creature kept hexproof after the Saga left")
	}
}

// Chapter II prevents all damage the chosen creature would deal while the
// Saga remains, and stops when it goes.
func TestOldFatSpiderChapterIIPreventsDamageWhileTheSagaRemains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pr7Creature(g, me.ID, "Mine", 2, "U")
	foe := pr7Creature(g, opp.ID, "Foe", 5, "R")
	saga := castCatalogSpell(t, g, "Old Fat Spider Can't See Me", "Enchantment — Saga", s2027OldFat, nil)
	passPriorityAroundTable(t, g)
	s2027AnswerPick(t, g, mine)
	passPriorityAroundTable(t, g)
	// Next turn of mine: the second lore counter and chapter II.
	for i := 0; i < 400 && loreCountersOn(g, saga) < 2; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	s2027AnswerPick(t, g, foe)
	passPriorityAroundTable(t, g)
	if n := s2027ShieldCount(g); n != 1 {
		t.Fatalf("%d shields, want the one chapter II made", n)
	}
	life := me.Life
	pr7Hit(t, g, foe, me.ID, 5)
	if me.Life != life {
		t.Fatalf("life %d, want %d: the creature's damage is prevented", me.Life, life)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(saga); err != nil {
			t.Fatal(err)
		}
	})
	pr7Hit(t, g, foe, me.ID, 5)
	if me.Life != life-5 {
		t.Fatalf("life %d, want %d: the damage is dealt once the Saga is gone", me.Life, life-5)
	}
}

// s2027AnswerPick answers the open target pick with `id`, if one is open.
func s2027AnswerPick(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	c := firstPendingTarget(g)
	if c == nil {
		return
	}
	if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}
