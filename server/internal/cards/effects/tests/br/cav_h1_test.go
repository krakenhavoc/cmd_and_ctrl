package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav-h1 (#3172): the cards whose caveats the slice closed.

// Persist's -1/-1 counter is an "enters with" clause: Doubling Season
// sees it ride the entry and doubles it, which a counter added "a beat
// later" by a separate step would also do but an entry replacement
// pins to the entry itself (the creature is never on the battlefield
// without it).
func TestPersistCounterRidesTheEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Doubling Season", "Enchantment", effectstest.OracleDoublingSeason, false)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Dead Ogre", TypeLine: "Creature — Ogre",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Persist", "Sorcery", effectstest.B12PersistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Ogre should be back")
	}
	if got := effectstest.B12Counter(t, g, dead, "-1/-1"); got != 2 {
		t.Errorf("-1/-1 counters = %d, want 2 (Doubling Season doubles the entry counter)", got)
	}
	if spec, _ := effects.Lookup(effectstest.B12PersistOracle); spec.Completeness != effects.CompletenessFull {
		t.Errorf("Persist completeness = %v, want Full", spec.Completeness)
	}
}

// Makeshift Mannequin: a creature that asks a question as it enters (a
// Clone's copy pick) still comes back with its counter and the granted
// sacrifice trigger, because the grant is the entry's tail.
func TestMakeshiftMannequinGrantLandsAfterACloneAnswersItsEntryQuestion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bears := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: effectstest.OracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Makeshift Mannequin", "Instant", effectstest.RtMannequinOracle, effectstest.LtCardTarget(clone))
	effectstest.ResolveWithCopyChoice(t, g, bears)
	c := effectstest.LayeredCard(t, g, clone)
	if c.Counters["mannequin"] != 1 {
		t.Fatalf("the Clone returned with counters %v, want a mannequin counter", c.Counters)
	}
	endo := effectstest.PushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", effectstest.RtEndoskeletonOracle, false)
	effectstest.RtActivate(t, g, me.ID, endo, 0, effectstest.LtCardTarget(clone), "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, clone) {
		t.Error("the returned Clone was targeted and is still on the battlefield: the sacrifice grant never landed")
	}
	if spec, _ := effects.Lookup(effectstest.RtMannequinOracle); spec.Completeness != effects.CompletenessFull {
		t.Errorf("Makeshift Mannequin completeness = %v, want Full", spec.Completeness)
	}
}

// Presumed Dead: a creature that comes back through a copy question is
// still suspected, because the suspect is the entry's tail.
func TestPresumedDeadSuspectsACloneThatAsksAsItReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bears := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, bears)
	effectstest.CastCatalogSpell(t, g, "Presumed Dead", "Instant", effectstest.SbPresumedDead, effectstest.DgCardRef(clone))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(clone) })
	effectstest.ResolveWithCopyChoice(t, g, bears)
	c := effectstest.LayeredCard(t, g, clone)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mondrak := effectstest.PushCatalogPermanent(g, me.ID, "Mondrak, Glory Dominus", "Legendary Creature — Phyrexian Horror", effectstest.MondrakOracle, false)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
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
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 2*game.PhyrexianLifePerSymbol {
		t.Errorf("life paid = %d, want %d for two Phyrexian symbols", got, 2*game.PhyrexianLifePerSymbol)
	}
	if effectstest.CounterCount(g, mondrak, "indestructible") != 1 {
		t.Error("Mondrak should have an indestructible counter")
	}
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(bear) {
		t.Error("both sacrificed permanents should be gone")
	}
	if spec, _ := effects.Lookup(effectstest.MondrakOracle); spec.Completeness != effects.CompletenessFull {
		t.Errorf("Mondrak completeness = %v, want Full", spec.Completeness)
	}
}

// Roiling Canopy: the intervening-if is judged again as the ability
// resolves against the Forests OTHER than the one that entered, so the
// entered Forest leaving in response changes nothing while five others
// stay.
func TestRoilingCanopyStillPumpsWhenTheEnteredForestLeavesInResponse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.RfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.B12PlayFromHand(t, g, "Roiling Canopy", "Land", effectstest.RfmaRoilingCanopy, game.CastSpellParams{})
	effectstest.RfmaForests(g, me.ID, 5)
	effectstest.AdvanceToMainOf(t, g, 0)
	entered := effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if err := effectstest.SgAnswerTargets(t, g, bear); err != nil {
		t.Fatalf("target: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(entered) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.RfPT(t, g, bear); p != 5 || th != 5 {
		t.Errorf("P/T = %d/%d, want 5/5: five other Forests remained", p, th)
	}
}

// ...and when a different Forest leaves, leaving only four others, it
// does not.
func TestRoilingCanopyDoesNothingWhenFiveOthersAreNoLongerThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.RfmaPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.B12PlayFromHand(t, g, "Roiling Canopy", "Land", effectstest.RfmaRoilingCanopy, game.CastSpellParams{})
	effectstest.RfmaForests(g, me.ID, 5)
	effectstest.AdvanceToMainOf(t, g, 0)
	entered := effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if err := effectstest.SgAnswerTargets(t, g, bear); err != nil {
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
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.RfPT(t, g, bear); p != 2 || th != 2 {
		t.Errorf("P/T = %d/%d, want 2/2: only four others remained", p, th)
	}
}

// Ratatwotwo: X is its power as last it existed when it has left by the
// time the bonus resolves (CR 608.2h).
func TestRatatwotwoBonusUsesLastKnownPowerWhenItHasLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.PushRatatwotwo(g, me.ID)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleChef, bear, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rat) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rat) {
		t.Fatal("Ratatwotwo should be gone")
	}
	// 2/2 +1/+1 (Chef Role) +2/+2 (Ratatwotwo's last power).
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("enchanted attacker has power %d, want 5", p)
	}
}

// Rosie Cotton: "other than Rosie Cotton" is object identity, so a
// second creature that is also named Rosie (a token copy) is a legal
// target for the counter.
func TestRosieCottonCanPutTheCounterOnAnotherRosie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	copyOfRosie := effectstest.SeedCreature(g, "Rosie Cotton of South Lane", me.ID)
	rosie := effectstest.CastAndResolveCreature(t, g, "Rosie Cotton of South Lane", "Legendary Creature — Halfling Peasant", effectstest.B11RosieCottonOracle)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the token trigger must ask which creature gets the counter")
	}
	if effectstest.HasID(prompt.PickTargetCards, rosie) {
		t.Error("Rosie offered herself as the target")
	}
	if !effectstest.HasID(prompt.PickTargetCards, copyOfRosie) {
		t.Errorf("another creature named Rosie Cotton must be a legal target: %v", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, copyOfRosie)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, copyOfRosie, "+1/+1"); got != 1 {
		t.Errorf("the other Rosie has %d +1/+1 counters, want 1", got)
	}
}

// Shark Typhoon: "when you cycle this card" is its own trigger, from the
// graveyard, reading the X the cycling cost paid. It goes on the stack
// above the cycling ability, so the Shark is there before the card is
// drawn.
func TestSharkTyphoonCyclingTriggerResolvesAboveTheCyclingAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogHandCard(me, "Shark Typhoon", "Enchantment", effectstest.RfrbSharkTyphoon)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{U}"); err != nil {
		t.Fatal(err)
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	// Resolve one stack object at a time until the Shark exists.
	for i := 0; i < 16 && effectstest.FindBattlefieldByName(g, "Shark") == uuid.Nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	sh := effectstest.FindBattlefieldByName(g, "Shark")
	if sh == uuid.Nil {
		t.Fatal("no Shark")
	}
	if len(me.Hand.Cards) != hand-1 {
		t.Errorf("hand = %d, want %d: the Shark came first and the draw is still on the stack", len(me.Hand.Cards), hand-1)
	}
	if c := effectstest.S58p6View(t, g, sh); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand = %d after the cycle resolved, want %d", len(me.Hand.Cards), hand)
	}
}

// Mycosynth Lattice: permanents are colorless.
func TestMycosynthLatticeMakesPermanentsColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if c := effectstest.LayeredCard(t, g, elf); len(c.EffectiveColors()) != 1 {
		t.Fatalf("setup: the Elf should be green, got %v", c.EffectiveColors())
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Mycosynth Lattice", "Artifact", "ae1f2ab5-c6a5-4d49-a746-3cb4668bf805", false)
	c := effectstest.LayeredCard(t, g, elf)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Phantasmal Image", TypeLine: "Creature — Illusion", OracleID: effectstest.OraclePhantasmalImage,
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
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)
	if g.Battlefield.Contains(id) {
		t.Error("a declined Phantasmal Image is a 0/0 and dies")
	}
	if !active.Graveyard.Contains(id) {
		t.Error("it should be in its owner's graveyard")
	}
}
