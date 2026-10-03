package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr7a_cards_test.go — ADR 0108 Delivery PR 7 (#1904): the
// recipient and "dealt to and dealt by" families of the not-one-use
// shield, each against the damage it must stop and the damage it must
// not.

const (
	pr7aIndestructibleAura = "e10e8d56-bba6-412d-970e-c24969f32b5b"
	pr7aDjerusResolve      = "c48ce123-bb07-4221-b3b3-0a458e101ca1"
	pr7aBraceForImpact     = "f250fb94-e519-4bfa-921e-eddeeafcf76c"
	pr7aRedeem             = "c2032f88-a000-4ed2-a7e5-61d29723d45e"
	pr7aEndure             = "f4cd32f0-d6aa-4497-b8db-a0bf4c3c31de"
	pr7aSafePassage        = "dfa459a1-b065-4488-88d3-4da388261b52"
	pr7aFavoredHoplite     = "145a15f9-89a9-4eb7-9924-e768f17e7e68"
	pr7aMazeOfIth          = "38a12bd7-4394-44a8-91a0-6a4ff7fa4f71"
	pr7aEnergyArc          = "959bcf99-3c9a-4f99-98b2-13514d3dac16"
	pr7aOko                = "f4bf8bd1-71b0-4eb1-976c-54677238b2a6"
	pr7aKiora              = "58cc2097-f4b4-4695-9076-78666050cdb6"
	pr7aAvacyn             = "6da0b6d2-3c5e-49ef-9264-076269c04744"
	pr7aMorningtidesLight  = "a25bcf65-4417-45e9-a4b1-94e0ea78af63"
	pr7aMutationalAdv      = "2daa5b89-e772-4a2b-ad52-bbbf148c7b2f"
	pr7aOriss              = "8d6e0dab-400a-4761-8343-92c7eb7e8735"
	pr7aMoonlightGeist     = "12a8c546-d0d1-4c3d-bfca-c1e59cef349a"
	pr7aRiotControl        = "a189e20f-1072-4608-9e50-da2a9d7a7ee3"
	pr7aUltimateHoly       = "1eae48aa-f01d-4398-a68e-3b5af0542ab1"
	pr7aTakeTheBait        = "5d39cce9-afa1-422a-8e6c-8f50c710a681"
	pr7aKurbis             = "991fab9c-8554-4e64-a933-6cc2ffa4edb1"
)

func pr7aCard(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// pr7aShields counts the live preventFromSource records.
func pr7aShields(g *game.Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModPreventFromSource {
				n++
			}
		}
	}
	return n
}

// pr7aBlockedCombat has `attacker` (the active seat's) attack `defender`
// and be blocked by `blocker`, then stops in the declare blockers step
// with priority to act.
func pr7aBlockedCombat(t *testing.T, g *game.Game, defender, attacker, blocker uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatal(err)
	}
}

// Indestructible Aura: every source's damage to the target this turn is
// prevented; another creature's is not.
func TestPR7aIndestructibleAuraProtectsTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	other := pr7Creature(g, me.ID, "Other", 2, "G")
	a := pr7Creature(g, opp.ID, "A", 3, "R")
	b := pr7Creature(g, opp.ID, "B", 3, "B")
	castCatalogSpell(t, g, "Indestructible Aura", "Instant", pr7aIndestructibleAura, []game.TargetRef{pr7aCard(bear)})
	passPriorityAroundTable(t, g)
	pr6Damage(t, g, a, bear, 3)
	pr6Damage(t, g, b, bear, 3)
	pr6Damage(t, g, a, other, 1)
	if pr6Marked(g, bear) != 0 || pr6Marked(g, other) != 1 {
		t.Fatalf("bear %d (want 0), other %d (want 1)", pr6Marked(g, bear), pr6Marked(g, other))
	}
}

// Djeru's Resolve untaps the creature and shields it.
func TestPR7aDjerusResolveUntapsAndShields(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	findBattlefieldCardForTest(g, bear).Tapped = true
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(bear)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, bear).Tapped {
		t.Error("the creature is still tapped")
	}
	pr6Damage(t, g, src, bear, 3)
	if got := pr6Marked(g, bear); got != 0 {
		t.Errorf("bear has %d damage, want 0", got)
	}
}

// Brace for Impact: a +1/+1 counter for each 1 damage prevented, once
// per instance (CR 615.5).
func TestPR7aBraceForImpactCountsThePreventedDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gold := pr7Creature(g, me.ID, "Gold", 2, "W", "U")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Brace for Impact", "Instant", pr7aBraceForImpact, []game.TargetRef{pr7aCard(gold)})
	passPriorityAroundTable(t, g)
	pr6Damage(t, g, src, gold, 3)
	pr6Damage(t, g, src, gold, 2)
	passPriorityAroundTable(t, g)
	if pr6Marked(g, gold) != 0 {
		t.Fatalf("the creature has %d damage, want 0", pr6Marked(g, gold))
	}
	if got := pr7Counters(g, gold); got != 5 {
		t.Errorf("%d +1/+1 counters, want 5", got)
	}
}

// Brace for Impact targets only a multicolored creature.
func TestPR7aBraceForImpactNeedsAMulticoloredTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mono := pr7Creature(g, me.ID, "Mono", 2, "W")
	if err := castCatalogSpellErr(t, g, "Brace for Impact", "Instant", pr7aBraceForImpact, []game.TargetRef{pr7aCard(mono)}); err == nil {
		t.Fatal("a monocolored creature was a legal target")
	}
}

// Redeem is one record protecting both targets.
func TestPR7aRedeemProtectsBothTargetsInOneRecord(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, me.ID, "A", 2, "W")
	b := pr7Creature(g, me.ID, "B", 2, "W")
	c := pr7Creature(g, me.ID, "C", 2, "W")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Redeem", "Instant", pr7aRedeem, []game.TargetRef{pr7aCard(a), pr7aCard(b)})
	passPriorityAroundTable(t, g)
	if n := pr7aShields(g); n != 1 {
		t.Fatalf("%d shields, want one record", n)
	}
	pr6Damage(t, g, src, a, 2)
	pr6Damage(t, g, src, b, 2)
	pr6Damage(t, g, src, c, 1)
	if pr6Marked(g, a) != 0 || pr6Marked(g, b) != 0 || pr6Marked(g, c) != 1 {
		t.Fatalf("a %d, b %d, c %d (want 0, 0, 1)", pr6Marked(g, a), pr6Marked(g, b), pr6Marked(g, c))
	}
}

// Endure protects you and every permanent you control, and nothing an
// opponent controls.
func TestPR7aEndureProtectsYouAndYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Endure", "Instant", pr7aEndure, nil)
	passPriorityAroundTable(t, g)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	pr6Damage(t, g, src, bear, 2)
	pr6Damage(t, g, src, theirs, 1)
	if me.Life != life || pr6Marked(g, bear) != 0 || pr6Marked(g, theirs) != 1 {
		t.Fatalf("life %d (want %d), bear %d (want 0), theirs %d (want 1)", me.Life, life, pr6Marked(g, bear), pr6Marked(g, theirs))
	}
}

// Safe Passage protects you and your creatures; an opponent's damage to
// them is prevented, and to another player's creature is not.
func TestPR7aSafePassageProtectsYouAndYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Safe Passage", "Instant", pr7aSafePassage, nil)
	passPriorityAroundTable(t, g)
	life, theirLife := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 3)
	pr7Hit(t, g, bear, opp.ID, 2)
	pr6Damage(t, g, src, bear, 2)
	pr6Damage(t, g, src, theirs, 1)
	if me.Life != life || pr6Marked(g, bear) != 0 || pr6Marked(g, theirs) != 1 || opp.Life != theirLife-2 {
		t.Fatalf("life %d (want %d), bear %d, theirs %d, opponent %d (want %d)",
			me.Life, life, pr6Marked(g, bear), pr6Marked(g, theirs), opp.Life, theirLife-2)
	}
}

// Favored Hoplite: casting a spell that targets it triggers heroic; a
// spell that targets another creature does not.
func TestPR7aFavoredHopliteHeroic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hoplite := pushCatalogPermanent(g, me.ID, "Favored Hoplite", "Creature — Human Soldier", pr7aFavoredHoplite, false)
	other := pr7Creature(g, me.ID, "Other", 2, "W")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Indestructible Aura", "Instant", pr7aIndestructibleAura, []game.TargetRef{pr7aCard(other)})
	passPriorityAroundTable(t, g)
	if got := pr7Counters(g, hoplite); got != 0 {
		t.Fatalf("heroic triggered on a spell targeting another creature (%d counters)", got)
	}
	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(hoplite)})
	passPriorityAroundTable(t, g)
	if got := pr7Counters(g, hoplite); got != 1 {
		t.Fatalf("%d counters, want 1", got)
	}
	pr6Damage(t, g, src, hoplite, 5)
	if got := pr6Marked(g, hoplite); got != 0 {
		t.Errorf("hoplite has %d damage, want 0", got)
	}
}

// Maze of Ith: the attacker is untapped, and its combat damage and the
// blocker's combat damage to it are both prevented by ONE record.
func TestPR7aMazeOfIthPreventsToAndByAsOneEffect(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maze := pushCatalogPermanent(g, me.ID, "Maze of Ith", "Land", pr7aMazeOfIth, false)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 3, 5)
	pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	pr7Activate(t, g, me.ID, maze, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{pr7aCard(attacker)}})
	if findBattlefieldCardForTest(g, attacker).Tapped {
		t.Fatal("the attacker was not untapped")
	}
	if n := pr7aShields(g); n != 1 {
		t.Fatalf("%d shields, want one record for both directions", n)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, attacker) != 0 || damageMarkedOn(g, blocker) != 0 {
		t.Fatalf("attacker %d, blocker %d damage: want both prevented", damageMarkedOn(g, attacker), damageMarkedOn(g, blocker))
	}
}

// The control for the combat tests above: with no shield, the same
// combat marks damage on both creatures.
func TestPR7aCombatHarnessDealsDamageWithoutAShield(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	attacker := pushVanillaCreature(g, g.Seats[0].ID, "Attacker", 4, 4)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 3, 5)
	pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, attacker) != 3 || damageMarkedOn(g, blocker) != 4 {
		t.Fatalf("attacker %d, blocker %d damage: want 3 and 4", damageMarkedOn(g, attacker), damageMarkedOn(g, blocker))
	}
}

// Maze of Ith is combat damage only: the creature's other damage is dealt.
func TestPR7aMazeOfIthLeavesNonCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maze := pushCatalogPermanent(g, me.ID, "Maze of Ith", "Land", pr7aMazeOfIth, false)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	pr7Activate(t, g, me.ID, maze, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{pr7aCard(attacker)}})
	life := opp.Life
	pr7Hit(t, g, attacker, opp.ID, 2)
	if opp.Life != life-2 {
		t.Fatalf("opponent life %d, want %d: non-combat damage is dealt", opp.Life, life-2)
	}
}

// Energy Arc: both creatures of a combat untap and are shielded both
// ways by one record.
func TestPR7aEnergyArcIsOneRecordOverThoseCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 3, 5)
	pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	castCatalogSpell(t, g, "Energy Arc", "Instant", pr7aEnergyArc, []game.TargetRef{pr7aCard(attacker), pr7aCard(blocker)})
	passPriorityAroundTable(t, g)
	if n := pr7aShields(g); n != 1 {
		t.Fatalf("%d shields, want one", n)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, attacker) != 0 || damageMarkedOn(g, blocker) != 0 {
		t.Fatalf("attacker %d, blocker %d damage: want both prevented", damageMarkedOn(g, attacker), damageMarkedOn(g, blocker))
	}
}

// Moonlight Geist shields itself both ways; a creature that left and
// came back would be a new object, so only this one is pinned.
func TestPR7aMoonlightGeistShieldsItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := pushCatalogPermanent(g, me.ID, "Moonlight Geist", "Creature — Spirit", pr7aMoonlightGeist, false)
	blocker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker", TypeLine: "Creature — Spider", Power: 3, Toughness: 5, Keywords: []string{"reach"}})
	pr7aBlockedCombat(t, g, opp.ID, geist, blocker)
	pr7Activate(t, g, me.ID, geist, 0, game.ActivateAbilityParams{})
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, geist) != 0 || damageMarkedOn(g, blocker) != 0 {
		t.Fatalf("geist %d, blocker %d damage: want both prevented", damageMarkedOn(g, geist), damageMarkedOn(g, blocker))
	}
}

// Kiora's +1 lasts until your next turn and covers all damage, both
// ways.
func TestPR7aKioraPlusOneLastsUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	kiora := pushCatalogPermanent(g, me.ID, "Kiora, the Crashing Wave", "Legendary Planeswalker — Kiora", pr7aKiora, false)
	findBattlefieldCardForTest(g, kiora).Counters = map[string]int{game.CounterLoyalty: 2}
	theirs := pr7Creature(g, opp.ID, "Theirs", 4, "R")
	mine := pr7Creature(g, me.ID, "Mine", 2, "G")
	pr7Activate(t, g, me.ID, kiora, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{pr7aCard(theirs)}})
	b39NextTurnOf(t, g, 1)
	life := me.Life
	pr7Hit(t, g, theirs, me.ID, 4)
	pr6Damage(t, g, mine, theirs, 2)
	if me.Life != life || pr6Marked(g, theirs) != 0 {
		t.Fatalf("life %d (want %d), their creature %d damage (want 0)", me.Life, life, pr6Marked(g, theirs))
	}
	b39NextTurnOf(t, g, 0)
	pr7Hit(t, g, theirs, me.ID, 4)
	if me.Life != life-4 {
		t.Fatalf("life %d, want %d: the shield ended as your turn began", me.Life, life-4)
	}
}

// Oko's 0: Oko becomes a copy of the creature and is shielded from all
// damage this turn.
func TestPR7aOkoBecomesACopyAndIsShielded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	oko := pushCatalogPermanent(g, me.ID, "Oko, the Trickster", "Legendary Planeswalker — Oko", pr7aOko, false)
	findBattlefieldCardForTest(g, oko).Counters = map[string]int{game.CounterLoyalty: 4}
	beast := apaPush(g, me.ID, me.ID, game.Card{Name: "Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4, Colors: []string{"G"}})
	src := pr7Creature(g, opp.ID, "Src", 5, "R")
	pr7Activate(t, g, me.ID, oko, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{pr7aCard(beast)}})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardForTest(g, oko)
	if c == nil || !c.IsCreature() {
		t.Fatal("Oko is not a creature copy")
	}
	pr6Damage(t, g, src, oko, 5)
	if got := pr6Marked(g, oko); got != 0 {
		t.Errorf("Oko has %d damage, want 0", got)
	}
}

// Avacyn's first ability: only sources of the chosen colour are stopped.
func TestPR7aAvacynShieldsAgainstTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avacyn := pushCatalogPermanent(g, me.ID, "Avacyn, Guardian Angel", "Legendary Creature — Angel", pr7aAvacyn, false)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	red := pr7Creature(g, opp.ID, "Red", 3, "R")
	black := pr7Creature(g, opp.ID, "Black", 3, "B")
	pr7Activate(t, g, me.ID, avacyn, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{pr7aCard(bear)}})
	answerColor(t, g, me.ID, "R")
	passPriorityAroundTable(t, g)
	pr6Damage(t, g, red, bear, 3)
	if got := pr6Marked(g, bear); got != 0 {
		t.Fatalf("bear has %d damage from the red source, want 0", got)
	}
	pr6Damage(t, g, black, bear, 1)
	if got := pr6Marked(g, bear); got != 1 {
		t.Errorf("bear has %d damage, want the black source's 1", got)
	}
}

// Morningtide's Light: the creatures are exiled, you are shielded until
// your next turn, and the spell exiles itself.
func TestPR7aMorningtidesLightExilesShieldsAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pr7Creature(g, opp.ID, "Theirs", 3, "R")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	spell := castCatalogSpell(t, g, "Morningtide's Light", "Sorcery", pr7aMorningtidesLight, []game.TargetRef{pr7aCard(theirs)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, theirs) != nil {
		t.Fatal("the target was not exiled")
	}
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Morningtide's Light is in %v, want exile", z)
	}
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	// The card returns as a new object (CR 400.7), under a new ID.
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Theirs" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil || !back.Tapped || back.Controller != opp.ID {
		t.Fatalf("the creature came back as %+v, want tapped under its owner's control", back)
	}
}

// Mutational Advantage: the permanents with counters are protected;
// one without is not.
func TestPR7aMutationalAdvantageProtectsThosePermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	countered := pr7Creature(g, me.ID, "Countered", 2, "G")
	findBattlefieldCardForTest(g, countered).Counters = map[string]int{game.CounterPlusOne: 1}
	plain := pr7Creature(g, me.ID, "Plain", 2, "G")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Mutational Advantage", "Instant", pr7aMutationalAdv, nil)
	passPriorityAroundTable(t, g)
	pr6Damage(t, g, src, countered, 3)
	pr6Damage(t, g, src, plain, 1)
	if pr6Marked(g, countered) != 0 || pr6Marked(g, plain) != 1 {
		t.Fatalf("countered %d (want 0), plain %d (want 1)", pr6Marked(g, countered), pr6Marked(g, plain))
	}
}

// Riot Control: a life per creature your opponents control, and no
// damage to you this turn.
func TestPR7aRiotControlGainsAndShields(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	pr7Creature(g, g.Seats[2].ID, "Third", 1, "G")
	pr7Creature(g, me.ID, "Mine", 1, "G")
	life := me.Life
	castCatalogSpell(t, g, "Riot Control", "Instant", pr7aRiotControl, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life %d, want %d", me.Life, life+2)
	}
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life+2 {
		t.Fatalf("life %d after the hit, want %d", me.Life, life+2)
	}
}

// Ultimate Magic: Holy cast from the hand grants indestructible but does
// not shield you.
func TestPR7aUltimateMagicHolyFromHandDoesNotShield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Ultimate Magic: Holy", "Instant", pr7aUltimateHoly, nil)
	passPriorityAroundTable(t, g)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d: cast from the hand, no shield", me.Life, life-3)
	}
}

// Ultimate Magic: Holy foretold and cast from exile shields you.
func TestPR7aUltimateMagicHolyFromExileShields(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(me, "Ultimate Magic: Holy", "Instant", pr7aUltimateHoly)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.Seq++ })
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "W"})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell}); err != nil {
		t.Fatalf("cast the foretold card: %v", err)
	}
	passPriorityAroundTable(t, g)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d, want %d: cast from exile, you are shielded", me.Life, life)
	}
}

// Take the Bait can't be cast on your own turn.
func TestPR7aTakeTheBaitOnlyOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	if err := castCatalogSpellErr(t, g, "Take the Bait", "Instant", pr7aTakeTheBait, nil); err == nil {
		t.Fatal("Take the Bait was cast during its controller's own turn")
	}
}

// Oriss's grandeur: the target player can't cast spells this turn.
func TestPR7aOrissGrandeurStopsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	oriss := pushCatalogPermanent(g, me.ID, "Oriss, Samite Guardian", "Legendary Creature — Human Cleric", pr7aOriss, false)
	extra := handCardForTest(me, "Oriss, Samite Guardian", "Legendary Creature — Human Cleric", pr7aOriss)
	if err := g.ActivateCatalogAbility(me.ID, oriss, 1, game.ActivateAbilityParams{
		Targets:    []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		DiscardIDs: []uuid.UUID{extra},
	}); err != nil {
		t.Fatalf("grandeur: %v", err)
	}
	passPriorityAroundTable(t, g)
	banned := false
	for _, s := range opp.Statics {
		if s.CastBan.Kind == game.CastBanOutright {
			banned = true
		}
	}
	if !banned {
		t.Error("the target player has no cast ban")
	}
	if z := g.FindCardZoneForEffect(extra); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("the discarded Oriss is in %v, want the graveyard", z)
	}
}

// Kurbis's ability shields another creature with a +1/+1 counter.
func TestPR7aKurbisShieldsACounteredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kurbis := pushCatalogPermanent(g, me.ID, "Kurbis, Harvest Celebrant", "Legendary Creature — Treefolk", pr7aKurbis, false)
	findBattlefieldCardForTest(g, kurbis).Counters = map[string]int{game.CounterPlusOne: 2}
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	findBattlefieldCardForTest(g, bear).Counters = map[string]int{game.CounterPlusOne: 1}
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	pr7Activate(t, g, me.ID, kurbis, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{pr7aCard(bear)}})
	pr6Damage(t, g, src, bear, 3)
	if got := pr6Marked(g, bear); got != 0 {
		t.Errorf("bear has %d damage, want 0", got)
	}
	if got := pr7Counters(g, kurbis); got != 1 {
		t.Errorf("Kurbis has %d counters, want 1 after paying", got)
	}
}
