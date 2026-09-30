package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const sorinOfHouseMarkovOracle = "84e84532-498b-40ea-9510-46fb2403939b"

// sorinCard is the double-faced card as the deck importer builds it,
// front face up.
func sorinCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   sorinOfHouseMarkovOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Sorin of House Markov", TypeLine: "Legendary Creature — Human Noble", ManaCost: "{1}{B}",
				Colors: []string{"B"}, Power: 1, Toughness: 4},
			{Name: "Sorin, Ravenous Neonate", TypeLine: "Legendary Planeswalker — Sorin",
				Colors: []string{"B"}, StartingLoyalty: 3},
		},
	}
	c.SetFace(0)
	return c
}

// pushSorinWalker puts the back face on the battlefield with `loyalty`.
func pushSorinWalker(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return pushWalkerForTest(g, owner, "Sorin, Ravenous Neonate", sorinOfHouseMarkovOracle+"#1", loyalty)
}

// Extort: casting a spell offers {W/B}; paying drains each opponent
// for 1 and gains the total.
func TestSorinOfHouseMarkovExtorts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sorin := pushBattlefieldCardWithTimestamp(g, sorinCard(me.ID))
	if !hasAbility(effectiveAbilities(t, g, sorin), "lifelink") {
		t.Error("printed lifelink")
	}
	lives := b17Life(g)

	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("casting a spell offers the extort payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	answerPayUnless(t, g, me.ID, true)

	opponents := len(g.Seats) - 1
	for i, p := range g.Seats {
		want := lives[i] - 1
		if i == 0 {
			want = lives[i] + opponents
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
}

// Declining the payment does nothing.
func TestSorinOfHouseMarkovExtortDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, sorinCard(me.ID))
	lives := b17Life(g)

	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	answerPayUnless(t, g, me.ID, false)
	for i, p := range g.Seats {
		if p.Life != lives[i] {
			t.Errorf("seat %d life changed on a declined extort: %d → %d", i, lives[i], p.Life)
		}
	}
}

// Three life gained by the postcombat main phase: Sorin is exiled and
// returns transformed — a new object, a planeswalker with loyalty 3.
func TestSorinOfHouseMarkovTransformsAfterGainingThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sorin := pushBattlefieldCardWithTimestamp(g, sorinCard(me.ID))
	advanceToMain(t, g)
	e2ChangeLife(g, me.ID, 3)

	advanceToStepInTurn(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)

	var walker *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Sorin, Ravenous Neonate" {
			walker = c
		}
	}
	if walker == nil {
		t.Fatal("Sorin returns transformed into Sorin, Ravenous Neonate")
	}
	if !walker.IsPlaneswalker() {
		t.Errorf("the back face is a planeswalker: %q", walker.TypeLine)
	}
	if got := walker.Counters[game.CounterLoyalty]; got != 3 {
		t.Errorf("loyalty %d, want the printed 3", got)
	}
	if walker.Owner != me.ID || walker.Controller != me.ID {
		t.Error("under his owner's control")
	}
	if c, ok := battlefieldCard(g, sorin); ok && c.Name == "Sorin of House Markov" {
		t.Error("the creature face is gone from the battlefield")
	}
}

// Two life is not enough: no trigger, the creature stays.
func TestSorinOfHouseMarkovStaysWithLessThanThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sorin := pushBattlefieldCardWithTimestamp(g, sorinCard(me.ID))
	advanceToMain(t, g)
	e2ChangeLife(g, me.ID, 2)

	advanceToStepInTurn(t, g, game.StepPostcombatMain)
	if triggersOnStackFrom(g, sorin) != 0 {
		t.Fatal("the intervening if keeps the ability from triggering")
	}
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, sorin)
	if !ok || got.Name != "Sorin of House Markov" {
		t.Error("Sorin stays a creature")
	}
}

// +2: a Food token.
func TestSorinRavenousNeonatePlusTwoMakesFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := pushSorinWalker(g, me.ID, 3)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Food", me.ID); n != 1 {
		t.Errorf("Food tokens: %d, want 1", n)
	}
	if got := loyaltyOf(g, walker); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

// −1: damage equal to the life gained this turn, read on resolution.
func TestSorinRavenousNeonateMinusOneDealsTheLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushSorinWalker(g, me.ID, 3)
	advanceToMain(t, g)
	e2ChangeLife(g, me.ID, 4)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("−1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 4 {
		t.Errorf("damage %d, want the 4 life gained this turn", got)
	}
}

// −6 with another white permanent: the creature is yours, a Vampire,
// and has a lifelink counter that grants lifelink while Sorin is here.
// Loyalty 7, so Sorin survives the −6 and his static is still there.
func TestSorinRavenousNeonateMinusSixStealsAndGivesLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushSorinWalker(g, me.ID, 7)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "White Knight", TypeLine: "Creature — Human Knight",
		Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	target := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, walker, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	passPriorityAroundTable(t, g)

	got, ok := battlefieldCard(g, target)
	if !ok {
		t.Fatal("the stolen creature is on the battlefield")
	}
	if got.Controller != me.ID {
		t.Error("gain control of target creature")
	}
	if !containsString(effectiveSubtypes(t, g, target), "Vampire") || !containsString(effectiveSubtypes(t, g, target), "Bear") {
		t.Errorf("a Vampire in addition to its other types: %v", effectiveSubtypes(t, g, target))
	}
	if got.Counters["lifelink"] != 1 {
		t.Errorf("a lifelink counter: %d", got.Counters["lifelink"])
	}
	if !hasAbility(effectiveAbilities(t, g, target), "lifelink") {
		t.Error("the lifelink counter grants lifelink while Sorin is on the battlefield")
	}

	// The declared caveat: with Sorin gone the counter stays and stops
	// granting lifelink.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(walker) })
	passPriorityAroundTable(t, g)
	if hasAbility(effectiveAbilities(t, g, target), "lifelink") {
		t.Error("the caveat: without Sorin the counter grants nothing")
	}
}

// −6 with no other white permanent: no counter.
func TestSorinRavenousNeonateMinusSixWithoutWhiteGivesNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushSorinWalker(g, me.ID, 6)
	target := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Paladin", TypeLine: "Creature — Human Knight",
		Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, walker, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	passPriorityAroundTable(t, g)

	got, _ := battlefieldCard(g, target)
	if got.Controller != me.ID {
		t.Error("gain control of target creature")
	}
	if got.Counters["lifelink"] != 0 {
		t.Error("the stolen white creature itself does not count as the other white permanent")
	}
}
