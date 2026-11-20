package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Indestructible Aura: every source's damage to the target this turn is
// prevented; another creature's is not.
func TestPR7aIndestructibleAuraProtectsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	other := effectstest.Pr7Creature(g, me.ID, "Other", 2, "G")
	a := effectstest.Pr7Creature(g, opp.ID, "A", 3, "R")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 3, "B")
	effectstest.CastCatalogSpell(t, g, "Indestructible Aura", "Instant", effectstest.Pr7aIndestructibleAura, []game.TargetRef{effectstest.Pr7aCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr6Damage(t, g, a, bear, 3)
	effectstest.Pr6Damage(t, g, b, bear, 3)
	effectstest.Pr6Damage(t, g, a, other, 1)
	if effectstest.Pr6Marked(g, bear) != 0 || effectstest.Pr6Marked(g, other) != 1 {
		t.Fatalf("bear %d (want 0), other %d (want 1)", effectstest.Pr6Marked(g, bear), effectstest.Pr6Marked(g, other))
	}
}

// Djeru's Resolve untaps the creature and shields it.
func TestPR7aDjerusResolveUntapsAndShields(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Djeru's Resolve", "Instant", effectstest.Pr7aDjerusResolve, []game.TargetRef{effectstest.Pr7aCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, bear).Tapped {
		t.Error("the creature is still tapped")
	}
	effectstest.Pr6Damage(t, g, src, bear, 3)
	if got := effectstest.Pr6Marked(g, bear); got != 0 {
		t.Errorf("bear has %d damage, want 0", got)
	}
}

// Brace for Impact: a +1/+1 counter for each 1 damage prevented, once
// per instance (CR 615.5).
func TestPR7aBraceForImpactCountsThePreventedDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gold := effectstest.Pr7Creature(g, me.ID, "Gold", 2, "W", "U")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Brace for Impact", "Instant", effectstest.Pr7aBraceForImpact, []game.TargetRef{effectstest.Pr7aCard(gold)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr6Damage(t, g, src, gold, 3)
	effectstest.Pr6Damage(t, g, src, gold, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.Pr6Marked(g, gold) != 0 {
		t.Fatalf("the creature has %d damage, want 0", effectstest.Pr6Marked(g, gold))
	}
	if got := effectstest.Pr7Counters(g, gold); got != 5 {
		t.Errorf("%d +1/+1 counters, want 5", got)
	}
}

// Brace for Impact targets only a multicolored creature.
func TestPR7aBraceForImpactNeedsAMulticoloredTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mono := effectstest.Pr7Creature(g, me.ID, "Mono", 2, "W")
	if err := effectstest.CastCatalogSpellErr(t, g, "Brace for Impact", "Instant", effectstest.Pr7aBraceForImpact, []game.TargetRef{effectstest.Pr7aCard(mono)}); err == nil {
		t.Fatal("a monocolored creature was a legal target")
	}
}

// Redeem is one record protecting both targets.
func TestPR7aRedeemProtectsBothTargetsInOneRecord(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Pr7Creature(g, me.ID, "A", 2, "W")
	b := effectstest.Pr7Creature(g, me.ID, "B", 2, "W")
	c := effectstest.Pr7Creature(g, me.ID, "C", 2, "W")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Redeem", "Instant", effectstest.Pr7aRedeem, []game.TargetRef{effectstest.Pr7aCard(a), effectstest.Pr7aCard(b)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.Pr7aShields(g); n != 1 {
		t.Fatalf("%d shields, want one record", n)
	}
	effectstest.Pr6Damage(t, g, src, a, 2)
	effectstest.Pr6Damage(t, g, src, b, 2)
	effectstest.Pr6Damage(t, g, src, c, 1)
	if effectstest.Pr6Marked(g, a) != 0 || effectstest.Pr6Marked(g, b) != 0 || effectstest.Pr6Marked(g, c) != 1 {
		t.Fatalf("a %d, b %d, c %d (want 0, 0, 1)", effectstest.Pr6Marked(g, a), effectstest.Pr6Marked(g, b), effectstest.Pr6Marked(g, c))
	}
}

// Endure protects you and every permanent you control, and nothing an
// opponent controls.
func TestPR7aEndureProtectsYouAndYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "G")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Endure", "Instant", effectstest.Pr7aEndure, nil)
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	effectstest.Pr6Damage(t, g, src, bear, 2)
	effectstest.Pr6Damage(t, g, src, theirs, 1)
	if me.Life != life || effectstest.Pr6Marked(g, bear) != 0 || effectstest.Pr6Marked(g, theirs) != 1 {
		t.Fatalf("life %d (want %d), bear %d (want 0), theirs %d (want 1)", me.Life, life, effectstest.Pr6Marked(g, bear), effectstest.Pr6Marked(g, theirs))
	}
}

// Safe Passage protects you and your creatures; an opponent's damage to
// them is prevented, and to another player's creature is not.
func TestPR7aSafePassageProtectsYouAndYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "G")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Safe Passage", "Instant", effectstest.Pr7aSafePassage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	life, theirLife := me.Life, opp.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	effectstest.Pr7Hit(t, g, bear, opp.ID, 2)
	effectstest.Pr6Damage(t, g, src, bear, 2)
	effectstest.Pr6Damage(t, g, src, theirs, 1)
	if me.Life != life || effectstest.Pr6Marked(g, bear) != 0 || effectstest.Pr6Marked(g, theirs) != 1 || opp.Life != theirLife-2 {
		t.Fatalf("life %d (want %d), bear %d, theirs %d, opponent %d (want %d)",
			me.Life, life, effectstest.Pr6Marked(g, bear), effectstest.Pr6Marked(g, theirs), opp.Life, theirLife-2)
	}
}

// Favored Hoplite: casting a spell that targets it triggers heroic; a
// spell that targets another creature does not.
func TestPR7aFavoredHopliteHeroic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hoplite := effectstest.PushCatalogPermanent(g, me.ID, "Favored Hoplite", "Creature — Human Soldier", effectstest.Pr7aFavoredHoplite, false)
	other := effectstest.Pr7Creature(g, me.ID, "Other", 2, "W")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Indestructible Aura", "Instant", effectstest.Pr7aIndestructibleAura, []game.TargetRef{effectstest.Pr7aCard(other)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr7Counters(g, hoplite); got != 0 {
		t.Fatalf("heroic triggered on a spell targeting another creature (%d counters)", got)
	}
	effectstest.CastCatalogSpell(t, g, "Djeru's Resolve", "Instant", effectstest.Pr7aDjerusResolve, []game.TargetRef{effectstest.Pr7aCard(hoplite)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr7Counters(g, hoplite); got != 1 {
		t.Fatalf("%d counters, want 1", got)
	}
	effectstest.Pr6Damage(t, g, src, hoplite, 5)
	if got := effectstest.Pr6Marked(g, hoplite); got != 0 {
		t.Errorf("hoplite has %d damage, want 0", got)
	}
}

// Maze of Ith: the attacker is untapped, and its combat damage and the
// blocker's combat damage to it are both prevented by ONE record.
func TestPR7aMazeOfIthPreventsToAndByAsOneEffect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maze := effectstest.PushCatalogPermanent(g, me.ID, "Maze of Ith", "Land", effectstest.Pr7aMazeOfIth, false)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 3, 5)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	effectstest.Pr7Activate(t, g, me.ID, maze, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.Pr7aCard(attacker)}})
	if effectstest.FindBattlefieldCardForTest(g, attacker).Tapped {
		t.Fatal("the attacker was not untapped")
	}
	if n := effectstest.Pr7aShields(g); n != 1 {
		t.Fatalf("%d shields, want one record for both directions", n)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, attacker) != 0 || effectstest.DamageMarkedOn(g, blocker) != 0 {
		t.Fatalf("attacker %d, blocker %d damage: want both prevented", effectstest.DamageMarkedOn(g, attacker), effectstest.DamageMarkedOn(g, blocker))
	}
}

// The control for the combat tests above: with no shield, the same
// combat marks damage on both creatures.
func TestPR7aCombatHarnessDealsDamageWithoutAShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Attacker", 4, 4)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 3, 5)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, attacker) != 3 || effectstest.DamageMarkedOn(g, blocker) != 4 {
		t.Fatalf("attacker %d, blocker %d damage: want 3 and 4", effectstest.DamageMarkedOn(g, attacker), effectstest.DamageMarkedOn(g, blocker))
	}
}

// Maze of Ith is combat damage only: the creature's other damage is dealt.
func TestPR7aMazeOfIthLeavesNonCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maze := effectstest.PushCatalogPermanent(g, me.ID, "Maze of Ith", "Land", effectstest.Pr7aMazeOfIth, false)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.Pr7Activate(t, g, me.ID, maze, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.Pr7aCard(attacker)}})
	life := opp.Life
	effectstest.Pr7Hit(t, g, attacker, opp.ID, 2)
	if opp.Life != life-2 {
		t.Fatalf("opponent life %d, want %d: non-combat damage is dealt", opp.Life, life-2)
	}
}

// Energy Arc: both creatures of a combat untap and are shielded both
// ways by one record.
func TestPR7aEnergyArcIsOneRecordOverThoseCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 3, 5)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	// Cast in the declare blockers step it is meant for: the generic
	// helper would walk on to a main phase, and with the combat damage
	// step skipped when nothing attacks (CR 508.8) that walk would
	// never come back to this combat's damage.
	if err := effectstest.S58p6CastNow(g, me, "Energy Arc", "Instant", effectstest.Pr7aEnergyArc, []game.TargetRef{effectstest.Pr7aCard(attacker), effectstest.Pr7aCard(blocker)}); err != nil {
		t.Fatalf("Energy Arc in the declare blockers step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.Pr7aShields(g); n != 1 {
		t.Fatalf("%d shields, want one", n)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, attacker) != 0 || effectstest.DamageMarkedOn(g, blocker) != 0 {
		t.Fatalf("attacker %d, blocker %d damage: want both prevented", effectstest.DamageMarkedOn(g, attacker), effectstest.DamageMarkedOn(g, blocker))
	}
}

// Moonlight Geist shields itself both ways; a creature that left and
// came back would be a new object, so only this one is pinned.
func TestPR7aMoonlightGeistShieldsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := effectstest.PushCatalogPermanent(g, me.ID, "Moonlight Geist", "Creature — Spirit", effectstest.Pr7aMoonlightGeist, false)
	blocker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker", TypeLine: "Creature — Spider", Power: 3, Toughness: 5, Keywords: []string{"reach"}})
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, geist, blocker)
	effectstest.Pr7Activate(t, g, me.ID, geist, 0, game.ActivateAbilityParams{})
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, geist) != 0 || effectstest.DamageMarkedOn(g, blocker) != 0 {
		t.Fatalf("geist %d, blocker %d damage: want both prevented", effectstest.DamageMarkedOn(g, geist), effectstest.DamageMarkedOn(g, blocker))
	}
}

// Kiora's +1 lasts until your next turn and covers all damage, both
// ways.
func TestPR7aKioraPlusOneLastsUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMainOf(t, g, 0)
	kiora := effectstest.PushCatalogPermanent(g, me.ID, "Kiora, the Crashing Wave", "Legendary Planeswalker — Kiora", effectstest.Pr7aKiora, false)
	effectstest.FindBattlefieldCardForTest(g, kiora).Counters = map[string]int{game.CounterLoyalty: 2}
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 4, "R")
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 2, "G")
	effectstest.Pr7Activate(t, g, me.ID, kiora, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.Pr7aCard(theirs)}})
	effectstest.B39NextTurnOf(t, g, 1)
	life := me.Life
	effectstest.Pr7Hit(t, g, theirs, me.ID, 4)
	effectstest.Pr6Damage(t, g, mine, theirs, 2)
	if me.Life != life || effectstest.Pr6Marked(g, theirs) != 0 {
		t.Fatalf("life %d (want %d), their creature %d damage (want 0)", me.Life, life, effectstest.Pr6Marked(g, theirs))
	}
	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.Pr7Hit(t, g, theirs, me.ID, 4)
	if me.Life != life-4 {
		t.Fatalf("life %d, want %d: the shield ended as your turn began", me.Life, life-4)
	}
}

// Oko's 0: Oko becomes a copy of the creature and is shielded from all
// damage this turn.
func TestPR7aOkoBecomesACopyAndIsShielded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMainOf(t, g, 0)
	oko := effectstest.PushCatalogPermanent(g, me.ID, "Oko, the Trickster", "Legendary Planeswalker — Oko", effectstest.Pr7aOko, false)
	effectstest.FindBattlefieldCardForTest(g, oko).Counters = map[string]int{game.CounterLoyalty: 4}
	beast := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4, Colors: []string{"G"}})
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 5, "R")
	effectstest.Pr7Activate(t, g, me.ID, oko, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.Pr7aCard(beast)}})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardForTest(g, oko)
	if c == nil || !c.IsCreature() {
		t.Fatal("Oko is not a creature copy")
	}
	effectstest.Pr6Damage(t, g, src, oko, 5)
	if got := effectstest.Pr6Marked(g, oko); got != 0 {
		t.Errorf("Oko has %d damage, want 0", got)
	}
}

// Avacyn's first ability: only sources of the chosen colour are stopped.
func TestPR7aAvacynShieldsAgainstTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avacyn := effectstest.PushCatalogPermanent(g, me.ID, "Avacyn, Guardian Angel", "Legendary Creature — Angel", effectstest.Pr7aAvacyn, false)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	red := effectstest.Pr7Creature(g, opp.ID, "Red", 3, "R")
	black := effectstest.Pr7Creature(g, opp.ID, "Black", 3, "B")
	effectstest.Pr7Activate(t, g, me.ID, avacyn, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.Pr7aCard(bear)}})
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr6Damage(t, g, red, bear, 3)
	if got := effectstest.Pr6Marked(g, bear); got != 0 {
		t.Fatalf("bear has %d damage from the red source, want 0", got)
	}
	effectstest.Pr6Damage(t, g, black, bear, 1)
	if got := effectstest.Pr6Marked(g, bear); got != 1 {
		t.Errorf("bear has %d damage, want the black source's 1", got)
	}
}

// Morningtide's Light: the creatures are exiled, you are shielded until
// your next turn, and the spell exiles itself.
func TestPR7aMorningtidesLightExilesShieldsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 3, "R")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	spell := effectstest.CastCatalogSpell(t, g, "Morningtide's Light", "Sorcery", effectstest.Pr7aMorningtidesLight, []game.TargetRef{effectstest.Pr7aCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, theirs) != nil {
		t.Fatal("the target was not exiled")
	}
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Morningtide's Light is in %v, want exile", z)
	}
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	countered := effectstest.Pr7Creature(g, me.ID, "Countered", 2, "G")
	effectstest.FindBattlefieldCardForTest(g, countered).Counters = map[string]int{game.CounterPlusOne: 1}
	plain := effectstest.Pr7Creature(g, me.ID, "Plain", 2, "G")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Mutational Advantage", "Instant", effectstest.Pr7aMutationalAdv, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr6Damage(t, g, src, countered, 3)
	effectstest.Pr6Damage(t, g, src, plain, 1)
	if effectstest.Pr6Marked(g, countered) != 0 || effectstest.Pr6Marked(g, plain) != 1 {
		t.Fatalf("countered %d (want 0), plain %d (want 1)", effectstest.Pr6Marked(g, countered), effectstest.Pr6Marked(g, plain))
	}
}

// Riot Control: a life per creature your opponents control, and no
// damage to you this turn.
func TestPR7aRiotControlGainsAndShields(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.Pr7Creature(g, g.Seats[2].ID, "Third", 1, "G")
	effectstest.Pr7Creature(g, me.ID, "Mine", 1, "G")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Riot Control", "Instant", effectstest.Pr7aRiotControl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life %d, want %d", me.Life, life+2)
	}
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life+2 {
		t.Fatalf("life %d after the hit, want %d", me.Life, life+2)
	}
}

// Ultimate Magic: Holy cast from the hand grants indestructible but does
// not shield you.
func TestPR7aUltimateMagicHolyFromHandDoesNotShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Ultimate Magic: Holy", "Instant", effectstest.Pr7aUltimateHoly, nil)
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d: cast from the hand, no shield", me.Life, life-3)
	}
}

// Ultimate Magic: Holy foretold and cast from exile shields you.
func TestPR7aUltimateMagicHolyFromExileShields(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	card := effectstest.HandCardForTest(me, "Ultimate Magic: Holy", "Instant", effectstest.Pr7aUltimateHoly)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.Seq++ })
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "W"})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell}); err != nil {
		t.Fatalf("cast the foretold card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d, want %d: cast from exile, you are shielded", me.Life, life)
	}
}

// Take the Bait can't be cast on your own turn.
func TestPR7aTakeTheBaitOnlyOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.CastCatalogSpellErr(t, g, "Take the Bait", "Instant", effectstest.Pr7aTakeTheBait, nil); err == nil {
		t.Fatal("Take the Bait was cast during its controller's own turn")
	}
}

// Oriss's grandeur: the target player can't cast spells this turn.
func TestPR7aOrissGrandeurStopsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMainOf(t, g, 0)
	oriss := effectstest.PushCatalogPermanent(g, me.ID, "Oriss, Samite Guardian", "Legendary Creature — Human Cleric", effectstest.Pr7aOriss, false)
	extra := effectstest.HandCardForTest(me, "Oriss, Samite Guardian", "Legendary Creature — Human Cleric", effectstest.Pr7aOriss)
	if err := g.ActivateCatalogAbility(me.ID, oriss, 1, game.ActivateAbilityParams{
		Targets:    []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		DiscardIDs: []uuid.UUID{extra},
	}); err != nil {
		t.Fatalf("grandeur: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kurbis := effectstest.PushCatalogPermanent(g, me.ID, "Kurbis, Harvest Celebrant", "Legendary Creature — Treefolk", effectstest.Pr7aKurbis, false)
	effectstest.FindBattlefieldCardForTest(g, kurbis).Counters = map[string]int{game.CounterPlusOne: 2}
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.FindBattlefieldCardForTest(g, bear).Counters = map[string]int{game.CounterPlusOne: 1}
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.Pr7Activate(t, g, me.ID, kurbis, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.Pr7aCard(bear)}})
	effectstest.Pr6Damage(t, g, src, bear, 3)
	if got := effectstest.Pr6Marked(g, bear); got != 0 {
		t.Errorf("bear has %d damage, want 0", got)
	}
	if got := effectstest.Pr7Counters(g, kurbis); got != 1 {
		t.Errorf("Kurbis has %d counters, want 1 after paying", got)
	}
}
