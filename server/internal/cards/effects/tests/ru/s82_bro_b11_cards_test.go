package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 BRO batch 11: the three Assembly-Workers, Mishra's Foundry,
// Levitating Statue, Alloy Animist, Mightstone's Animation, Tawnos's
// Tinkering, Argivian Avenger and Surge Engine.

const (
	s82b11MineWorker     = "4de71245-9d9b-44ca-9c13-37dc3dfbbcd5"
	s82b11PowerPlant     = "054c6250-ac96-42f6-931d-925761ab7764"
	s82b11TowerWorker    = "2f995678-c315-4afe-845b-8c8ae863de6e"
	s82b11Foundry        = "b43e9772-6ad4-49c7-9557-b18ee1e4587d"
	s82b11Statue         = "259ca16d-7753-402b-bcd0-2425500dee99"
	s82b11Animist        = "4acacdd2-a736-43a2-ad67-891b9bd124f8"
	s82b11Animation      = "de68a69a-0c4e-43b9-821d-2fe2269b5e48"
	s82b11Tinkering      = "bd101a9b-4c1e-44c8-b9ef-cd79c4d28f36"
	s82b11Avenger        = "d215cd4a-412c-4632-ab2d-de51103e6a7d"
	s82b11SurgeEngine    = "864d9db7-34ca-4440-8d6d-096f9eddedd4"
	s82b11WorkerTypeLine = "Artifact Creature — Assembly-Worker"
)

func TestS82BroB11CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{s82b11MineWorker, s82b11PowerPlant, s82b11TowerWorker, s82b11Foundry,
		s82b11Statue, s82b11Animist, s82b11Animation, s82b11Tinkering, s82b11Avenger, s82b11SurgeEngine} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82BroB11MineWorkerGainsOneLifeOrThreeWithTheTrio(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	worker := effectstest.PushCatalogPermanent(g, me.ID, "Mine Worker", s82b11WorkerTypeLine, s82b11MineWorker, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Power Plant Worker", s82b11WorkerTypeLine, s82b11PowerPlant, false)
	life := me.Life
	if err := effectstest.S82b11Activate(t, g, me, worker, 0); err != nil {
		t.Fatal(err)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d with only Power Plant Worker, want +1", life, me.Life)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	worker2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Mine Worker", s82b11WorkerTypeLine, s82b11MineWorker, false)
	effectstest.PushCatalogPermanent(g2, me2.ID, "Power Plant Worker", s82b11WorkerTypeLine, s82b11PowerPlant, false)
	effectstest.PushCatalogPermanent(g2, me2.ID, "Tower Worker", s82b11WorkerTypeLine, s82b11TowerWorker, false)
	life = me2.Life
	if err := effectstest.S82b11Activate(t, g2, me2, worker2, 0); err != nil {
		t.Fatal(err)
	}
	if me2.Life != life+3 {
		t.Errorf("life %d → %d with the trio, want +3", life, me2.Life)
	}
}

func TestS82BroB11TheTrioMustBeYoursAndBeCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worker := effectstest.PushCatalogPermanent(g, me.ID, "Mine Worker", s82b11WorkerTypeLine, s82b11MineWorker, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Power Plant Worker", s82b11WorkerTypeLine, s82b11PowerPlant, false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Tower Worker", s82b11WorkerTypeLine, s82b11TowerWorker, false)
	life := me.Life
	if err := effectstest.S82b11Activate(t, g, me, worker, 0); err != nil {
		t.Fatal(err)
	}
	if me.Life != life+1 {
		t.Errorf("an opponent's Tower Worker counted: life %d → %d, want +1", life, me.Life)
	}
}

func TestS82BroB11PowerPlantWorkerPumpsOrGrowsOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	worker := effectstest.PushCatalogPermanent(g, me.ID, "Power Plant Worker", s82b11WorkerTypeLine, s82b11PowerPlant, false)
	before := effectstest.EffectivePower(t, g, worker)
	effectstest.FillPool(me, 6)
	if err := effectstest.S82b11Activate(t, g, me, worker, 0); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.EffectivePower(t, g, worker); got != before+2 {
		t.Errorf("power %d → %d, want +2 until end of turn", before, got)
	}
	if got := effectstest.PlusOnes(g, worker); got != 0 {
		t.Errorf("%d counters without the trio, want 0", got)
	}
	if err := effectstest.S82b11Activate(t, g, me, worker, 0); err == nil {
		t.Error("activated twice in one turn")
	}
}

func TestS82BroB11PowerPlantWorkerPutsTwoCountersWithTheTrio(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	worker := effectstest.PushCatalogPermanent(g, me.ID, "Power Plant Worker", s82b11WorkerTypeLine, s82b11PowerPlant, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Mine Worker", s82b11WorkerTypeLine, s82b11MineWorker, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Tower Worker", s82b11WorkerTypeLine, s82b11TowerWorker, false)
	before := effectstest.EffectivePowerWithCounters(t, g, worker)
	effectstest.FillPool(me, 3)
	if err := effectstest.S82b11Activate(t, g, me, worker, 0); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.PlusOnes(g, worker); got != 2 {
		t.Errorf("%d +1/+1 counters with the trio, want 2", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, worker); got != before+2 {
		t.Errorf("power %d → %d, want +2 from the counters only", before, got)
	}
	// No pump on top: the counters replace it, so end of turn leaves them.
}

func TestS82BroB11TowerWorkerAddsOneOrThreeColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tower := effectstest.PushCatalogPermanent(g, me.ID, "Tower Worker", s82b11WorkerTypeLine, s82b11TowerWorker, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Mine Worker", s82b11WorkerTypeLine, s82b11MineWorker, false)
	if err := g.ActivateManaAbility(me.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool %d without the trio, want 1", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	tower2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Tower Worker", s82b11WorkerTypeLine, s82b11TowerWorker, false)
	effectstest.PushCatalogPermanent(g2, me2.ID, "Mine Worker", s82b11WorkerTypeLine, s82b11MineWorker, false)
	effectstest.PushCatalogPermanent(g2, me2.ID, "Power Plant Worker", s82b11WorkerTypeLine, s82b11PowerPlant, false)
	if err := g2.ActivateManaAbility(me2.ID, tower2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := me2.ManaPool.Total(); got != 3 {
		t.Errorf("pool %d with the trio, want 3", got)
	}
}

func TestS82BroB11MishrasFoundryAnimatesAndPumpsAnAttackingAssemblyWorker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foundry := effectstest.PushCatalogPermanent(g, me.ID, "Mishra's Foundry", "Land", s82b11Foundry, false)
	worker := effectstest.PushCatalogPermanent(g, me.ID, "Mine Worker", s82b11WorkerTypeLine, s82b11MineWorker, false)
	idle := effectstest.PushCatalogPermanent(g, me.ID, "Tower Worker", s82b11WorkerTypeLine, s82b11TowerWorker, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.FillPool(me, 2)
	if err := effectstest.S82b11Activate(t, g, me, foundry, 0); err != nil {
		t.Fatal(err)
	}
	c := effectstest.EffectiveCardForTest(t, g, foundry)
	if !c.IsCreature() || !c.IsLand() || !c.IsArtifact() || !c.HasSubtype("Assembly-Worker") {
		t.Errorf("animated Foundry: creature %v land %v artifact %v Assembly-Worker %v",
			c.IsCreature(), c.IsLand(), c.IsArtifact(), c.HasSubtype("Assembly-Worker"))
	}
	if p, tough := effectstest.EffectivePower(t, g, foundry), effectstest.EffectiveToughness(t, g, foundry); p != 2 || tough != 2 {
		t.Errorf("Foundry is %d/%d, want 2/2", p, tough)
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{worker, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)

	effectstest.FillPool(me, 1)
	if err := effectstest.S82b11Activate(t, g, me, foundry, 1, idle); err == nil {
		t.Error("pumped an Assembly-Worker that is not attacking")
	}
	if err := effectstest.S82b11Activate(t, g, me, foundry, 1, bear); err == nil {
		t.Error("pumped an attacking creature that is not an Assembly-Worker")
	}
	before := effectstest.EffectivePower(t, g, worker)
	if err := effectstest.S82b11Activate(t, g, me, foundry, 1, worker); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.EffectivePower(t, g, worker); got != before+2 {
		t.Errorf("attacking Assembly-Worker power %d → %d, want +2", before, got)
	}
}

func TestS82BroB11LevitatingStatueGrowsOnNoncreatureSpellsAndAnimates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	statue := effectstest.PushCatalogPermanent(g, me.ID, "Levitating Statue", "Artifact", s82b11Statue, false)

	effectstest.CastCatalogSpell(t, g, "Relic", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, statue); got != 1 {
		t.Fatalf("%d counters after a noncreature spell, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, statue); got != 1 {
		t.Errorf("a creature spell grew the Statue: %d counters", got)
	}

	if c := effectstest.EffectiveCardForTest(t, g, statue); c.IsCreature() {
		t.Fatal("the Statue is a creature before it is animated")
	}
	effectstest.FillPool(me, 2)
	if err := effectstest.S82b11Activate(t, g, me, statue, 0); err != nil {
		t.Fatal(err)
	}
	c := effectstest.EffectiveCardForTest(t, g, statue)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Construct") {
		t.Errorf("animated Statue: creature %v artifact %v Construct %v", c.IsCreature(), c.IsArtifact(), c.HasSubtype("Construct"))
	}
	if p, tough := effectstest.EffectivePowerWithCounters(t, g, statue), s82b11ToughnessWithCounters(t, g, statue); p != 2 || tough != 2 {
		t.Errorf("Statue is %d/%d, want a 1/1 with one counter (2/2)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, statue, "flying") {
		t.Error("no flying")
	}
}

func TestS82BroB11AlloyAnimistMakesANoncreatureArtifactA44(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	animist := effectstest.PushCatalogPermanent(g, me.ID, "Alloy Animist", "Creature — Human Druid", s82b11Animist, false)
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Relic", "Artifact", "", false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Plain Land", "Land", "", false)

	fillColored := func() { effectstest.FillPoolColored(me, "G", 3) }
	fillColored()
	if err := effectstest.S82b11Activate(t, g, me, animist, 0, theirs); err == nil {
		t.Error("targeted an opponent's artifact")
	}
	if err := effectstest.S82b11Activate(t, g, me, animist, 0, land); err == nil {
		t.Error("targeted a nonartifact land")
	}
	if err := effectstest.S82b11Activate(t, g, me, animist, 0, animist); err == nil {
		t.Error("targeted a creature")
	}
	if err := effectstest.S82b11Activate(t, g, me, animist, 0, relic); err != nil {
		t.Fatal(err)
	}
	c := effectstest.EffectiveCardForTest(t, g, relic)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("Relic: creature %v artifact %v", c.IsCreature(), c.IsArtifact())
	}
	if p, tough := effectstest.EffectivePower(t, g, relic), effectstest.EffectiveToughness(t, g, relic); p != 4 || tough != 4 {
		t.Errorf("Relic is %d/%d, want 4/4", p, tough)
	}
}

func TestS82BroB11AlloyAnimistOnAnArtifactLandMakesItNoLongerALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	animist := effectstest.PushCatalogPermanent(g, me.ID, "Alloy Animist", "Creature — Human Druid", s82b11Animist, false)
	citadel := effectstest.PushCatalogPermanent(g, me.ID, "Citadel", "Artifact Land", "", false)
	effectstest.FillPoolColored(me, "G", 3)
	if err := effectstest.S82b11Activate(t, g, me, animist, 0, citadel); err != nil {
		t.Fatal(err)
	}
	c := effectstest.EffectiveCardForTest(t, g, citadel)
	if !c.IsCreature() || !c.IsArtifact() || c.IsLand() {
		t.Errorf("artifact land: creature %v artifact %v land %v, want an artifact creature that is no longer a land",
			c.IsCreature(), c.IsArtifact(), c.IsLand())
	}
}

func TestS82BroB11MightstonesAnimationDrawsAndMakesTheArtifactA44Creature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	creature := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand := effectstest.HandSizeOf(g, me)
	if err := effectstest.CastCatalogSpellErr(t, g, "Mightstone's Animation", "Enchantment — Aura", s82b11Animation, effectstest.CardTarget(creature)); err == nil {
		t.Error("enchanted a nonartifact creature")
	}
	effectstest.CastCatalogSpell(t, g, "Mightstone's Animation", "Enchantment — Aura", s82b11Animation, effectstest.CardTarget(relic))
	effectstest.PassPriorityAroundTable(t, g)
	// The Aura left the hand (-1 for a second push, +1 drawn).
	if got := effectstest.HandSizeOf(g, me); got < hand {
		t.Errorf("hand %d → %d: no card drawn", hand, got)
	}
	c := effectstest.EffectiveCardForTest(t, g, relic)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("Relic: creature %v artifact %v", c.IsCreature(), c.IsArtifact())
	}
	if p, tough := effectstest.EffectivePower(t, g, relic), effectstest.EffectiveToughness(t, g, relic); p != 4 || tough != 4 {
		t.Errorf("Relic is %d/%d, want 4/4", p, tough)
	}
}

func TestS82BroB11TawnossTinkeringAnimatesANoncreatureAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Plain Land", "Land", "", false)
	effectstest.TapForTest(t, g, land)
	effectstest.CastCatalogSpell(t, g, "Tawnos's Tinkering", "Instant", s82b11Tinkering, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, land); got != 2 {
		t.Fatalf("%d counters, want 2", got)
	}
	c := effectstest.EffectiveCardForTest(t, g, land)
	if !c.IsCreature() || !c.IsLand() {
		t.Errorf("land: creature %v land %v, want a creature that is still a land", c.IsCreature(), c.IsLand())
	}
	if p, tough := effectstest.EffectivePowerWithCounters(t, g, land), s82b11ToughnessWithCounters(t, g, land); p != 2 || tough != 2 {
		t.Errorf("land is %d/%d, want 0/0 with two counters (2/2)", p, tough)
	}
	if b, ok := effectstest.BattlefieldCard(g, land); !ok || b.Tapped {
		t.Error("the land is still tapped")
	}
}

func TestS82BroB11TawnossTinkeringOnACreatureOnlyGrowsAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	effectstest.CastCatalogSpell(t, g, "Tawnos's Tinkering", "Instant", s82b11Tinkering, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, bear); p != 4 {
		t.Errorf("Bear power %d, want 4 (no 0/0 base applied to a creature)", p)
	}
	if b, ok := effectstest.BattlefieldCard(g, bear); !ok || b.Tapped {
		t.Error("the Bear is still tapped")
	}
}

func TestS82BroB11TawnossTinkeringRefusesAnOpponentsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Relic", "Artifact", "", false)
	if err := effectstest.CastCatalogSpellErr(t, g, "Tawnos's Tinkering", "Instant", s82b11Tinkering, effectstest.CardTarget(theirs)); err == nil {
		t.Error("targeted an opponent's artifact")
	}
}

func TestS82BroB11ArgivianAvengerShrinksAndGainsTheChosenKeyword(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	avenger := effectstest.PushCatalogPermanent(g, me.ID, "Argivian Avenger", "Artifact Creature — Shapeshifter", s82b11Avenger, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == avenger {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 5, 5
			}
		}
	})
	before := effectstest.EffectivePower(t, g, avenger)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, avenger, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 2) // deathtouch
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, avenger); got != before-1 {
		t.Errorf("power %d → %d, want -1", before, got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, avenger, "deathtouch") {
		t.Error("no deathtouch")
	}
	if effectstest.HasEffectiveKeyword(t, g, avenger, "flying") || effectstest.HasEffectiveKeyword(t, g, avenger, "haste") {
		t.Error("gained a keyword that was not chosen")
	}
}

func TestS82BroB11SurgeEngineBecomesUnblockableBlueAndDrawsThreeOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	engine := effectstest.PushCatalogPermanent(g, me.ID, "Surge Engine", "Artifact Creature — Construct", s82b11SurgeEngine, false)
	if !effectstest.HasEffectiveKeyword(t, g, engine, "defender") {
		// the layered view reads the printed keyword off the catalog row
		t.Fatal("no defender before the first ability")
	}

	effectstest.FillPoolColored(me, "U", 12)
	if err := effectstest.S82b11Activate(t, g, me, engine, 1); err == nil {
		t.Error("{2}{U} activated while the Engine has defender")
	}
	if err := effectstest.S82b11Activate(t, g, me, engine, 2); err == nil {
		t.Error("the draw activated while the Engine is not blue")
	}

	if err := effectstest.S82b11Activate(t, g, me, engine, 0); err != nil {
		t.Fatal(err)
	}
	if effectstest.HasEffectiveKeyword(t, g, engine, "defender") {
		t.Error("still has defender")
	}
	if effectstest.RestrictionsOf(t, g, engine)&game.CantBeBlocked == 0 {
		t.Error("can be blocked")
	}

	if err := effectstest.S82b11Activate(t, g, me, engine, 2); err == nil {
		t.Error("the draw activated while the Engine is not blue")
	}
	if err := effectstest.S82b11Activate(t, g, me, engine, 1); err != nil {
		t.Fatal(err)
	}
	if c := effectstest.EffectiveCardForTest(t, g, engine); !c.HasColor("U") {
		t.Error("not blue")
	}
	if p, tough := effectstest.EffectivePower(t, g, engine), effectstest.EffectiveToughness(t, g, engine); p != 5 || tough != 4 {
		t.Errorf("Engine is %d/%d, want 5/4", p, tough)
	}

	hand := effectstest.HandSizeOf(g, me)
	if err := effectstest.S82b11Activate(t, g, me, engine, 2); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.HandSizeOf(g, me); got != hand+3 {
		t.Errorf("hand %d → %d, want +3", hand, got)
	}
	if err := effectstest.S82b11Activate(t, g, me, engine, 2); err == nil {
		t.Error("the draw activated twice")
	}
}

// s82b11ToughnessWithCounters is the counter-aware toughness.
func s82b11ToughnessWithCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n, found := 0, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n, found = c.CurrentToughness(), true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return n
}
