package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b228_test.go — Ixalan, batch 228 (#5532): each test plays the real
// catalog entry through the real activation, trigger and combat paths.

const (
	oldB228HuatliKnight = "04aa210a-235f-4e07-87d1-0d28cdf6888b"
	oldB228Vraska       = "3c037103-5af0-40a8-bbbe-7f3f5958b6e9"
	oldB228HuatliPoet   = "3ec889bc-f4d7-492a-91c3-d3691eb90eac"
	oldB228JaceCastaway = "d2341a16-ffde-42fa-9746-0d8888e812d4"
	oldB228JaceMind     = "fe0fcb4e-92fd-4e64-99d1-043092ea21e3"
	oldB228Chupacabra   = "178e73f9-fc6e-4425-870e-45f96644010a"
	oldB228Tracker      = "26184baf-5d48-4919-8df2-c98ac460b78f"
	oldB228Vanguard     = "5007425e-a626-46f7-ab7d-777af97ee724"
	oldB228Caravel      = "7986fcb6-bafb-4bbc-9f2f-fd9e2a5fcad1"
	oldB228Walker       = "d80562c2-cf90-43a4-9721-28da40fdfeed"
)

func TestOldB228CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB228HuatliKnight: "Huatli, Dinosaur Knight",
		oldB228Vraska:       "Vraska, Relic Seeker",
		oldB228HuatliPoet:   "Huatli, Warrior Poet",
		oldB228JaceCastaway: "Jace, Cunning Castaway",
		oldB228JaceMind:     "Jace, Ingenious Mind-Mage",
		oldB228Chupacabra:   "Lurking Chupacabra",
		oldB228Tracker:      "Deadeye Tracker",
		oldB228Vanguard:     "Emperor's Vanguard",
		oldB228Caravel:      "Shadowed Caravel",
		oldB228Walker:       "Wildgrowth Walker",
	} {
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

func oldB228Card(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// oldB228Explore has `explorer` explore under `controller` with a land on top
// of the library, so the land goes to hand and no question is asked.
func oldB228Explore(g *game.Game, controller *game.Player, explorer uuid.UUID) {
	effectstest.SeedLibraryTop(controller, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		_ = g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: explorer}, controller.ID)
	})
}

func oldB228Counters(g *game.Game, id uuid.UUID) int {
	return effectstest.CounterCount(g, id, game.CounterPlusOne)
}

// --- Wildgrowth Walker ---

func TestOldB228WildgrowthWalkerGrowsAndGainsLifeOnYourExplores(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Push(g, me.ID, "Wildgrowth Walker", "Creature — Elemental", oldB228Walker, 1, 3)
	scout := effectstest.B12Creature(g, me.ID, "Scout", "Creature — Human Scout", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Scout", "Creature — Human Scout", 1, 1)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	oldB228Explore(g, me, scout)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB228Counters(g, walker); got != 1 {
		t.Errorf("the Walker has %d counters, want 1", got)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	oldB228Explore(g, opp, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB228Counters(g, walker) != 1 || me.Life != life+3 {
		t.Error("an opponent's creature exploring triggered the Walker")
	}
	oldB228Explore(g, me, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB228Counters(g, walker); got != 2 {
		t.Errorf("the Walker has %d counters after exploring itself (a land on top, so no explore counter), want 2", got)
	}
}

// --- Lurking Chupacabra ---

func TestOldB228LurkingChupacabraShrinksAnOpposingCreatureWhenYouExplore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Lurking Chupacabra", "Creature — Beast Horror", oldB228Chupacabra, 2, 3)
	scout := effectstest.B12Creature(g, me.ID, "Scout", "Creature — Human Scout", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	oldB228Explore(g, me, scout)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		effectstest.PassPriorityAroundTable(t, g)
		p = effectstest.LatestPickTarget(g, me.ID)
	}
	if p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, oldB228Card(mine)); err == nil {
			t.Fatal("a creature I control was an acceptable target")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, oldB228Card(theirs)); err != nil {
			t.Fatalf("pick: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tg := effectstest.PtOf(t, g, theirs); pw != 1 || tg != 1 {
		t.Errorf("their creature is %d/%d, want 1/1", pw, tg)
	}
	if pw, tg := effectstest.PtOf(t, g, mine); pw != 3 || tg != 3 {
		t.Errorf("my creature is %d/%d, want 3/3", pw, tg)
	}
}

// --- Shadowed Caravel ---

func TestOldB228ShadowedCaravelGrowsOnYourExploresAndCrews(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	caravel := effectstest.B12Push(g, me.ID, "Shadowed Caravel", "Artifact — Vehicle", oldB228Caravel, 2, 2)
	scout := effectstest.B12Creature(g, me.ID, "Scout", "Creature — Human Scout", 2, 2)
	effectstest.AdvanceToMain(t, g)
	oldB228Explore(g, me, scout)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB228Counters(g, caravel); got != 1 {
		t.Errorf("the Caravel has %d counters, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, caravel, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{scout}}); err != nil {
		t.Fatalf("crew 2 with a power-2 creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, caravel); !c.IsCreature() {
		t.Error("the Caravel is not a creature after crewing")
	}
}

// --- Emperor's Vanguard ---

func TestOldB228EmperorsVanguardExploresOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanguard := effectstest.B12Push(g, me.ID, "Emperor's Vanguard", "Creature — Human Scout", oldB228Vanguard, 4, 3)
	effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	hand, life := me.Hand.Size(), opp.Life
	effectstest.BrAttack(t, g, vanguard)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("the opponent is at %d, want %d", opp.Life, life-4)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d (the land explored into it)", got, hand+1)
	}
}

func TestOldB228EmperorsVanguardGrowsOnANonlandExplore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vanguard := effectstest.B12Push(g, me.ID, "Emperor's Vanguard", "Creature — Human Scout", oldB228Vanguard, 4, 3)
	effectstest.SeedLibraryTop(me, "Spell", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: vanguard}, me.ID) })
	effectstest.AnswerExploreQuestion(t, g, me.ID, false)
	if got := oldB228Counters(g, vanguard); got != 1 {
		t.Errorf("the Vanguard has %d counters, want 1", got)
	}
}

// --- Deadeye Tracker ---

func TestOldB228DeadeyeTrackerExilesTwoFromOneOpponentsGraveyardAndExplores(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	tracker := effectstest.PushCatalogPermanent(g, me.ID, "Deadeye Tracker", "Creature — Human Pirate", oldB228Tracker, false)
	a := effectstest.GraveyardCardOf(g, opp, "A", "Sorcery", "{1}", "")
	b := effectstest.GraveyardCardOf(g, opp, "B", "Sorcery", "{1}", "")
	c := effectstest.GraveyardCardOf(g, third, "C", "Sorcery", "{1}", "")
	mine := effectstest.GraveyardCardOf(g, me, "Mine", "Sorcery", "{1}", "")
	effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	activate := func(targets ...uuid.UUID) error {
		effectstest.B06AddMana(me, "B", "B")
		return g.ActivateCatalogAbility(me.ID, tracker, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(targets...)})
	}
	if err := activate(a, c); err == nil {
		t.Fatal("two cards from two different graveyards were accepted")
	}
	g.PlayerByIDForEffect(me.ID).ManaPool = nil
	if err := activate(a, mine); err == nil {
		t.Fatal("a card from my own graveyard was accepted")
	}
	g.PlayerByIDForEffect(me.ID).ManaPool = nil
	if err := activate(a); err == nil {
		t.Fatal("one target was accepted for a two-target ability")
	}
	g.PlayerByIDForEffect(me.ID).ManaPool = nil
	if err := activate(a, b); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, a) || !effectstest.InExile(g, b) {
		t.Error("the two cards were not exiled")
	}
	if effectstest.InExile(g, c) || effectstest.InExile(g, mine) {
		t.Error("a card that was not targeted was exiled")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d (the Tracker explored a land into it)", got, hand+1)
	}
	if !effectstest.IsTapped(g, tracker) {
		t.Error("the Tracker was not tapped")
	}
}

// --- Huatli, Dinosaur Knight ---

func TestOldB228HuatliDinosaurKnightPlusTwoCountersADinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Dinosaur Knight", oldB228HuatliKnight, 4)
	dino := effectstest.B12Creature(g, me.ID, "Dino", "Creature — Dinosaur", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, huatli, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(bear)}); err == nil {
		t.Fatal("a non-Dinosaur was accepted")
	}
	effectstest.S82Activate(t, g, me.ID, huatli, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(dino)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB228Counters(g, dino); got != 2 {
		t.Errorf("the Dinosaur has %d counters, want 2", got)
	}
	if got := effectstest.LoyaltyOf(g, huatli); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestOldB228HuatliDinosaurKnightPlusTwoWithNoTargetIsFine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Dinosaur Knight", oldB228HuatliKnight, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, huatli, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, huatli); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestOldB228HuatliDinosaurKnightMinusThreeBites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Dinosaur Knight", oldB228HuatliKnight, 4)
	dino := effectstest.B12Creature(g, me.ID, "Dino", "Creature — Dinosaur", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, huatli, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dino, theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, theirs); got != 4 {
		t.Errorf("their creature has %d damage, want 4 (the Dinosaur's power)", got)
	}
	if got := effectstest.DamageMarkedOn(g, dino); got != 0 {
		t.Errorf("the Dinosaur took %d damage from a one-sided bite", got)
	}
}

func TestOldB228HuatliDinosaurKnightUltimatePumpsOnlyYourDinosaurs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Dinosaur Knight", oldB228HuatliKnight, 7)
	dino := effectstest.B12Creature(g, me.ID, "Dino", "Creature — Dinosaur", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Dino", "Creature — Dinosaur", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, huatli, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{dino: 7, bear: 2, theirs: 3} {
		if p, _ := effectstest.PtOf(t, g, id); p != want {
			t.Errorf("a creature has power %d, want %d", p, want)
		}
	}
}

// --- Vraska, Relic Seeker ---

func TestOldB228VraskaPlusTwoMakesAMenacingPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Relic Seeker", oldB228Vraska, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pirates := effectstest.TrekB37Tokens(g, me.ID, "Pirate")
	if len(pirates) != 1 || pirates[0].Power != 2 || pirates[0].Toughness != 2 || !game.HasKeyword(&pirates[0], "menace") {
		t.Fatalf("pirates = %+v, want one 2/2 black Pirate with menace", pirates)
	}
}

func TestOldB228VraskaMinusThreeDestroysAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Relic Seeker", oldB228Vraska, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vraska, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(land)}); err == nil {
		t.Fatal("a land was accepted for artifact, creature, or enchantment")
	}
	effectstest.S82Activate(t, g, me.ID, vraska, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature was not destroyed")
	}
	if got := len(effectstest.TrekB37Tokens(g, me.ID, "Treasure")); got != 1 {
		t.Errorf("%d Treasure tokens, want 1", got)
	}
	if got := effectstest.LoyaltyOf(g, vraska); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestOldB228VraskaMinusTenSetsALifeTotalToOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Relic Seeker", oldB228Vraska, 10)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 1 {
		t.Errorf("the opponent is at %d, want 1", opp.Life)
	}
}

// --- Huatli, Warrior Poet ---

func TestOldB228HuatliWarriorPoetPlusTwoGainsLifeEqualToGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Warrior Poet", oldB228HuatliPoet, 3)
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 5, 5)
	effectstest.B12Creature(g, opp.ID, "Their Huge", "Creature — Giant", 9, 9)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, huatli, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
}

func TestOldB228HuatliWarriorPoetZeroMakesATrampleDinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Warrior Poet", oldB228HuatliPoet, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, huatli, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	dinos := effectstest.TrekB37Tokens(g, me.ID, "Dinosaur")
	if len(dinos) != 1 || dinos[0].Power != 3 || dinos[0].Toughness != 3 || !game.HasKeyword(&dinos[0], "trample") {
		t.Fatalf("dinosaurs = %+v, want one 3/3 green Dinosaur with trample", dinos)
	}
	if got := effectstest.LoyaltyOf(g, huatli); got != 3 {
		t.Errorf("loyalty %d, want 3 (a 0 ability)", got)
	}
}

func TestOldB228HuatliWarriorPoetMinusXDividesDamageAndStopsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Warrior Poet", oldB228HuatliPoet, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 9)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 9)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 9)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, huatli, 2, game.ActivateAbilityParams{
		XValue: 4, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 3, b: 1},
	})
	if got := effectstest.LoyaltyOf(g, huatli); got != 1 {
		t.Errorf("loyalty %d after −4, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, a) != 3 || effectstest.DamageMarkedOn(g, b) != 1 || effectstest.DamageMarkedOn(g, c) != 0 {
		t.Errorf("damage a=%d b=%d c=%d, want 3 1 0", effectstest.DamageMarkedOn(g, a), effectstest.DamageMarkedOn(g, b), effectstest.DamageMarkedOn(g, c))
	}
	effectstest.BrAttack(t, g, attacker)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(id, attacker); err == nil {
			t.Error("a creature dealt damage this way blocked")
		}
	}
	if err := g.DeclareBlocker(c, attacker); err != nil {
		t.Errorf("a creature that took no damage could not block: %v", err)
	}
}

// --- Jace, Cunning Castaway ---

func TestOldB228JaceCunningCastawayPlusOneLootsWhenYourCreaturesConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Cunning Castaway", oldB228JaceCastaway, 3)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.SeedLibraryTop(me, "Drawn", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.BrAttack(t, g, a, b)
	for i := 0; i < 8 && effectstest.DiscardChoiceFor(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opp.Life != 40-4 {
		t.Fatalf("the opponent is at %d, want 36", opp.Life)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand is %d, want %d: two creatures connecting is one draw (the discard is still to come)", got, hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d after the discard, want %d", got, hand)
	}
}

func TestOldB228JaceCunningCastawayPlusOneTriggersOncePerPlayerDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Cunning Castaway", oldB228JaceCastaway, 3)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.SeedLibraryTop(me, "Drawn 2", "Sorcery", "")
	effectstest.SeedLibraryTop(me, "Drawn 1", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	if err := g.DeclareAttacker(b, third.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	for i := 0; i < 8 && effectstest.DiscardChoiceFor(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no loot happened")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("two players were dealt combat damage, so there should be a second loot")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d, want %d (two loots)", got, hand)
	}
	if got := me.Graveyard.Size(); got != 2 {
		t.Errorf("graveyard is %d, want the 2 discards", got)
	}
}

func TestOldB228JaceCunningCastawayMinusTwoMakesAnIllusionThatDiesToASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Cunning Castaway", oldB228JaceCastaway, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	illusions := effectstest.TrekB37Tokens(g, me.ID, "Illusion")
	if len(illusions) != 1 || illusions[0].Power != 2 || illusions[0].Toughness != 2 {
		t.Fatalf("illusions = %+v, want one 2/2 Illusion", illusions)
	}
	id := illusions[0].InstanceID
	// An ability that targets it does not sacrifice it ("a spell").
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: me.ID, Target: id, CardID: id})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Illusion was sacrificed with no spell targeting it")
	}
	// A spell that targets it does.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d", effectstest.S82TargetCard(id))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the Illusion survived a spell that targeted it")
	}
}

func TestOldB228JaceCunningCastawayMinusFiveMakesTwoNonlegendaryCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Cunning Castaway", oldB228JaceCastaway, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	copies := effectstest.TrekB37Tokens(g, me.ID, "Jace, Cunning Castaway")
	if len(copies) != 2 {
		t.Fatalf("%d token copies, want 2", len(copies))
	}
	for _, c := range copies {
		if c.OracleID != oldB228JaceCastaway {
			t.Error("a copy does not have Jace's oracle identity")
		}
		if strings.Contains(c.TypeLine, "Legendary") {
			t.Errorf("a copy's type line is %q, want it nonlegendary", c.TypeLine)
		}
	}
	if !g.Battlefield.Contains(jace) {
		t.Error("Jace himself left the battlefield (6 − 5 = 1 loyalty; the nonlegendary copies do not trip the legend rule)")
	}
}

// --- Jace, Ingenious Mind-Mage ---

func TestOldB228JaceIngeniousMindMageDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Ingenious Mind-Mage", oldB228JaceMind, 5)
	effectstest.SeedLibraryTop(me, "Top", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
	if got := effectstest.LoyaltyOf(g, jace); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestOldB228JaceIngeniousMindMageUntapsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Ingenious Mind-Mage", oldB228JaceMind, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear); _ = g.TapTargetForEffect(theirs) })
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
		t.Error("my Bear is still tapped")
	}
	if !effectstest.IsTapped(g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
	if got := effectstest.LoyaltyOf(g, jace); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestOldB228JaceIngeniousMindMageMinusNineStealsUpToThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Ingenious Mind-Mage", oldB228JaceMind, 9)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	d := effectstest.B12Creature(g, opp.ID, "D", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, jace, 2, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(a, b, c, d)}); err == nil {
		t.Fatal("four targets were accepted for up to three")
	}
	effectstest.S82Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b, c)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if got := effectstest.ControllerOf(t, g, id); got != me.ID {
			t.Errorf("a target is controlled by %v, want me", got)
		}
	}
	if got := effectstest.ControllerOf(t, g, d); got != opp.ID {
		t.Error("the fourth creature changed hands")
	}
}
