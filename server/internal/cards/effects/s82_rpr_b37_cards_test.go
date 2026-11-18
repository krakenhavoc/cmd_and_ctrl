package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b37_cards_test.go — S82 reprint pass, batch 37 (#5908): Inkmoth
// Nexus, Lavaclaw Reaches, Meteor Crater, Mirrorpool, Mystifying Maze,
// Petrified Field, Power Depot, Shambling Vent, Stirring Wildwood, Terrain
// Generator and Empress Galina.

const (
	rprB37Inkmoth   = "675281ff-b81f-4e5e-9f85-9f8cd202b50b"
	rprB37Lavaclaw  = "340de307-982a-4b26-9dc0-99113dc766cd"
	rprB37Meteor    = "ebddf68c-1b22-4313-83e1-13a5c1aec19e"
	rprB37Mirror    = "57b86d5c-3269-44bc-a838-3c5439d820d9"
	rprB37Maze      = "58bd67a8-1833-4827-aa33-1c141568f481"
	rprB37Petrified = "c4bc5bc4-e589-42c5-91fa-2ebc96448e85"
	rprB37Depot     = "64687880-03f9-4f38-985b-1027c797e33f"
	rprB37Vent      = "4725abd4-06bc-464a-bc5b-c9e0f71ec079"
	rprB37Wildwood  = "e900f211-dc8b-40f7-a217-1b1462b15c26"
	rprB37Terrain   = "a949c96c-362c-45a3-bd5c-ce5ce153ee9e"
	rprB37Galina    = "ea2614bc-00ad-4859-98f5-689f33357fe6"
)

func TestRprB37CardsAreFull(t *testing.T) {
	want := map[string]string{
		rprB37Inkmoth: "Inkmoth Nexus", rprB37Lavaclaw: "Lavaclaw Reaches", rprB37Meteor: "Meteor Crater",
		rprB37Mirror: "Mirrorpool", rprB37Maze: "Mystifying Maze", rprB37Petrified: "Petrified Field",
		rprB37Depot: "Power Depot", rprB37Vent: "Shambling Vent", rprB37Wildwood: "Stirring Wildwood",
		rprB37Terrain: "Terrain Generator", rprB37Galina: "Empress Galina",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestRprB37GuildManlandsEnterTappedAndAnimate(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, kw string
		generic          int
		colours          []string
		p, tough         int
	}{
		{"Shambling Vent", rprB37Vent, "lifelink", 1, []string{"W", "B"}, 2, 3},
		{"Stirring Wildwood", rprB37Wildwood, "reach", 1, []string{"G", "W"}, 3, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !isTapped(g, id) {
				t.Errorf("%s entered untapped", tc.name)
			}
			g2 := newCatalogGame(t)
			land := lciQRUZRestless(t, g2, tc.name, tc.oracle, tc.generic, tc.colours...)
			lciQRUZAssertAnimated(t, g2, land, "Elemental", tc.colours, tc.p, tc.tough)
			if !hasEffectiveKeyword(t, g2, land, tc.kw) {
				t.Errorf("the animated land lacks %s", tc.kw)
			}
			advancePastCleanupOf(t, g2, 0)
			if c := s58p6View(t, g2, land); c.IsCreature() || hasEffectiveKeyword(t, g2, land, tc.kw) {
				t.Errorf("after cleanup: creature %v, keyword %v", c.IsCreature(), hasEffectiveKeyword(t, g2, land, tc.kw))
			}
		})
	}
}

func TestRprB37ManlandRefusesWithoutItsColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := s58p6Push(g, me.ID, "Shambling Vent", "Land", rprB37Vent, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("animated with no {W}{B}")
	}
}

func TestRprB37InkmothNexusIsAFlyingInfectBlinkmoth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := s58p6Push(g, me.ID, "Inkmoth Nexus", "Land", rprB37Inkmoth, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	ngheAnimate(t, g, me, land, 1)
	c := s58p6View(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.IsArtifact() {
		t.Fatalf("creature %v land %v artifact %v, want all", c.IsCreature(), c.IsLand(), c.IsArtifact())
	}
	if !c.HasSubtype("Phyrexian") || !c.HasSubtype("Blinkmoth") {
		t.Errorf("subtypes %v", c.Effective().Subtypes)
	}
	if len(c.EffectiveColors()) != 0 {
		t.Errorf("colours %v, want colourless", c.EffectiveColors())
	}
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("%d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	for _, kw := range []string{"flying", "infect"} {
		if !hasEffectiveKeyword(t, g, land, kw) {
			t.Errorf("lacks %s", kw)
		}
	}
	advancePastCleanupOf(t, g, 0)
	if c := s58p6View(t, g, land); c.IsCreature() || c.IsArtifact() || hasEffectiveKeyword(t, g, land, "infect") {
		t.Error("the animation outlived the turn")
	}
}

func TestRprB37LavaclawReachesGrantsAnXPump(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := s58p6Push(g, me.ID, "Lavaclaw Reaches", "Land", rprB37Lavaclaw, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	ngheAnimate(t, g, me, land, 1, "B", "R")
	lciQRUZAssertAnimated(t, g, land, "Elemental", []string{"B", "R"}, 2, 2)
	me.ManaPool.EmptyPool()
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, land, 1, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("granted {X} pump: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, land); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
	if tough := effectiveToughness(t, g, land); tough != 2 {
		t.Errorf("toughness %d, want 2", tough)
	}
}

func TestRprB37MeteorCraterAddsAColourOfAPermanentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crater := seedPermanentWithOracle(g, me.ID, "Meteor Crater", "Land", rprB37Meteor)
	if err := g.ActivateManaAbility(me.ID, crater, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.ManaPool.Total() != 0 || manaPickFor(g, me.ID) != nil {
		t.Fatal("produced mana with no coloured permanent")
	}
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{R}",
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == crater {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, crater, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate with a red permanent: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "R" {
		t.Errorf("pool %v, want one {R}", me.ManaPool)
	}
}

func TestRprB37PetrifiedFieldReturnsALandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	field := pushCatalogPermanent(g, me.ID, "Petrified Field", "Land", rprB37Petrified, false)
	land := pushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	bear := pushGraveyardCardTyped(me, "Grizzly Bears", "Creature — Bear")
	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(bear),
	}); err == nil {
		t.Fatal("targeted a nonland card")
	}
	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(land),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, still := battlefieldCard(g, field); still {
		t.Error("the Field is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if !handHas(me, land) {
		t.Error("the land card was not returned to hand")
	}
	if !graveyardHas(me, bear) {
		t.Error("the creature card left the graveyard")
	}
}

func TestRprB37MystifyingMazeExilesAnAttackerAndReturnsItTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	maze := pushCatalogPermanent(g, opp.ID, "Mystifying Maze", "Land", rprB37Maze, false)
	bear := b12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	fillPool(opp, 4)
	if err := g.ActivateCatalogAbility(opp.ID, maze, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(idle),
	}); err == nil {
		t.Fatal("targeted a non-attacking creature")
	}
	if err := g.ActivateCatalogAbility(opp.ID, maze, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, there := battlefieldCard(g, bear); there {
		t.Fatal("the attacker was not exiled")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Attacking Bear" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("the creature did not return at the end step")
	}
	if !back.Tapped || back.Controller != me.ID {
		t.Errorf("returned tapped %v, controller %v; want tapped under its owner", back.Tapped, back.Controller == me.ID)
	}
}

func TestRprB37TerrainGeneratorPutsABasicLandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gen := pushCatalogPermanent(g, me.ID, "Terrain Generator", "Land", rprB37Terrain, false)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Nonbasic", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	basic := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: basic, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, gen, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{2} paid from an empty pool")
	}
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, gen, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	offered := chooseCardsChoiceFor(g, me.ID)
	if offered == nil {
		t.Fatal("no prompt to put a basic land")
	}
	if len(offered.ChooseCards) != 1 || offered.ChooseCards[0] != basic {
		t.Errorf("offered %v, want only the basic land %v", offered.ChooseCards, basic)
	}
	answerChooseCards(t, g, me.ID, basic)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, basic)
	if !ok {
		t.Fatal("the basic land did not enter")
	}
	if !c.Tapped {
		t.Error("the basic land entered untapped")
	}
}

func TestRprB37MirrorpoolCopiesACreatureAsAToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pool := pushCatalogPermanent(g, me.ID, "Mirrorpool", "Land", rprB37Mirror, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "G", 5)
	if err := g.ActivateCatalogAbility(me.ID, pool, 1, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(bear),
	}); err == nil {
		t.Fatal("{4}{C} paid with no colourless mana")
	}
	me.ManaPool.EmptyPool()
	fillPool(me, 4)
	fillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, pool, 1, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, still := battlefieldCard(g, pool); still {
		t.Error("Mirrorpool is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d Bears on the battlefield, want 2", n)
	}
}

func TestRprB37MirrorpoolEntersTappedAndRefusesNonSpellTargets(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Mirrorpool", "Land", rprB37Mirror, game.CastSpellParams{})
	if !isTapped(g, id) {
		t.Error("Mirrorpool entered untapped")
	}
	me := g.Seats[0]
	pool := pushCatalogPermanent(g, me.ID, "Mirrorpool", "Land", rprB37Mirror, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "C", 1)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, pool, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(bear),
	}); err == nil {
		t.Fatal("the spell-copy ability took a creature permanent as its target")
	}
}

func TestRprB37PowerDepotEntersTappedWithACounterAndRestrictedMana(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Power Depot", "Artifact Land", rprB37Depot, game.CastSpellParams{})
	if !isTapped(g, id) {
		t.Error("Power Depot entered untapped")
	}
	if c, _ := battlefieldCard(g, id); c.Counters["+1/+1"] != 1 {
		t.Errorf("counters %v, want one +1/+1 (modular 1)", c.Counters)
	}
	g2 := newCatalogGame(t)
	me := g2.Seats[0]
	depot := pushCatalogPermanent(g2, me.ID, "Power Depot", "Artifact Land", rprB37Depot, false)
	if err := g2.ActivateManaAbility(me.ID, depot, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colourless: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "C" {
		t.Fatalf("pool %v, want {C}", me.ManaPool)
	}
	if len(me.ManaPool.Units()[0].Restrictions) != 0 {
		t.Error("the {C} ability carries a restriction")
	}
	me.ManaPool.EmptyPool()
	g2.WithWriteLock(func() {
		for i := range g2.Battlefield.Cards {
			if g2.Battlefield.Cards[i].InstanceID == depot {
				g2.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g2.ActivateManaAbility(me.ID, depot, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("coloured: %v", err)
	}
	pick := manaPickFor(g2, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("pick %v, want a five-colour choice", pick)
	}
	if err := g2.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("pick: %v", err)
	}
	units := me.ManaPool.Units()
	if len(units) != 1 || units[0].Color != "G" || len(units[0].Restrictions) == 0 {
		t.Errorf("pool %v, want one restricted {G}", units)
	}
}

func TestRprB37EmpressGalinaStealsALegendaryPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	galina := pushCatalogPermanent(g, me.ID, "Empress Galina", "Legendary Creature — Merfolk Noble", rprB37Galina, false)
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	plain := b12Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "U", 2)
	if err := g.ActivateCatalogAbility(me.ID, galina, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(plain),
	}); err == nil {
		t.Fatal("targeted a nonlegendary permanent")
	}
	if err := g.ActivateCatalogAbility(me.ID, galina, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(legend),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, legend)
	if !ok || c.Controller != me.ID {
		t.Fatalf("legend controller = %v on battlefield %v, want me", c.Controller == me.ID, ok)
	}
	advancePastCleanupOf(t, g, 0)
	if c, _ := battlefieldCard(g, legend); c.Controller != me.ID {
		t.Error("control ended with the turn; it is indefinite")
	}
}

func TestRprB37MirrorpoolCopiesYourInstantOrSorcerySpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pool := pushCatalogPermanent(g, me.ID, "Mirrorpool", "Land", rprB37Mirror, false)
	before := me.Hand.Size()
	spell := castCatalogSpell(t, g, "Divination", "Sorcery", "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", nil)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, pool, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(spell),
	}); err == nil {
		t.Fatal("{2}{C} paid with no colourless-only mana")
	}
	me.ManaPool.EmptyPool()
	fillPool(me, 2)
	fillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, pool, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(spell),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, still := battlefieldCard(g, pool); still {
		t.Error("Mirrorpool is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 4 {
		t.Errorf("hand delta %d, want 4 (the spell and its copy each draw two)", got)
	}
}
