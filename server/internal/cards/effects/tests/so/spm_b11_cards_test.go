package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b11_cards_test.go — Marvel's Spider-Man, batch 11 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spmB11WebShootersOracle   = "ec528198-a38a-4ef7-9804-78a25c94ecc1"
	spmB11WebWarriorsOracle   = "3b23e4f8-882d-4cce-9b61-171ae2cc3a18"
	spmB11WraithOracle        = "38684c31-e8c6-4825-8740-1a3d34b58b39"
	spmB11AmateurHeroOracle   = "fc8ee6e1-a2e1-4033-a164-7968fc79dd27"
	spmB11DocOckOracle        = "3e0738af-00ec-4707-a56d-85b5a5ce4529"
	spmB11DoubleTroubleOracle = "2ef063d2-8702-43a3-a7d0-fc255f8512d3"
	spmB11FutureFlightOracle  = "bfd5b48a-698e-4ade-afb0-954ba288f9a7"
	spmB11GhostSpiderOracle   = "1d95b5c0-d193-4a03-9954-23861ac1c762"
	spmB11TentaclesOracle     = "dac76b28-23e8-4138-8442-51579a590f16"
)

func TestSpmB11EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB11WebShootersOracle:   "Web-Shooters",
		spmB11WebWarriorsOracle:   "Web-Warriors",
		spmB11WraithOracle:        "Wraith, Vicious Vigilante",
		spmB11AmateurHeroOracle:   "Amateur Hero",
		spmB11DocOckOracle:        "Doc Ock, Evil Inventor",
		spmB11DoubleTroubleOracle: "Double Trouble",
		spmB11FutureFlightOracle:  "Future Flight",
		spmB11GhostSpiderOracle:   "Ghost-Spider, Gwen Stacy",
		spmB11TentaclesOracle:     "Grasping Tentacles",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func spmB11Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

// spmB11Power is the creature's current power, counters and effects included.
func spmB11Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	g.RecomputeLayersIfStaleLocked()
	return spmB11Live(t, g, id).CurrentPower()
}

// --- Wraith, Vicious Vigilante ---

func TestSpmB11WraithHasDoubleStrikeAndCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	w := effectstest.B12Push(g, me.ID, "Wraith, Vicious Vigilante", "Legendary Creature — Human Detective Hero", spmB11WraithOracle, 1, 1)
	effectstest.AssertKeywords(t, g, w, "double strike")
	if effectstest.RestrictionsOf(t, g, w)&game.CantBeBlocked == 0 {
		t.Error("Wraith can be blocked")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.RestrictionsOf(t, g, bear)&game.CantBeBlocked != 0 {
		t.Error("another creature can't be blocked")
	}
}

// --- Amateur Hero ---

func TestSpmB11AmateurHeroGainsTwoOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Amateur Hero", "Creature — Human Hero", spmB11AmateurHeroOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before+2 {
		t.Errorf("life %d -> %d, want +2", before, got)
	}
}

// --- Web-Warriors ---

func TestSpmB11WebWarriorsCountersOtherCreaturesYouControlOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Ally A", "Creature — Human", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Ally B", "Creature — Human", 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Web-Warriors", "Creature — Spider Hero", spmB11WebWarriorsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := spmB11Power(t, g, a); got != 3 {
		t.Errorf("Ally A power %d, want 3", got)
	}
	if got := spmB11Power(t, g, b); got != 2 {
		t.Errorf("Ally B power %d, want 2", got)
	}
	if got := spmB11Power(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature got a counter: power %d", got)
	}
	if n := spmB11Live(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Errorf("Web-Warriors put a counter on itself: %d", n)
	}
}

// --- Double Trouble ---

func TestSpmB11DoubleTroubleDoublesEachOfYourCreaturesPowerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Ally A", "Creature — Human", 3, 2)
	b := effectstest.B12Creature(g, me.ID, "Ally B", "Creature — Human", 1, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.AddCounterByForEffect(me.ID, b, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	effectstest.CastCatalogSpell(t, g, "Double Trouble", "Instant", spmB11DoubleTroubleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := spmB11Power(t, g, a); got != 6 {
		t.Errorf("Ally A power %d, want 6", got)
	}
	// Ally B is 1/4 with a +1/+1 counter: 2 power doubles to 4, toughness stays 5.
	if got := spmB11Power(t, g, b); got != 4 {
		t.Errorf("Ally B power %d, want 4", got)
	}
	if got := spmB11Live(t, g, b).CurrentToughness(); got != 5 {
		t.Errorf("Ally B toughness %d, want 5 (toughness is not doubled)", got)
	}
	if got := spmB11Power(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature was doubled: power %d", got)
	}
}

// --- Future Flight ---

func TestSpmB11FutureFlightDrawsTwoAndGrantsPlusTwoAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Future Flight", "Enchantment — Aura", spmB11FutureFlightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	// The Aura left the hand; two cards were drawn.
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2 (the Aura itself was already counted out)", hand, got)
	}
	if got := spmB11Power(t, g, bear); got != 4 {
		t.Errorf("enchanted creature power %d, want 4", got)
	}
	effectstest.AssertKeywords(t, g, bear, "flying")
}

// --- Ghost-Spider, Gwen Stacy ---

func TestSpmB11GhostSpiderDealsDamageEqualToAttackingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	gs := effectstest.B12Push(g, me.ID, "Ghost-Spider, Gwen Stacy", "Legendary Creature — Spider Human Hero", spmB11GhostSpiderOracle, 4, 4)
	effectstest.AssertKeywords(t, g, gs, "menace")
	a := effectstest.B12Creature(g, me.ID, "Ally A", "Creature — Human", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Ally B", "Creature — Human", 1, 1)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Human", 1, 1)
	_ = idle
	oppLife, thirdLife := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, third.ID)
	effectstest.DeclareAttack(t, g, opp.ID, gs, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	// Three attackers, so three damage to the defending player before combat damage.
	if got := effectstest.LifeOf(g, opp.ID); got > oppLife-3 {
		t.Errorf("defending player life %d -> %d, want at least 3 lost", oppLife, got)
	}
	if got := effectstest.LifeOf(g, third.ID); got != thirdLife {
		t.Errorf("a non-defending player lost life: %d -> %d", thirdLife, got)
	}
}

func TestSpmB11GhostSpiderAttackingAloneDealsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gs := effectstest.B12Push(g, me.ID, "Ghost-Spider, Gwen Stacy", "Legendary Creature — Spider Human Hero", spmB11GhostSpiderOracle, 4, 4)
	oppLife := effectstest.LifeOf(g, opp.ID)
	effectstest.DeclareAttack(t, g, opp.ID, gs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("defending player life %d -> %d, want exactly 1 lost before combat damage", oppLife, got)
	}
}

// --- Doc Ock, Evil Inventor ---

func TestSpmB11DocOckAnimatesANoncreatureArtifactForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Doc Ock, Evil Inventor", "Legendary Creature — Human Scientist Villain", spmB11DocOckOracle, 8, 8)
	rock := effectstest.B12Permanent(g, me.ID, "Trinket", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.MscM2PickTargets(t, g, rock)
	c := spmB11Live(t, g, rock)
	if !c.IsCreature() || !c.HasCardType("artifact") {
		t.Fatalf("Trinket types %q, want an artifact creature", c.TypeLine)
	}
	if !c.HasSubtype("Robot") || !c.HasSubtype("Villain") {
		t.Errorf("Trinket subtypes %v, want Robot Villain", c.Effective().Subtypes)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, rock); p != 8 || tough != 8 {
		t.Errorf("Trinket is %d/%d, want 8/8", p, tough)
	}
	// No duration: it is still a creature in the next phase.
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if !spmB11Live(t, g, rock).IsCreature() {
		t.Error("the animation ended with the phase")
	}
}

func TestSpmB11DocOckCantTargetACreatureArtifactOrAnotherPlayersArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Doc Ock, Evil Inventor", "Legendary Creature — Human Scientist Villain", spmB11DocOckOracle, 8, 8)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Trinket", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Trinket", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, id := range []uuid.UUID{golem, theirs} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Errorf("target %s was accepted", id)
		}
	}
	effectstest.MscM2PickTargets(t, g, mine)
	if !spmB11Live(t, g, mine).IsCreature() {
		t.Error("the legal target was not animated")
	}
	if spmB11Live(t, g, theirs).IsCreature() {
		t.Error("an opponent's artifact was animated")
	}
}

// --- Web-Shooters ---

func TestSpmB11WebShootersGrowsGrantsReachAndTapsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gear := effectstest.B12Push(g, me.ID, "Web-Shooters", effectstest.EquipTypeLine, spmB11WebShootersOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	effectstest.EquipTo(t, g, me.ID, gear, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3", p, tough)
	}
	effectstest.AssertKeywords(t, g, bear, "reach")
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.MscM2PickTargets(t, g, blocker)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, blocker); !tapped {
		t.Error("the opponent's creature was not tapped")
	}
}

func TestSpmB11WebShootersTriggerOnlyOnTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gear := effectstest.B12Push(g, me.ID, "Web-Shooters", effectstest.EquipTypeLine, spmB11WebShootersOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	effectstest.EquipTo(t, g, me.ID, gear, bear)
	effectstest.DeclareAttack(t, g, opp.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, blocker); tapped {
		t.Error("an unequipped attacker tapped a creature")
	}
}

// --- Grasping Tentacles ---

func spmB11Tentacles(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// Eight cards will be milled from the top; the artifact sits among them,
	// and a second artifact was already in the graveyard.
	for i := 0; i < 10; i++ {
		effectstest.SeedLibraryTop(opp, "Filler", "Sorcery", "")
	}
	milledArtifact := uuid.New()
	g.WithWriteLock(func() {
		opp.Library.PushTop(game.Card{InstanceID: milledArtifact, Name: "Milled Gizmo", TypeLine: "Artifact",
			Owner: opp.ID, Controller: opp.ID})
	})
	oldArtifact := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: oldArtifact, Name: "Old Gizmo", TypeLine: "Artifact",
			Owner: opp.ID, Controller: opp.ID})
		opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	})
	lib := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Grasping Tentacles", "Sorcery", spmB11TentaclesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib-8 {
		t.Errorf("opponent library %d -> %d, want exactly 8 milled", lib, got)
	}
	return g, me, opp, milledArtifact, oldArtifact
}

func TestSpmB11GraspingTentaclesMillsEightThenStealsAnArtifact(t *testing.T) {
	g, me, _, milled, _ := spmB11Tentacles(t)
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, milled)
	if !ok {
		t.Fatal("the artifact did not enter the battlefield")
	}
	if c.Controller != me.ID {
		t.Errorf("artifact controlled by %s, want the caster", c.Controller)
	}
}

func TestSpmB11GraspingTentaclesOffersOnlyArtifactsAndMayDecline(t *testing.T) {
	g, me, opp, milled, old := spmB11Tentacles(t)
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no artifact prompt")
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range ch.ChooseCards {
		seen[id] = true
	}
	if len(seen) != 2 || !seen[milled] || !seen[old] {
		t.Errorf("options %v, want exactly the milled and the old artifact", ch.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, milled); ok {
		t.Error("declining still put an artifact onto the battlefield")
	}
	if !opp.Graveyard.Contains(milled) || !opp.Graveyard.Contains(old) {
		t.Error("a declined pick left the graveyard")
	}
}
