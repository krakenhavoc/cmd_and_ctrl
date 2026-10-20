package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guffc_test.go — the Guffster's Party deck request (#4184), slice c:
// Onakke Oathkeeper, Teyo, Sarkhan the Masterless, Narset, Enlightened
// Master, Vronos, Final Showdown and Guff Rewrites History.

const (
	guffcOnakkeOracle = "080df9eb-c26b-494a-96b4-d7e47e501866"
	guffcTeyoOracle   = "0dfdf1a1-a3a9-490f-bd9d-dcba088c6785"
	guffcSarkhan      = "67dbc5dd-6d05-43e7-83f7-70e10eb54807"
	guffcNarsetOracle = "aeaf7768-ed67-487d-8506-e667babb88ba"
	guffcVronosOracle = "86851bdb-9b2f-4407-97a1-a6a9f2e0ef66"
	guffcShowdown     = "7e7ec3d6-a84f-4cc3-93f4-4d181d41e126"
	guffcRewrites     = "7ff7af5f-e4e3-40af-8882-756564bfa479"
)

func TestGuffcSliceCompleteness(t *testing.T) {
	full := []string{guffcOnakkeOracle, guffcTeyoOracle, guffcSarkhan, guffcNarsetOracle, guffcVronosOracle, guffcShowdown, guffcRewrites}
	for _, oracle := range full {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	// Teyo's −2 shipped with ADR 0206 (#4197): both loyalty rows.
	if spec, _ := Lookup(guffcTeyoOracle); len(spec.Activated) != 2 {
		t.Errorf("Teyo has %d activated rows, want +1 and −2", len(spec.Activated))
	}
}

func guffcWalker(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string, loyalty int) uuid.UUID {
	t.Helper()
	id := pushCatalogPermanent(g, owner, name, typeLine, oracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	return id
}

func guffcRefs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

// --- Onakke Oathkeeper ---------------------------------------------

// Attacking a planeswalker the Oathkeeper's controller controls costs
// {1} per creature; attacking the player costs nothing.
func TestGuffcOnakkeTaxesAttacksOnPlaneswalkersOnly(t *testing.T) {
	g, bear, opp := taxTable(t)
	walker := pushCatalogPermanent(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "test-walker-onakke", false)
	pushCatalogPermanent(g, opp.ID, "Onakke Oathkeeper", "Creature — Ogre Spirit", guffcOnakkeOracle, false)

	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); !p.IsFree() {
		t.Errorf("attacking the player cost %q; Onakke taxes only attacks on planeswalkers", p.Cost)
	}
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: walker}}); p.Total.Generic != 1 {
		t.Errorf("attacking the planeswalker cost %q, want {1}", p.Cost)
	}
}

// The graveyard ability exiles the Oathkeeper as its cost and returns a
// planeswalker card; a creature card is not a legal target.
func TestGuffcOnakkeReturnsAPlaneswalkerCardFromTheGraveyard(t *testing.T) {
	g, me, _ := spendTable(t)
	oath := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: oath, Name: "Onakke Oathkeeper", TypeLine: "Creature — Ogre Spirit",
		OracleID: guffcOnakkeOracle, Owner: me.ID, Controller: me.ID})
	walker := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: walker, Name: "Old Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 3, Owner: me.ID, Controller: me.ID})
	bear := pushGraveyardCardForTest(me, "Dead Bear")

	apaMana(me, "W", "W", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, oath, 0, game.ActivateAbilityParams{Targets: guffcRefs(bear)}); err == nil {
		t.Fatal("a creature card was accepted as a planeswalker card")
	}
	b16Activate(t, g, me.ID, oath, 0, game.ActivateAbilityParams{Targets: guffcRefs(walker)})
	if !g.Exile.Contains(oath) {
		t.Error("the Oathkeeper was not exiled for the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(walker) {
		t.Fatal("the planeswalker card did not return to the battlefield")
	}
}

// --- Teyo, Geometric Tactician -------------------------------------

func TestGuffcTeyoMakesAFlyingWall(t *testing.T) {
	g, _, _ := spendTable(t)
	castCatalogSpell(t, g, "Teyo, Geometric Tactician", "Legendary Planeswalker — Teyo", guffcTeyoOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	var wall game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wall" {
			wall = c
		}
	}
	if wall.InstanceID == uuid.Nil {
		t.Fatal("no Wall token")
	}
	eff := effectiveCardForTest(t, g, wall.InstanceID)
	if eff.CurrentPower() != 0 || eff.CurrentToughness() != 4 || !game.HasKeyword(&eff, "defender") || !game.HasKeyword(&eff, "flying") {
		t.Errorf("the Wall is %d/%d keywords %v, want a 0/4 with defender and flying", eff.CurrentPower(), eff.CurrentToughness(), eff.Keywords)
	}
	if got := effectiveColors(t, g, wall.InstanceID); len(got) != 1 {
		t.Errorf("the Wall's colours are %v, want white", got)
	}
}

func TestGuffcTeyoPlusOneDrawsForYouAndTheOpponent(t *testing.T) {
	g, me, opp := spendTable(t)
	teyo := guffcWalker(t, g, me.ID, "Teyo, Geometric Tactician", "Legendary Planeswalker — Teyo", guffcTeyoOracle, 3)
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Instant", Owner: me.ID})
	pushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Instant", Owner: opp.ID})
	meHand, oppHand := me.Hand.Size(), opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, teyo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("+1 targeted its controller; it says target opponent")
	}
	b16Activate(t, g, me.ID, teyo, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != meHand+1 || opp.Hand.Size() != oppHand+1 {
		t.Errorf("hands %d and %d, want %d and %d", me.Hand.Size(), opp.Hand.Size(), meHand+1, oppHand+1)
	}
	if loyaltyOf(g, teyo) != 4 {
		t.Errorf("loyalty %d, want 4", loyaltyOf(g, teyo))
	}
}

// --- Sarkhan the Masterless ----------------------------------------

// A creature attacking Sarkhan's controller takes 1 damage from each
// Dragon they control.
func TestGuffcSarkhanDragonsPingAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	guffcWalker(t, g, opp.ID, "Sarkhan the Masterless", "Legendary Planeswalker — Sarkhan", guffcSarkhan, 5)
	pushNamedPermanent(g, opp.ID, "Dragon One", "Creature — Dragon")
	pushNamedPermanent(g, opp.ID, "Dragon Two", "Creature — Dragon")
	pushNamedPermanent(g, other.ID, "Other Dragon", "Creature — Dragon")
	ogre := pushVanillaCreature(g, me.ID, "Ogre", 3, 3)

	declareAttack(t, g, opp.ID, ogre)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, ogre)
	if !ok {
		t.Fatal("the 3/3 died to two pings")
	}
	if c.DamageMarked != 2 {
		t.Errorf("damage %d, want 2 (one per Dragon its controller controls)", c.DamageMarked)
	}
}

// An attack on another player triggers nothing.
func TestGuffcSarkhanIgnoresAttacksOnOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	sarkhan := guffcWalker(t, g, opp.ID, "Sarkhan the Masterless", "Legendary Planeswalker — Sarkhan", guffcSarkhan, 5)
	pushNamedPermanent(g, opp.ID, "Dragon One", "Creature — Dragon")
	ogre := pushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	declareAttack(t, g, other.ID, ogre)
	if n := triggersOnStackFrom(g, sarkhan); n != 0 {
		t.Errorf("%d Sarkhan triggers for an attack on another player, want 0", n)
	}
}

// +1 animates the planeswalkers you control as it resolves; one that
// arrives later stays a planeswalker only.
func TestGuffcSarkhanPlusOneAnimatesYourPlaneswalkers(t *testing.T) {
	g, me, opp := spendTable(t)
	sarkhan := guffcWalker(t, g, me.ID, "Sarkhan the Masterless", "Legendary Planeswalker — Sarkhan", guffcSarkhan, 5)
	other := guffcWalker(t, g, me.ID, "Other Walker", "Legendary Planeswalker — Test", "", 3)
	theirs := guffcWalker(t, g, opp.ID, "Their Walker", "Legendary Planeswalker — Test", "", 3)
	b16Activate(t, g, me.ID, sarkhan, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	late := guffcWalker(t, g, me.ID, "Late Walker", "Legendary Planeswalker — Test", "", 3)

	for _, id := range []uuid.UUID{sarkhan, other} {
		c := effectiveCardForTest(t, g, id)
		if !c.IsCreature() || !c.IsPlaneswalker() || !c.HasSubtype("Dragon") || c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "flying") {
			t.Errorf("%s: creature %v walker %v Dragon %v %d/%d, want a 4/4 flying Dragon planeswalker", c.Name,
				c.IsCreature(), c.IsPlaneswalker(), c.HasSubtype("Dragon"), c.CurrentPower(), c.CurrentToughness())
		}
		if got := effectiveColors(t, g, id); len(got) != 1 || got[0] != "R" {
			t.Errorf("%s colours %v, want red", c.Name, got)
		}
	}
	for _, id := range []uuid.UUID{theirs, late} {
		if c := effectiveCardForTest(t, g, id); c.IsCreature() {
			t.Errorf("%s became a creature", c.Name)
		}
	}
}

func TestGuffcSarkhanMinusThreeMakesADragon(t *testing.T) {
	g, me, _ := spendTable(t)
	sarkhan := guffcWalker(t, g, me.ID, "Sarkhan the Masterless", "Legendary Planeswalker — Sarkhan", guffcSarkhan, 5)
	b16Activate(t, g, me.ID, sarkhan, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !onBattlefieldByName(g, "Dragon") || loyaltyOf(g, sarkhan) != 2 {
		t.Errorf("no Dragon token, or loyalty %d (want 2)", loyaltyOf(g, sarkhan))
	}
}

// --- Narset, Enlightened Master ------------------------------------

// The attack exiles the top four; a noncreature card among them can be
// cast with no mana, a creature card can't.
func TestGuffcNarsetExilesFourAndCastsNoncreatureFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	narset := pushCatalogPermanent(g, me.ID, "Narset, Enlightened Master", "Legendary Creature — Human Monk", guffcNarsetOracle, false)
	me.Library.Cards = nil
	// pushLibraryCardForTest pushes to the bottom: the first pushed is the top.
	instant := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID})
	creature := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Ponder", TypeLine: "Sorcery", ManaCost: "{U}", Owner: me.ID})
	bottom := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Fifth", TypeLine: "Instant", ManaCost: "{U}", Owner: me.ID})

	declareAttack(t, g, opp.ID, narset)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(bottom) || !g.Exile.Contains(instant) || !g.Exile.Contains(creature) {
		t.Fatalf("exile has %d cards; want exactly the top four", len(g.Exile.Cards))
	}
	if len(g.Exile.Cards) != 4 {
		t.Fatalf("exiled %d cards, want 4", len(g.Exile.Cards))
	}
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{FromZone: "exile", Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Errorf("casting the instant free with an empty pool: %v", err)
	}
	passPriorityAroundTable(t, g)
	// In the second main phase, where a creature's timing would allow
	// it, the creature card is still refused: the permission is for
	// noncreature spells.
	advanceTo(t, g, game.StepPostcombatMain)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a creature card was castable from Narset's exile")
	}
}

// --- Vronos, Masked Inquisitor -------------------------------------

// +1 phases the other targeted walker out at the next end step, not
// before; Vronos itself can't be targeted.
func TestGuffcVronosPlusOnePhasesOutAtTheEndStep(t *testing.T) {
	g, me, _ := spendTable(t)
	vronos := guffcWalker(t, g, me.ID, "Vronos, Masked Inquisitor", "Legendary Planeswalker — Vronos", guffcVronosOracle, 5)
	other := guffcWalker(t, g, me.ID, "Other Walker", "Legendary Planeswalker — Test", "", 3)
	if err := g.ActivateCatalogAbility(me.ID, vronos, 0, game.ActivateAbilityParams{Targets: guffcRefs(vronos)}); err == nil {
		t.Fatal("+1 targeted Vronos; it says other")
	}
	b16Activate(t, g, me.ID, vronos, 0, game.ActivateAbilityParams{Targets: guffcRefs(other)})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(other) {
		t.Fatal("the walker phased out before the end step")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) || !g.PhasedOut.Contains(other) {
		t.Error("the walker did not phase out at the end step")
	}
	if !g.Battlefield.Contains(vronos) {
		t.Error("Vronos phased out")
	}
}

// A walker that left and came back before the end step is a new object
// and stays.
func TestGuffcVronosPlusOneSkipsANewObject(t *testing.T) {
	g, me, _ := spendTable(t)
	vronos := guffcWalker(t, g, me.ID, "Vronos, Masked Inquisitor", "Legendary Planeswalker — Vronos", guffcVronosOracle, 5)
	other := guffcWalker(t, g, me.ID, "Other Walker", "Legendary Planeswalker — Test", "", 3)
	b16Activate(t, g, me.ID, vronos, 0, game.ActivateAbilityParams{Targets: guffcRefs(other)})
	passPriorityAroundTable(t, g)
	// Stand-in for a blink: the card is a new object (CR 400.7).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == other {
				g.Battlefield.Cards[i].ObjectEpoch++
			}
		}
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(other) {
		t.Error("a new object phased out")
	}
}

// −2 returns up to one nonland permanent per opponent; two from one
// opponent are refused.
func TestGuffcVronosMinusTwoBouncesOnePerOpponent(t *testing.T) {
	g, me, opp := spendTable(t)
	opp2 := g.Seats[2]
	vronos := guffcWalker(t, g, me.ID, "Vronos, Masked Inquisitor", "Legendary Planeswalker — Vronos", guffcVronosOracle, 5)
	a := pushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 2)
	c := pushVanillaCreature(g, opp2.ID, "C", 2, 2)
	land := pushNamedPermanent(g, opp2.ID, "Their Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, vronos, 1, game.ActivateAbilityParams{Targets: guffcRefs(a, b)}); err == nil {
		t.Fatal("two permanents of one opponent were accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, vronos, 1, game.ActivateAbilityParams{Targets: guffcRefs(land)}); err == nil {
		t.Fatal("a land was accepted")
	}
	b16Activate(t, g, me.ID, vronos, 1, game.ActivateAbilityParams{Targets: guffcRefs(a, c)})
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp2.Hand.Contains(c) || !g.Battlefield.Contains(b) {
		t.Error("the targets were not returned to their owners' hands")
	}
}

// −7: the artifact becomes a 9/9 Construct artifact creature with
// vigilance and indestructible, for good.
func TestGuffcVronosMinusSevenMakesANineNine(t *testing.T) {
	g, me, _ := spendTable(t)
	vronos := guffcWalker(t, g, me.ID, "Vronos, Masked Inquisitor", "Legendary Planeswalker — Vronos", guffcVronosOracle, 8)
	rock := pushNamedPermanent(g, me.ID, "Rock", "Artifact")
	b16Activate(t, g, me.ID, vronos, 2, game.ActivateAbilityParams{Targets: guffcRefs(rock)})
	passPriorityAroundTable(t, g)
	advanceToEndStepOf(t, g, 1)
	c := effectiveCardForTest(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Construct") || c.CurrentPower() != 9 || c.CurrentToughness() != 9 ||
		!game.HasKeyword(&c, "vigilance") || !game.HasKeyword(&c, "indestructible") {
		t.Errorf("the artifact is creature %v %d/%d keywords %v, want a 9/9 Construct with vigilance and indestructible next turn",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), c.Keywords)
	}
}

// --- Final Showdown ------------------------------------------------

// All three bullets: abilities go, the chosen creature survives the
// wrath, every other creature dies.
func TestGuffcFinalShowdownAllThreeBullets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := pushVanillaCreature(g, me.ID, "Keeper", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "Other Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castModal(t, g, "Final Showdown", "Instant", guffcShowdown, []int{0, 1, 2}, nil)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no choice of a creature you control: %+v", g.PendingChoices)
	}
	if hasID(pick.ChooseCards, theirs) {
		t.Error("an opponent's creature was offered")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("the wrath ran before the choice was answered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{keep}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(keep) {
		t.Error("the chosen creature died")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an unchosen creature survived the wrath")
	}
}

// The first bullet alone strips a creature's keywords until end of
// turn.
func TestGuffcFinalShowdownStripsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	flyer := pushPermanent(g, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	castModal(t, g, "Final Showdown", "Instant", guffcShowdown, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if c := effectiveCardForTest(t, g, flyer); game.HasKeyword(&c, "flying") {
		t.Error("the creature kept flying")
	}
}

// --- Guff Rewrites History -----------------------------------------

func guffcLib(p *game.Player, cards ...game.Card) {
	p.Library.Cards = nil
	for _, c := range cards {
		c.Owner, c.Controller = p.ID, p.ID
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		p.Library.PushTop(c)
	}
}

// One target per player; each goes into its owner's library, each of
// those players walks to a nonland card and is offered it free.
func TestGuffcRewritesHistoryShufflesAndOffersEachPlayerACast(t *testing.T) {
	g := newCatalogGame(t)
	var targets []uuid.UUID
	hits := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		targets = append(targets, pushVanillaCreature(g, p.ID, "Bear of "+p.Name, 2, 2))
		hit := uuid.New()
		hits[p.ID] = hit
		// The top of the library is the last pushed: a land, then the hit.
		guffcLib(p, game.Card{InstanceID: hit, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"},
			game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	}
	me := g.Seats[0]
	enchant := pushNamedPermanent(g, me.ID, "My Aura", "Enchantment")
	bad := append([]uuid.UUID{enchant}, targets[1:]...)
	if _, err := castModalErr(t, g, "Guff Rewrites History", "Instant", guffcRewrites, nil, guffcRefs(bad...)); err == nil {
		t.Fatal("an enchantment was accepted")
	}
	if _, err := castModalErr(t, g, "Guff Rewrites History", "Instant", guffcRewrites, nil, guffcRefs(targets[:3]...)); err == nil {
		t.Fatal("a cast leaving one opponent untargeted was accepted")
	}
	castCatalogSpell(t, g, "Guff Rewrites History", "Instant", guffcRewrites, guffcRefs(targets...))
	passPriorityAroundTable(t, g)

	for i, p := range g.Seats {
		if g.Battlefield.Contains(targets[i]) {
			t.Errorf("%s's target is still on the battlefield", p.Name)
		}
	}
	// The shuffled-in creature is a nonland card too, so the walk may
	// stop on it or on the Divination.
	prompts := 0
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceMayCast {
			continue
		}
		prompts++
		card, ok := cardInZone(g.Exile, c.MayCastCard)
		if !ok || card.Owner != c.Chooser || card.IsLand() {
			t.Errorf("%s was offered %s (in exile %v), want a nonland card of their own", c.Chooser, card.Name, ok)
		}
		if card.InstanceID != hits[c.Chooser] && card.Name[:4] != "Bear" {
			t.Errorf("%s was offered %q", c.Chooser, card.Name)
		}
	}
	if prompts != len(g.Seats) {
		t.Errorf("%d may-cast offers, want one per player (%d)", prompts, len(g.Seats))
	}
}

// #4205: a player who declines the free cast keeps the card in exile;
// the cards the walk passed over go to the bottom of their library.
func TestGuffcRewritesHistoryDeclinedCardStaysInExile(t *testing.T) {
	g := newCatalogGame(t)
	var targets []uuid.UUID
	hits := map[uuid.UUID]uuid.UUID{}
	lands := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		targets = append(targets, pushVanillaCreature(g, p.ID, "Bear of "+p.Name, 2, 2))
		hit, land := uuid.New(), uuid.New()
		hits[p.ID], lands[p.ID] = hit, land
		// The top is the last pushed: a land, then the Divination. The
		// shuffled-in bear may land above either, so the offered card is
		// read from the prompt, not assumed.
		guffcLib(p, game.Card{InstanceID: hit, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"},
			game.Card{InstanceID: lands[p.ID], Name: "Island", TypeLine: "Basic Land — Island"})
	}
	castCatalogSpell(t, g, "Guff Rewrites History", "Instant", guffcRewrites, guffcRefs(targets...))
	passPriorityAroundTable(t, g)

	var offered []uuid.UUID
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			offered = append(offered, c.MayCastCard)
		}
	}
	if len(offered) != len(g.Seats) {
		t.Fatalf("%d offers, want %d", len(offered), len(g.Seats))
	}
	for len(pendingMayCasts(g)) > 0 {
		c := pendingMayCasts(g)[0]
		if err := g.ResolveMayCast(c.ID, c.Chooser, false); err != nil {
			t.Fatalf("decline: %v", err)
		}
	}
	for _, id := range offered {
		if !g.Exile.Contains(id) {
			t.Errorf("declined card %s left exile", id)
		}
		for _, p := range g.Seats {
			if p.Library.Contains(id) {
				t.Errorf("declined card %s went to %s's library", id, p.Name)
			}
		}
	}
	if len(g.Exile.Cards) != len(offered) {
		t.Errorf("exile holds %d cards, want only the %d declined ones", len(g.Exile.Cards), len(offered))
	}
	for _, p := range g.Seats {
		if g.Exile.Contains(lands[p.ID]) {
			t.Errorf("%s's passed-over land stayed in exile", p.Name)
		}
	}
}

func pendingMayCasts(g *game.Game) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			out = append(out, c)
		}
	}
	return out
}
