package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav-h1 (#3172): the cards whose caveats the slice closed.

// Persist's -1/-1 counter is an "enters with" clause: Doubling Season
// sees it ride the entry and doubles it, which a counter added "a beat
// later" by a separate step would also do but an entry replacement
// pins to the entry itself (the creature is never on the battlefield
// without it).
func TestPersistCounterRidesTheEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Doubling Season", "Enchantment", oracleDoublingSeason, false)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Dead Ogre", TypeLine: "Creature — Ogre",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Persist", "Sorcery", b12PersistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Ogre should be back")
	}
	if got := b12Counter(t, g, dead, "-1/-1"); got != 2 {
		t.Errorf("-1/-1 counters = %d, want 2 (Doubling Season doubles the entry counter)", got)
	}
	if spec, _ := Lookup(b12PersistOracle); spec.Completeness != CompletenessFull {
		t.Errorf("Persist completeness = %v, want Full", spec.Completeness)
	}
}

// Makeshift Mannequin: a creature that asks a question as it enters (a
// Clone's copy pick) still comes back with its counter and the granted
// sacrifice trigger, because the grant is the entry's tail.
func TestMakeshiftMannequinGrantLandsAfterACloneAnswersItsEntryQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Makeshift Mannequin", "Instant", rtMannequinOracle, ltCardTarget(clone))
	resolveWithCopyChoice(t, g, bears)
	c := layeredCard(t, g, clone)
	if c.Counters["mannequin"] != 1 {
		t.Fatalf("the Clone returned with counters %v, want a mannequin counter", c.Counters)
	}
	endo := pushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", rtEndoskeletonOracle, false)
	rtActivate(t, g, me.ID, endo, 0, ltCardTarget(clone), "C", "C")
	passPriorityAroundTable(t, g)
	if onBattlefield(g, clone) {
		t.Error("the returned Clone was targeted and is still on the battlefield: the sacrifice grant never landed")
	}
	if spec, _ := Lookup(rtMannequinOracle); spec.Completeness != CompletenessFull {
		t.Errorf("Makeshift Mannequin completeness = %v, want Full", spec.Completeness)
	}
}

// Presumed Dead: a creature that comes back through a copy question is
// still suspected, because the suspect is the entry's tail.
func TestPresumedDeadSuspectsACloneThatAsksAsItReturns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bears)
	castCatalogSpell(t, g, "Presumed Dead", "Instant", sbPresumedDead, dgCardRef(clone))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(clone) })
	resolveWithCopyChoice(t, g, bears)
	c := layeredCard(t, g, clone)
	if !c.Suspected {
		t.Error("the returned Clone is not suspected")
	}
	if c.Tapped {
		t.Error("Presumed Dead returns the creature untapped")
	}
}

// Mondrak: {1}{W/P}{W/P}, sacrifice two other artifacts and/or creatures
// puts an indestructible counter on it.
func TestMondrakIndestructibleCounterAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mondrak := pushCatalogPermanent(g, me.ID, "Mondrak, Glory Dominus", "Legendary Creature — Phyrexian Horror", mondrakOracle, false)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, mondrak, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock, land}, PhyrexianLife: 2,
	}); err == nil {
		t.Fatal("a land is neither an artifact nor a creature and can't be sacrificed")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, mondrak, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock, bear}, PhyrexianLife: 2,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := life - me.Life; got != 2*game.PhyrexianLifePerSymbol {
		t.Errorf("life paid = %d, want %d for two Phyrexian symbols", got, 2*game.PhyrexianLifePerSymbol)
	}
	if counterCount(g, mondrak, "indestructible") != 1 {
		t.Error("Mondrak should have an indestructible counter")
	}
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(bear) {
		t.Error("both sacrificed permanents should be gone")
	}
	if spec, _ := Lookup(mondrakOracle); spec.Completeness != CompletenessFull {
		t.Errorf("Mondrak completeness = %v, want Full", spec.Completeness)
	}
}

// Roiling Canopy: the intervening-if is judged again as the ability
// resolves against the Forests OTHER than the one that entered, so the
// entered Forest leaving in response changes nothing while five others
// stay.
func TestRoilingCanopyStillPumpsWhenTheEnteredForestLeavesInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := rfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	b12PlayFromHand(t, g, "Roiling Canopy", "Land", rfmaRoilingCanopy, game.CastSpellParams{})
	rfmaForests(g, me.ID, 5)
	advanceToMainOf(t, g, 0)
	entered := b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if err := sgAnswerTargets(t, g, bear); err != nil {
		t.Fatalf("target: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(entered) })
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, bear); p != 5 || th != 5 {
		t.Errorf("P/T = %d/%d, want 5/5: five other Forests remained", p, th)
	}
}

// ...and when a different Forest leaves, leaving only four others, it
// does not.
func TestRoilingCanopyDoesNothingWhenFiveOthersAreNoLongerThere(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := rfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	b12PlayFromHand(t, g, "Roiling Canopy", "Land", rfmaRoilingCanopy, game.CastSpellParams{})
	rfmaForests(g, me.ID, 5)
	advanceToMainOf(t, g, 0)
	entered := b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if err := sgAnswerTargets(t, g, bear); err != nil {
		t.Fatalf("target: %v", err)
	}
	var other uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == me.ID && c.InstanceID != entered {
			other = c.InstanceID
			break
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, bear); p != 2 || th != 2 {
		t.Errorf("P/T = %d/%d, want 2/2: only four others remained", p, th)
	}
}

// Ratatwotwo: X is its power as last it existed when it has left by the
// time the bonus resolves (CR 608.2h).
func TestRatatwotwoBonusUsesLastKnownPowerWhenItHasLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := pushRatatwotwo(g, me.ID)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	giveRole(t, g, RoleChef, bear, me.ID)
	declareAttack(t, g, opp.ID, bear)
	answerTriggerOrderInOfferedOrder(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rat) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rat) {
		t.Fatal("Ratatwotwo should be gone")
	}
	// 2/2 +1/+1 (Chef Role) +2/+2 (Ratatwotwo's last power).
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("enchanted attacker has power %d, want 5", p)
	}
}

// Rosie Cotton: "other than Rosie Cotton" is object identity, so a
// second creature that is also named Rosie (a token copy) is a legal
// target for the counter.
func TestRosieCottonCanPutTheCounterOnAnotherRosie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	copyOfRosie := seedCreature(g, "Rosie Cotton of South Lane", me.ID)
	rosie := castAndResolveCreature(t, g, "Rosie Cotton of South Lane", "Legendary Creature — Halfling Peasant", b11RosieCottonOracle)
	passPriorityAroundTable(t, g)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the token trigger must ask which creature gets the counter")
	}
	if hasID(prompt.PickTargetCards, rosie) {
		t.Error("Rosie offered herself as the target")
	}
	if !hasID(prompt.PickTargetCards, copyOfRosie) {
		t.Errorf("another creature named Rosie Cotton must be a legal target: %v", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, copyOfRosie)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, copyOfRosie, "+1/+1"); got != 1 {
		t.Errorf("the other Rosie has %d +1/+1 counters, want 1", got)
	}
}

// Shark Typhoon: "when you cycle this card" is its own trigger, from the
// graveyard, reading the X the cycling cost paid. It goes on the stack
// above the cycling ability, so the Shark is there before the card is
// drawn.
func TestSharkTyphoonCyclingTriggerResolvesAboveTheCyclingAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	id := pushCatalogHandCard(me, "Shark Typhoon", "Enchantment", rfrbSharkTyphoon)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{U}"); err != nil {
		t.Fatal(err)
	}
	pushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	// Resolve one stack object at a time until the Shark exists.
	for i := 0; i < 16 && findBattlefieldByName(g, "Shark") == uuid.Nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	sh := findBattlefieldByName(g, "Shark")
	if sh == uuid.Nil {
		t.Fatal("no Shark")
	}
	if len(me.Hand.Cards) != hand-1 {
		t.Errorf("hand = %d, want %d: the Shark came first and the draw is still on the stack", len(me.Hand.Cards), hand-1)
	}
	if c := s58p6View(t, g, sh); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand = %d after the cycle resolved, want %d", len(me.Hand.Cards), hand)
	}
}

// Mycosynth Lattice: permanents are colorless.
func TestMycosynthLatticeMakesPermanentsColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if c := layeredCard(t, g, elf); len(c.EffectiveColors()) != 1 {
		t.Fatalf("setup: the Elf should be green, got %v", c.EffectiveColors())
	}
	pushCatalogPermanent(g, me.ID, "Mycosynth Lattice", "Artifact", "ae1f2ab5-c6a5-4d49-a746-3cb4668bf805", false)
	c := layeredCard(t, g, elf)
	if got := c.EffectiveColors(); len(got) != 0 {
		t.Errorf("colors with the Lattice out = %v, want none", got)
	}
	if !c.IsArtifact() {
		t.Error("the Elf is also an artifact")
	}
}

// Phantasmal Image: a card that went through deck import has a printing,
// so its printed 0 toughness is real and declining the copy kills it.
func TestPhantasmalImageThatDeclinesTheCopyDies(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Phantasmal Image", TypeLine: "Creature — Illusion", OracleID: oraclePhantasmalImage,
		ScryfallID: "00000000-0000-0000-0000-0000000000a1", Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	resolveWithCopyChoice(t, g, uuid.Nil)
	if g.Battlefield.Contains(id) {
		t.Error("a declined Phantasmal Image is a 0/0 and dies")
	}
	if !active.Graveyard.Contains(id) {
		t.Error("it should be in its owner's graveyard")
	}
}
