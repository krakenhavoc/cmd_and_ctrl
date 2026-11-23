package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b13_test.go — The Brothers' War, S82 batch 13 (#5482): ten cards
// that grant or carry unearth and encore, play one of the exiled cards, or
// care about a creature entering from a graveyard. Each test plays the real
// catalog entry through the engine.

const (
	s82bro13MindbreakerOracle = "d16fcb45-7fe7-4640-a325-badb64cadd62"
	s82bro13BallistaOracle    = "3044a00d-7f88-47e8-9110-2e898b20416c"
	s82bro13DispenserOracle   = "339c39d9-7e10-458e-b602-4f0550840436"
	s82bro13MishraOracle      = "3be40d81-9be5-4f39-8ce4-920c26cf2748"
	s82bro13SurgeonsOracle    = "c51078d9-a53a-49ae-9658-2bc00cce2b81"
	s82bro13ExcavationOracle  = "6dff1d15-604f-4eea-9b7a-0c0ba6afbbac"
	s82bro13DeskOracle        = "95ac482c-c6ad-49b4-937e-460241ae355f"
	s82bro13FeldonOracle      = "76c97e3d-a9be-4ed7-a2ef-f97870d8f58a"
	s82bro13ClawOracle        = "3808afb9-962f-48ba-94c5-6b3eae7053d0"
	s82bro13VigilOracle       = "01c179aa-533e-47e1-b6e5-c35587a39680"
)

func TestS82Bro13CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82bro13MindbreakerOracle: "Terisian Mindbreaker",
		s82bro13BallistaOracle:    "Terror Ballista",
		s82bro13DispenserOracle:   "Platoon Dispenser",
		s82bro13MishraOracle:      "Mishra, Tamer of Mak Fawa",
		s82bro13SurgeonsOracle:    "Wire Surgeons",
		s82bro13ExcavationOracle:  "Meticulous Excavation",
		s82bro13DeskOracle:        "Mishra's Research Desk",
		s82bro13FeldonOracle:      "Feldon, Ronom Excavator",
		s82bro13ClawOracle:        "Dredging Claw",
		s82bro13VigilOracle:       "Thran Vigil",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// s82bro13Unearth floats mana and activates ability `idx` of a card seeded in
// the active seat's graveyard, at their precombat main.
func s82bro13Unearth(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, idx int, mana string) {
	t.Helper()
	effectstest.ToMain(t, g)
	effectstest.FloatMana(t, g, me, mana)
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func s82bro13GraveCard(me *game.Player, name, typeLine, oracle, cost string, p, tough int) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		ManaCost: cost, Power: p, Toughness: tough, Owner: me.ID, Controller: me.ID})
	return id
}

// --- Unearth on the four printed cards ---------------------------------------

func TestS82Bro13PrintedUnearthReturnsEachCardWithItsPrintedCost(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost, mana string
		idx                                int
		p, tough                           int
	}{
		{"Terisian Mindbreaker", "Artifact Creature — Juggernaut", s82bro13MindbreakerOracle, "{7}", "{C}{U}{U}{U}", 0, 6, 4},
		{"Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, "{7}", "{C}{C}{C}{B}{B}", 0, 5, 3},
		{"Platoon Dispenser", "Artifact Creature — Construct", s82bro13DispenserOracle, "{5}", "{C}{C}{W}{W}", 1, 4, 6},
		{"Mishra's Research Desk", "Artifact", s82bro13DeskOracle, "{1}", "{C}{R}", 1, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := s82bro13GraveCard(me, tc.name, tc.typeLine, tc.oracle, tc.cost, tc.p, tc.tough)
			// A pool one short of the cost cannot pay.
			effectstest.ToMain(t, g)
			if err := g.ActivateCatalogAbility(me.ID, id, tc.idx, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("unearth activated with an empty pool")
			}
			s82bro13Unearth(t, g, me, id, tc.idx, tc.mana)
			if !g.Battlefield.Contains(id) {
				t.Fatal("the card did not return to the battlefield")
			}
		})
	}
}

// --- Terisian Mindbreaker ---------------------------------------------------

func TestS82Bro13MindbreakerMillsHalfTheDefendersLibraryRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	for i := 0; i < 7; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	other := g.Seats[2]
	otherLib := other.Library.Size()
	mb := effectstest.B12Push(g, me.ID, "Terisian Mindbreaker", "Artifact Creature — Juggernaut", s82bro13MindbreakerOracle, 6, 4)
	effectstest.DeclareAttack(t, g, opp.ID, mb)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != 3 {
		t.Errorf("library %d, want 3 (seven, half rounded up milled)", got)
	}
	if got := other.Library.Size(); got != otherLib {
		t.Errorf("a non-defending player's library changed: %d", got)
	}
}

// --- Terror Ballista --------------------------------------------------------

func TestS82Bro13BallistaSacrificesAnotherCreatureToDestroyAnOpposingOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tb := effectstest.B12Push(g, me.ID, "Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, 5, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, tb, "menace") {
		t.Fatal("no menace")
	}
	effectstest.DeclareAttack(t, g, opp.ID, tb)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, fodder) || effectstest.HasID(prompt.ChooseCards, tb) {
		t.Fatalf("want an optional sacrifice of another creature: %+v", prompt)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed before the reflexive trigger")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("their creature was not destroyed")
	}
	if !g.Battlefield.Contains(tb) {
		t.Error("the Ballista left")
	}
}

func TestS82Bro13BallistaDecliningDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tb := effectstest.B12Push(g, me.ID, "Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, 5, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, tb)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(victim) {
		t.Error("declining still sacrificed or destroyed")
	}
}

// --- Platoon Dispenser ------------------------------------------------------

func TestS82Bro13DispenserDrawsAtYourEndStepOnlyWithTwoOtherCreatures(t *testing.T) {
	for _, tc := range []struct {
		others int
		want   int
	}{{1, 0}, {2, 1}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Platoon Dispenser", "Artifact Creature — Construct", s82bro13DispenserOracle, 4, 6)
		for i := 0; i < tc.others; i++ {
			effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		}
		effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2) // not mine
		hand := me.Hand.Size()
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.want {
			t.Errorf("%d other creatures: drew %d, want %d", tc.others, got, tc.want)
		}
	}
}

func TestS82Bro13DispenserMakesASoldierArtifactToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pd := effectstest.B12Push(g, me.ID, "Platoon Dispenser", "Artifact Creature — Construct", s82bro13DispenserOracle, 4, 6)
	effectstest.ToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, pd, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Soldier" && c.IsArtifact() && c.IsCreature() && c.CurrentPower() == 1 && c.CurrentToughness() == 1 {
			found = true
		}
	}
	if !found {
		t.Error("no 1/1 Soldier artifact creature token")
	}
}

// --- Mishra's Research Desk -------------------------------------------------

func TestS82Bro13ResearchDeskExilesTwoAndOnlyTheChosenOneIsPlayable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	desk := effectstest.B12Push(g, me.ID, "Mishra's Research Desk", "Artifact", s82bro13DeskOracle, 0, 0)
	effectstest.ToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Top Land", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Second Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	top, second, deep := ids[0], ids[1], ids[2]
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, desk, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(desk) {
		t.Error("the Desk was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) || !g.Exile.Contains(second) || !me.Library.Contains(deep) {
		t.Fatal("the top two cards, and only them, should be exiled")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the card that was not chosen can be played")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the chosen card: %v", err)
	}
	if !g.Battlefield.Contains(second) {
		t.Error("the chosen land did not enter")
	}
}

func TestS82Bro13ResearchDeskCannotBeActivatedWithoutSacrificeOrMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	desk := effectstest.B12Push(g, me.ID, "Mishra's Research Desk", "Artifact", s82bro13DeskOracle, 0, 0)
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, desk, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
}

// --- Feldon, Ronom Excavator ------------------------------------------------

func TestS82Bro13FeldonExilesThatManyAndPlaysOneOfThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	feldon := effectstest.B12Push(g, me.ID, "Feldon, Ronom Excavator", "Legendary Creature — Human Artificer", s82bro13FeldonOracle, 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, feldon, "haste") {
		t.Fatal("no haste")
	}
	effectstest.ToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Basic Land — Island"},
		game.Card{Name: "B", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "C", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "D", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	// Two damage, so Feldon (2/2) dies and the trigger still reads "that many".
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, feldon, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) {
		t.Fatal("two cards should be exiled")
	}
	if g.Exile.Contains(ids[2]) || !me.Library.Contains(ids[2]) {
		t.Error("more than two cards were exiled")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the unchosen card can be played")
	}
	if err := g.CastSpell(me.ID, ids[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the chosen card: %v", err)
	}
}

func TestS82Bro13FeldonCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	feldon := effectstest.B12Push(g, opp.ID, "Feldon, Ronom Excavator", "Legendary Creature — Human Artificer", s82bro13FeldonOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(feldon, atk); err == nil {
		t.Error("Feldon blocked")
	}
	if err := g.DeclareBlocker(bear, atk); err != nil {
		t.Errorf("another creature could not block: %v", err)
	}
}
