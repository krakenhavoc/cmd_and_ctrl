package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deification_test.go — #5646: Deification, clause by clause, on the
// engine's own paths: the cast, the as-enters prompt, the layer pass,
// the targeting gate and the damage pipeline.

const deificationOracle = "8c8f93d7-d89a-401b-8764-76be71ad16e3"

// castDeification casts Deification from seat 0's hand, resolves it and
// answers its as-enters prompt with `pwType`. It returns the permanent.
func castDeification(t *testing.T, g *game.Game, pwType string) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	toMainForCost(t, g)
	floatMana(t, g, me, "{W}{W}")
	id := castFromHandForTest(t, g, me, "Deification", "Enchantment", "{1}{W}", deificationOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	answerAnchorWord(t, g, me.ID, pwType)
	if !onBattlefieldNow(g, id) {
		t.Fatal("Deification is not on the battlefield")
	}
	return id
}

// walkerTargetableBy reports whether `by` may target the planeswalker.
func walkerTargetableBy(g *game.Game, by, id uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() {
		ok = slices.Contains(g.LegalTargetsForEffect(game.TargetSource{Controller: by}, TargetPermanent("target permanent")).Cards, id)
	})
	return ok
}

// "As this enchantment enters, choose a planeswalker type": every CR
// 205.3j type is offered, the types of the planeswalkers you control
// first, and the answer is stored on the permanent.
func TestDeificationChoosesAPlaneswalkerTypeAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fixtureWalker(g, me.ID, pw2797WalkerOracle, "Legendary Planeswalker — Teferi", 3)
	toMainForCost(t, g)
	floatMana(t, g, me, "{W}{W}")
	id := castFromHandForTest(t, g, me, "Deification", "Enchantment", "{1}{W}", deificationOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)

	words := pickOptionWords(t, g, me.ID)
	if len(words) != len(game.PlaneswalkerTypes()) {
		t.Fatalf("%d types offered, want all %d", len(words), len(game.PlaneswalkerTypes()))
	}
	if words[0] != "Teferi" {
		t.Errorf("first option %q, want the type of the planeswalker I control", words[0])
	}
	answerAnchorWord(t, g, me.ID, "Teferi")
	var chosen string
	g.ReadSnapshot(func() { chosen = g.ChosenOptionOf(id) })
	if chosen != "Teferi" {
		t.Errorf("chosen type %q, want Teferi", chosen)
	}
}

// "Planeswalkers you control of the chosen type have hexproof": mine of
// that type, not mine of another type, not an opponent's, and not before
// the choice is answered.
func TestDeificationGivesYourPlaneswalkersOfTheChosenTypeHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teferi := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Legendary Planeswalker — Teferi", 3)
	jace := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Legendary Planeswalker — Jace", 3)
	theirs := fixtureWalker(g, opp.ID, pw2797WalkerOracle, "Legendary Planeswalker — Teferi", 3)

	toMainForCost(t, g)
	floatMana(t, g, me, "{W}{W}")
	castFromHandForTest(t, g, me, "Deification", "Enchantment", "{1}{W}", deificationOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if hasAbility(effectiveAbilities(t, g, teferi), "hexproof") {
		t.Error("a planeswalker has hexproof before any type is chosen")
	}
	answerAnchorWord(t, g, me.ID, "Teferi")

	if !hasAbility(effectiveAbilities(t, g, teferi), "hexproof") {
		t.Error("my Teferi lacks hexproof")
	}
	if hasAbility(effectiveAbilities(t, g, jace), "hexproof") {
		t.Error("my Jace has hexproof: it is not of the chosen type")
	}
	if hasAbility(effectiveAbilities(t, g, theirs), "hexproof") {
		t.Error("an opponent's Teferi has hexproof")
	}
	if walkerTargetableBy(g, opp.ID, teferi) {
		t.Error("an opponent may target my hexproof Teferi")
	}
	if !walkerTargetableBy(g, me.ID, teferi) {
		t.Error("I may not target my own hexproof Teferi")
	}
	if !walkerTargetableBy(g, opp.ID, jace) {
		t.Error("an opponent may not target my Jace")
	}
	// A planeswalker of the chosen type that arrives later is covered.
	late := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Legendary Planeswalker — Teferi", 2)
	if !hasAbility(effectiveAbilities(t, g, late), "hexproof") {
		t.Error("a Teferi that arrived later lacks hexproof")
	}
}

// The last line: damage that would remove every loyalty counter removes
// all but one, while I control a creature. The damage is dealt in full,
// so a lifelink source gains all of it.
func TestDeificationDamageLeavesOneLoyaltyCounterWhileYouControlACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teferi := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Legendary Planeswalker — Teferi", 4)
	jace := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Legendary Planeswalker — Jace", 4)
	theirs := fixtureWalker(g, opp.ID, pw2797WalkerOracle, "Legendary Planeswalker — Teferi", 4)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	linker := pushKeywordCreature(g, opp.ID, "Linker", "lifelink")
	castDeification(t, g, "Teferi")

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(linker, teferi, 3) })
	if got := loyaltyCount(g, teferi); got != 1 {
		t.Fatalf("3 damage to a 4-loyalty Teferi left %d, want 1: it would not remove them all", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(teferi, game.CounterLoyalty, 3) })
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(linker, teferi, 6) })
	runStateChecks(g)
	if got := loyaltyCount(g, teferi); got != 1 {
		t.Fatalf("6 damage to a 4-loyalty Teferi left %d loyalty, want 1", got)
	}
	if !onBattlefieldNow(g, teferi) {
		t.Fatal("Teferi left the battlefield")
	}
	if opp.Life != before+6 {
		t.Errorf("the lifelink source's controller gained %d, want 6: the damage is dealt in full", opp.Life-before)
	}
	// At one counter, more damage removes none.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(linker, teferi, 2) })
	runStateChecks(g)
	if got := loyaltyCount(g, teferi); got != 1 || !onBattlefieldNow(g, teferi) {
		t.Fatalf("a Teferi at 1 loyalty dealt 2 has %d (on battlefield %v), want 1 and true", got, onBattlefieldNow(g, teferi))
	}

	// Another type, and an opponent's planeswalker of the chosen type,
	// are not covered.
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(linker, jace, 4)
		_ = g.DealDamageToCreatureForEffect(bear, theirs, 4)
	})
	runStateChecks(g)
	if onBattlefieldNow(g, jace) {
		t.Error("my Jace survived lethal damage: it is not of the chosen type")
	}
	if onBattlefieldNow(g, theirs) {
		t.Error("an opponent's Teferi survived lethal damage")
	}

	// "As long as you control a creature": without one, the damage takes
	// every counter and the planeswalker goes to the graveyard.
	removeFromBattlefield(g, bear)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(linker, teferi, 1) })
	runStateChecks(g)
	if onBattlefieldNow(g, teferi) {
		t.Error("Teferi survived lethal damage while I controlled no creature")
	}
}

// Only damage is replaced: a loyalty cost takes the last counters as it
// always does, and the planeswalker is put into the graveyard.
func TestDeificationDoesNotChangeLoyaltyCosts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	teferi := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Legendary Planeswalker — Teferi", 3)
	pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	castDeification(t, g, "Teferi")

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate −3: %v", err)
	}
	runStateChecks(g)
	if onBattlefieldNow(g, teferi) {
		t.Error("a Teferi that paid its last 3 loyalty is still on the battlefield")
	}
}

// The chosen type rides the snapshot: in a restored game the hexproof and
// the damage cap are both still there.
func TestDeificationSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teferi := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Legendary Planeswalker — Teferi", 4)
	pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	hitter := pushVanillaCreature(g, opp.ID, "Hitter", 2, 2)
	castDeification(t, g, "Teferi")

	r := restoreRoundTrip(t, g, true)
	if !hasAbility(effectiveAbilities(t, r, teferi), "hexproof") {
		t.Error("Teferi lost hexproof in the restored game")
	}
	r.WithWriteLock(func() { _ = r.DealDamageToCreatureForEffect(hitter, teferi, 9) })
	runStateChecks(r)
	if got := loyaltyCount(r, teferi); got != 1 || !onBattlefieldNow(r, teferi) {
		t.Errorf("restored game: loyalty %d (on battlefield %v), want 1 and true", got, onBattlefieldNow(r, teferi))
	}
}
