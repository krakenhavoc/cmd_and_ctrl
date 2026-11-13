package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gn3_cards_test.go — the Game Night: Free-for-All legends (#5485):
// Imaryll, Elfhame Elite; Maeve, Insidious Singer; Vogar, Necropolis
// Tyrant; and Zamriel, Seraph of Steel. Each test plays the real catalog
// entry and asserts what its text says.

const (
	gn3ImaryllOracle = "ef6eb92f-93e6-40e4-a7a6-8e826bb6a5e0"
	gn3MaeveOracle   = "2494daad-c81d-4a80-ba5d-e7011af8de46"
	gn3VogarOracle   = "1d704c2f-cd24-4a04-a51a-da7984dcd358"
	gn3ZamrielOracle = "acb0fa6a-89c8-4957-881e-c61659e86f97"
)

func TestGN3CardsAreRegisteredFull(t *testing.T) {
	for oracle, want := range map[string]struct {
		name     string
		keywords []string
	}{
		gn3ImaryllOracle: {"Imaryll, Elfhame Elite", nil},
		gn3MaeveOracle:   {"Maeve, Insidious Singer", nil},
		gn3VogarOracle:   {"Vogar, Necropolis Tyrant", []string{"menace"}},
		gn3ZamrielOracle: {"Zamriel, Seraph of Steel", []string{"flying"}},
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != want.name {
			t.Errorf("%s: registered=%v name=%q", want.name, ok, spec.Name)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want Full with no caveats, got %v / %v", want.name, spec.Completeness, spec.Caveats)
		}
		if len(spec.PrintedKeywords) != len(want.keywords) {
			t.Errorf("%s: printed keywords %v, want %v", want.name, spec.PrintedKeywords, want.keywords)
			continue
		}
		for i, kw := range want.keywords {
			if spec.PrintedKeywords[i] != kw {
				t.Errorf("%s: printed keywords %v, want %v", want.name, spec.PrintedKeywords, want.keywords)
			}
		}
	}
}

// --- Imaryll, Elfhame Elite -----------------------------------------------

// gn3Imaryll seeds Imaryll for seat 0 and `elves` other Elves beside her,
// plus a Bear of mine and an Elf of seat 1's that must not be counted.
func gn3Imaryll(g *game.Game, elves int) uuid.UUID {
	me, opp := g.Seats[0], g.Seats[1]
	imaryll := apaPush(g, me.ID, me.ID, game.Card{
		Name: "Imaryll, Elfhame Elite", TypeLine: "Legendary Creature — Elf Warrior",
		OracleID: gn3ImaryllOracle, Power: 3, Toughness: 3,
	})
	for i := 0; i < elves; i++ {
		apaPush(g, me.ID, me.ID, game.Card{Name: "Elf Mystic", TypeLine: "Creature — Elf Druid", Power: 1, Toughness: 1})
	}
	apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Elf", TypeLine: "Creature — Elf Warrior", Power: 1, Toughness: 1})
	return imaryll
}

func TestGN3ImaryllGetsPlusOneForEachOtherElfYouControlWhenSheAttacks(t *testing.T) {
	g := newCatalogGame(t)
	defender := g.Seats[2]
	imaryll := gn3Imaryll(g, 2)
	if got := effectivePower(t, g, imaryll); got != 3 {
		t.Fatalf("Imaryll's power before she attacks is %d, want 3", got)
	}

	pr7bAttack(t, g, defender.ID, imaryll)
	if p, tough := effectivePower(t, g, imaryll), effectiveToughness(t, g, imaryll); p != 5 || tough != 5 {
		t.Fatalf("Imaryll is %d/%d after attacking with two other Elves, want 5/5 (not her, not a Bear, not an opponent's Elf)", p, tough)
	}
	pr7bDamageStep(t, g)
	if lost := 40 - defender.Life; lost != 5 {
		t.Errorf("the defender lost %d, want 5", lost)
	}
}

func TestGN3BlockedImaryllMayAssignHerDamageAsThoughUnblocked(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	imaryll := gn3Imaryll(g, 1)
	wall := mscWlBlocker(g, "Wall", 1, 3)
	mscWlBlockedCombat(t, g, []uuid.UUID{imaryll}, map[uuid.UUID]uuid.UUID{wall: imaryll})

	c := mscWlDamagePrompt(t, g)
	f := c.DamageAssignment
	if !f.AsThoughUnblocked || len(f.BlockerIDs) != 1 || f.AttackerPower != 4 {
		t.Fatalf("the prompt is %+v, want one blocker, power 4 (3 + one other Elf), as-though-unblocked", f)
	}
	if err := mscWlAnswer(t, g, c, 4, game.DamageAssignmentEntry{BlockerID: wall, Amount: 0}); err != nil {
		t.Fatalf("all of it to the player: %v", err)
	}
	if opp.Life != 36 {
		t.Errorf("the defending player has %d life, want 36", opp.Life)
	}
	if got := pr6Marked(g, wall); got != 0 {
		t.Errorf("the blocker was dealt %d damage, want none", got)
	}
}

func TestGN3BlockedImaryllMayStillAssignToHerBlocker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	imaryll := gn3Imaryll(g, 1)
	wall := mscWlBlocker(g, "Wall", 1, 3)
	mscWlBlockedCombat(t, g, []uuid.UUID{imaryll}, map[uuid.UUID]uuid.UUID{wall: imaryll})

	c := mscWlDamagePrompt(t, g)
	if err := mscWlAnswer(t, g, c, 0, game.DamageAssignmentEntry{BlockerID: wall, Amount: 4}); err != nil {
		t.Fatalf("all of it to the blocker: %v", err)
	}
	if opp.Life != 40 {
		t.Errorf("the defending player lost life (%d) though Imaryll assigned to the blocker", opp.Life)
	}
	if findBattlefieldCardForTest(g, wall) != nil {
		t.Error("the blocker survived 4 damage on toughness 3")
	}
}

// --- Maeve, Insidious Singer ----------------------------------------------

// gn3MaeveActivate pays {2}{U} and activates Maeve on `target`, then lets
// the ability resolve.
func gn3MaeveActivate(t *testing.T, g *game.Game, me *game.Player, maeve, target uuid.UUID) {
	t.Helper()
	b06AddMana(me, "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, maeve, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate Maeve: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestGN3MaeveDrawsWhenTheGoadedCreatureAttacksAnOpponentThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	maeve := b12Push(g, me.ID, "Maeve, Insidious Singer", "Legendary Creature — Siren", gn3MaeveOracle, 3, 4)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)

	// On the opponent's turn, before combat, Maeve goads their Bear.
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	passTo(t, g, me.ID)
	gn3MaeveActivate(t, g, me, maeve, bear)
	if b33Goaded(t, g, bear) != me.ID {
		t.Fatal("the target is not goaded by Maeve's controller")
	}

	hand := me.Hand.Size()
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(other, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Maeve's controller drew %d, want 1 (the goaded Bear attacked an opponent; the other Bear was never goaded)", got)
	}
}

func TestGN3MaeveDrawsOnlyThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	maeve := b12Push(g, me.ID, "Maeve, Insidious Singer", "Legendary Creature — Siren", gn3MaeveOracle, 3, 4)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Goaded on my own turn: the goad lasts until my next turn, the draw
	// watch only until this one ends.
	gn3MaeveActivate(t, g, me, maeve, bear)
	hand := me.Hand.Size()
	advanceToDeclareAttackersOf(t, g, 1)
	if b33Goaded(t, g, bear) != me.ID {
		t.Fatal("the goad ended before Maeve's controller's next turn")
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("Maeve's controller drew %d on a later turn, want 0 (\"this turn\")", got)
	}
}

// --- Vogar, Necropolis Tyrant ---------------------------------------------

func TestGN3VogarGrowsOnlyDuringYourTurnAndDrawsForItsCountersWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("the game starts on seat %d's turn, want seat 0's", g.Turn.ActiveSeat)
	}
	vogar := b12Push(g, me.ID, "Vogar, Necropolis Tyrant", "Legendary Creature — Zombie Giant", gn3VogarOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	later := b12Creature(g, opp.ID, "Their Later Bear", "Creature — Bear", 2, 2)

	for _, id := range []uuid.UUID{theirs, mine} {
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
		passPriorityAroundTable(t, g)
	}
	if got := counterCount(g, vogar, game.CounterPlusOne); got != 2 {
		t.Fatalf("two other creatures died during my turn: %d counters on Vogar, want 2", got)
	}

	advanceToNextSeatsTurn(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(later) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, vogar, game.CounterPlusOne); got != 2 {
		t.Fatalf("a creature died during an opponent's turn: %d counters on Vogar, want still 2", got)
	}

	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(vogar) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vogar) {
		t.Fatal("Vogar was not destroyed")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("Vogar died with two +1/+1 counters and its controller drew %d, want 2", got)
	}
}

// --- Zamriel, Seraph of Steel ---------------------------------------------

// gn3Equip attaches a plain Equipment controlled by `owner` to `creature`.
func gn3Equip(t *testing.T, g *game.Game, owner, creature uuid.UUID) {
	t.Helper()
	eq := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain Blade", TypeLine: equipTypeLine, Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(eq, game.TargetRef{Kind: game.TargetCard, ID: creature}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
}

func TestGN3ZamrielMakesYourEquippedCreaturesIndestructibleDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	zamriel := b12Push(g, me.ID, "Zamriel, Seraph of Steel", "Legendary Creature — Angel", gn3ZamrielOracle, 3, 4)
	equipped, bare, theirs := seedBear(g, me.ID), seedBear(g, me.ID), seedBear(g, opp.ID)
	gn3Equip(t, g, me.ID, equipped)
	gn3Equip(t, g, opp.ID, theirs)

	if !effectiveAbilitiesContain(t, g, equipped, "indestructible") {
		t.Error("an equipped creature I control lacks indestructible on my turn")
	}
	if effectiveAbilitiesContain(t, g, bare, "indestructible") {
		t.Error("an unequipped creature has indestructible")
	}
	if effectiveAbilitiesContain(t, g, theirs, "indestructible") {
		t.Error("an opponent's equipped creature has indestructible")
	}
	if effectiveAbilitiesContain(t, g, zamriel, "indestructible") {
		t.Error("an unequipped Zamriel has indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(equipped) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(equipped) {
		t.Fatal("the equipped creature was destroyed on my turn")
	}

	advanceToNextSeatsTurn(t, g)
	if effectiveAbilitiesContain(t, g, equipped, "indestructible") {
		t.Error("the equipped creature has indestructible outside my turn")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(equipped) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(equipped) {
		t.Error("the equipped creature survived a destroy on an opponent's turn")
	}
}
