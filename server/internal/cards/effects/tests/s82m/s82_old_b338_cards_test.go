package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b338_cards_test.go — OLD batch 338 (Fate Reforged: counters, card
// draw, graveyard recursion, #5552): twelve cards played through the engine.

const (
	oldB338WardenOracle     = "0068a341-92a5-470e-a565-9ace69eb04a8"
	oldB338ShamanOracle     = "4f87eb3c-c261-4d92-a260-4b20daddb601"
	oldB338DaghatarOracle   = "7b303881-ac8d-4459-b04e-768a6804763c"
	oldB338AwakeningOracle  = "b6e87f78-d4dd-44b2-bf70-92436ce583c9"
	oldB338ScrollOracle     = "f2c7adde-492b-4585-8385-42d7e0d6e914"
	oldB338FascinationOracl = "28f0765a-de09-4be3-b952-aee75b9ede03"
	oldB338ChampionOracle   = "54af9e23-459e-470d-9b0e-e91a0fa45c98"
	oldB338StrikeOracle     = "a01491b7-89c8-4d6e-a516-431440ff3e88"
	oldB338BeastmasterOracl = "aab463bb-e790-403b-a88d-06ef81c7302b"
	oldB338WoeReaperOracle  = "636ed3e3-c31d-403a-a3e7-03e01350ce4f"
	oldB338RallyOracle      = "9d612686-2d7f-49a3-b9c4-1e99a30134f6"
	oldB338ReclamationOracl = "c34dec60-031a-4cd3-b2e5-db7a02a34d96"
)

func TestOldB338EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB338WardenOracle:     "Warden of the First Tree",
		oldB338ShamanOracle:     "Shaman of the Great Hunt",
		oldB338DaghatarOracle:   "Daghatar the Adamant",
		oldB338AwakeningOracle:  "Fearsome Awakening",
		oldB338ScrollOracle:     "Scroll of the Masters",
		oldB338FascinationOracl: "Fascination",
		oldB338ChampionOracle:   "Wandering Champion",
		oldB338StrikeOracle:     "Cunning Strike",
		oldB338BeastmasterOracl: "Abzan Beastmaster",
		oldB338WoeReaperOracle:  "Mardu Woe-Reaper",
		oldB338RallyOracle:      "Rally the Ancestors",
		oldB338ReclamationOracl: "Sudden Reclamation",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB338In(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func oldB338Counters(g *game.Game, id uuid.UUID) int {
	return effectstest.CounterCount(g, id, game.CounterPlusOne)
}

// oldB338Main puts the table in the active seat's first main phase.
func oldB338Main(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	return g.Seats[g.Turn.ActiveSeat]
}

func oldB338Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, idx int, mana string, targets ...uuid.UUID) {
	t.Helper()
	effectstest.Acrb1AddMana(t, g, me.ID, mana)
	if err := g.ActivateCatalogAbility(me.ID, src, idx, game.ActivateAbilityParams{Targets: effectstest.CardRefs(targets...)}); err != nil {
		t.Fatalf("activate row %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Warden of the First Tree ------------------------------------------------

func TestOldB338WardenClimbsItsLadder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB338Main(t, g)
	w := effectstest.B12Push(g, me.ID, "Warden of the First Tree", "Creature — Human", oldB338WardenOracle, 1, 1)

	oldB338Activate(t, g, me, w, 0, "{C}{B}")
	c := effectstest.GaLayered(t, g, w)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !c.HasSubtype("Human") || !c.HasSubtype("Warrior") {
		t.Fatalf("after the first ability: %d/%d, Warrior=%v, want a 3/3 Human Warrior", c.CurrentPower(), c.CurrentToughness(), c.HasSubtype("Warrior"))
	}

	oldB338Activate(t, g, me, w, 1, "{C}{C}{B}{B}")
	c = effectstest.GaLayered(t, g, w)
	if !c.HasSubtype("Spirit") || !c.HasSubtype("Warrior") || !c.HasSubtype("Human") || c.CurrentPower() != 3 {
		t.Fatalf("after the second ability: Spirit=%v Warrior=%v power %d", c.HasSubtype("Spirit"), c.HasSubtype("Warrior"), c.CurrentPower())
	}
	for _, kw := range []string{"trample", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, w, kw) {
			t.Errorf("the Spirit Warrior lacks %s", kw)
		}
	}

	oldB338Activate(t, g, me, w, 2, "{C}{C}{C}{B}{B}{B}")
	c = effectstest.GaLayered(t, g, w)
	if oldB338Counters(g, w) != 5 || c.CurrentPower() != 8 || c.CurrentToughness() != 8 {
		t.Errorf("after the third ability: %d counters, %d/%d, want 5 counters and 8/8", oldB338Counters(g, w), c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB338WardensLaterAbilitiesNeedTheEarlierTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB338Main(t, g)
	w := effectstest.B12Push(g, me.ID, "Warden of the First Tree", "Creature — Human", oldB338WardenOracle, 1, 1)
	oldB338Activate(t, g, me, w, 1, "{C}{C}{B}{B}")
	if c := effectstest.GaLayered(t, g, w); c.HasSubtype("Spirit") || effectstest.HasEffectiveKeyword(t, g, w, "lifelink") {
		t.Error("a Human that is not a Warrior became a Spirit")
	}
	oldB338Activate(t, g, me, w, 2, "{C}{C}{C}{B}{B}{B}")
	if oldB338Counters(g, w) != 0 {
		t.Error("a non-Spirit got counters")
	}
	oldB338Activate(t, g, me, w, 0, "{C}{B}")
	oldB338Activate(t, g, me, w, 2, "{C}{C}{C}{B}{B}{B}")
	if oldB338Counters(g, w) != 0 {
		t.Error("a Warrior that is not a Spirit got counters")
	}
}

// --- Shaman of the Great Hunt --------------------------------------------------

func TestOldB338ShamanPutsACounterOnTheCreatureThatConnected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shaman of the Great Hunt", "Creature — Orc Shaman", oldB338ShamanOracle, 4, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB338Counters(g, bear) != 1 {
		t.Errorf("the creature that dealt damage has %d counters, want 1", oldB338Counters(g, bear))
	}
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB338Counters(g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

func TestOldB338ShamanFerociousDrawsForEachBigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB338Main(t, g)
	shaman := effectstest.B12Push(g, me.ID, "Shaman of the Great Hunt", "Creature — Orc Shaman", oldB338ShamanOracle, 4, 2)
	effectstest.B12Push(g, me.ID, "Big", "Creature — Giant", "", 5, 5)
	effectstest.B12Push(g, me.ID, "Small", "Creature — Bear", "", 3, 3)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	oldB338Activate(t, g, me, shaman, 0, "{C}{C}{G}{U}")
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (the Shaman and the Giant have power 4 or greater)", got)
	}
}

// --- Daghatar the Adamant ---------------------------------------------------------

func TestOldB338DaghatarEntersWithFourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB338Main(t, g)
	id := effectstest.CastAndResolveCreature(t, g, "Daghatar the Adamant", "Legendary Creature — Human Warrior", oldB338DaghatarOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || oldB338Counters(g, id) != 4 {
		t.Errorf("on battlefield=%v with %d counters, want 4", g.Battlefield.Contains(id), oldB338Counters(g, id))
	}
}

func TestOldB338DaghatarMovesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB338Main(t, g)
	d := effectstest.B12Push(g, me.ID, "Daghatar the Adamant", "Legendary Creature — Human Warrior", oldB338DaghatarOracle, 0, 0)
	effectstest.B28AddCounter(t, g, d, 4)
	from := effectstest.B12Push(g, me.ID, "From", "Creature — Bear", "", 2, 2)
	effectstest.B28AddCounter(t, g, from, 1)
	to := effectstest.B12Push(g, me.ID, "To", "Creature — Bear", "", 2, 2)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{B}{G}")
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(from, from)}); err == nil {
		t.Fatal("one creature was accepted for both targets")
	}
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(from, to)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB338Counters(g, from) != 0 || oldB338Counters(g, to) != 1 {
		t.Errorf("From has %d, To has %d, want 0 and 1", oldB338Counters(g, from), oldB338Counters(g, to))
	}
}

func TestOldB338DaghatarMovesNothingFromACounterlessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB338Main(t, g)
	d := effectstest.B12Push(g, me.ID, "Daghatar the Adamant", "Legendary Creature — Human Warrior", oldB338DaghatarOracle, 4, 4)
	from := effectstest.B12Push(g, me.ID, "From", "Creature — Bear", "", 2, 2)
	to := effectstest.B12Push(g, me.ID, "To", "Creature — Bear", "", 2, 2)
	oldB338Activate(t, g, me, d, 0, "{C}{B}{G}", from, to)
	if oldB338Counters(g, to) != 0 {
		t.Error("a counter appeared from nowhere")
	}
}

// --- Fearsome Awakening ----------------------------------------------------------------

func TestOldB338FearsomeAwakeningCountersOnlyADragon(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{{"Creature — Dragon", 2}, {"Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		me := oldB338Main(t, g)
		target := effectstest.PushGraveyardPermanent(me, "Target", tc.typeLine, "{2}")
		spell := effectstest.CastCatalogSpell(t, g, "Fearsome Awakening", "Sorcery", oldB338AwakeningOracle, effectstest.CardRefs(target))
		_ = spell
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(target) {
			t.Fatalf("%s: the creature did not return", tc.typeLine)
		}
		if got := oldB338Counters(g, target); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.typeLine, got, tc.want)
		}
	}
}

// --- Scroll of the Masters ------------------------------------------------------------

func TestOldB338ScrollCountsNoncreatureSpellsAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB338Main(t, g)
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Scroll of the Masters", "Artifact", oldB338ScrollOracle, false)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Hill Giant", "Creature — Giant", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, scroll, game.CounterLore); n != 0 {
		t.Fatalf("a creature spell put %d lore counters on the Scroll", n)
	}
	effectstest.B19CastXModal(t, g, "Fascination", "Sorcery", oldB338FascinationOracl, "{X}{U}{U}", 0, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, scroll, game.CounterLore); n != 1 {
		t.Fatalf("a noncreature spell put %d lore counters on the Scroll, want 1", n)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(scroll, game.CounterLore, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	oldB338Activate(t, g, me, scroll, 0, "{C}{C}{C}", bear)
	if c := effectstest.GaLayered(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4 (two lore counters)", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Fascination ---------------------------------------------------------------------------

func TestOldB338FascinationDrawsOrMillsXForEachPlayer(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		g := effectstest.NewCatalogGame(t)
		oldB338Main(t, g)
		hands, libs, yards := map[uuid.UUID]int{}, map[uuid.UUID]int{}, map[uuid.UUID]int{}
		for _, s := range g.Seats {
			effectstest.SeedLibrary(s, "A", "B", "C", "D", "E")
			hands[s.ID], libs[s.ID], yards[s.ID] = s.Hand.Size(), s.Library.Size(), s.Graveyard.Size()
		}
		effectstest.B19CastXModal(t, g, "Fascination", "Sorcery", oldB338FascinationOracl, "{X}{U}{U}", 2, []int{mode}, nil)
		effectstest.PassPriorityAroundTable(t, g)
		for _, s := range g.Seats {
			if mode == 0 && (s.Hand.Size()-hands[s.ID] != 2 || libs[s.ID]-s.Library.Size() != 2) {
				t.Errorf("draw mode: a player's hand grew by %d and library fell by %d, want 2 and 2", s.Hand.Size()-hands[s.ID], libs[s.ID]-s.Library.Size())
			}
			if mode == 1 && (libs[s.ID]-s.Library.Size() != 2 || s.Hand.Size() != hands[s.ID]) {
				t.Errorf("mill mode: a library fell by %d, want 2 (hand %d → %d)", libs[s.ID]-s.Library.Size(), hands[s.ID], s.Hand.Size())
			}
		}
	}
}

// --- Wandering Champion ---------------------------------------------------------------------

func TestOldB338WanderingChampionRummagesWithABlueOrRedPermanent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		colors   []string
		rummages bool
	}{{"blue", []string{"U"}, true}, {"red", []string{"R"}, true}, {"green", []string{"G"}, false}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		champ := effectstest.B12Push(g, me.ID, "Wandering Champion", "Creature — Human Monk", oldB338ChampionOracle, 3, 1)
		other := effectstest.B12Push(g, me.ID, "Colored", "Creature — Bear", "", 1, 1)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, other).Colors = tc.colors })
		effectstest.SeedLibrary(me, "A", "B")
		effectstest.DealCombatDamageToPlayer(g, champ, opp.ID, 3)
		effectstest.PassPriorityAroundTable(t, g)
		if !tc.rummages {
			if effectstest.DiscardOwed(g, me.ID) != 0 {
				t.Errorf("%s: the trigger asked for a discard with no blue or red permanent", tc.name)
			}
			continue
		}
		if effectstest.DiscardOwed(g, me.ID) != 1 {
			t.Fatalf("%s: no optional discard was offered", tc.name)
		}
		hand := me.Hand.Size()
		effectstest.DiscardFromHand(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand {
			t.Errorf("%s: hand %d after a discard and a draw, want %d", tc.name, me.Hand.Size(), hand)
		}
	}
}

// --- Cunning Strike -------------------------------------------------------------------------------

func TestOldB338CunningStrikeDamagesBothAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB338Main(t, g)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 3, 3)
	effectstest.SeedLibrary(me, "A", "B")
	life, hand := opp.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Cunning Strike", "Instant", oldB338StrikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim, Slot: 0},
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the player's life went %d → %d, want 2 less", life, opp.Life)
	}
	if c := effectstest.FindCardAnywhere(t, g, victim); c.DamageMarked != 2 {
		t.Errorf("the creature has %d damage, want 2", c.DamageMarked)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Abzan Beastmaster ---------------------------------------------------------------------------

func TestOldB338BeastmasterDrawsOnlyWithTheGreatestToughness(t *testing.T) {
	for _, tc := range []struct {
		name         string
		theirs, mine int
		want         int
	}{{"tied", 3, 3, 1}, {"mine is greater", 2, 5, 1}, {"theirs is greater", 6, 3, 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Abzan Beastmaster", "Creature — Dog Shaman", oldB338BeastmasterOracl, 2, 1)
		effectstest.B12Push(g, me.ID, "Mine", "Creature — Wall", "", 1, tc.mine)
		effectstest.B12Push(g, opp.ID, "Theirs", "Creature — Wall", "", 1, tc.theirs)
		effectstest.SeedLibrary(me, "A", "B", "C", "D")
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		hand := me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		// The turn's draw step comes later, so the upkeep trigger alone moves the hand.
		if got := me.Hand.Size() - hand; got != tc.want {
			t.Errorf("%s: drew %d cards at upkeep, want %d", tc.name, got, tc.want)
		}
	}
}

// --- Mardu Woe-Reaper ---------------------------------------------------------------------------

func TestOldB338WoeReaperExilesACreatureCardForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB338Main(t, g)
	theirCard := effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	notCreature := effectstest.PushGraveyardPermanent(me, "Relic", "Artifact", "{1}")
	life := me.Life
	effectstest.CastAndResolveCreature(t, g, "Mardu Woe-Reaper", "Creature — Human Warrior", oldB338WoeReaperOracle)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if effectstest.HasID(prompt.PickTargetCards, notCreature) || !effectstest.HasID(prompt.PickTargetCards, theirCard) {
		t.Fatalf("targets %v: a creature card in any graveyard, and only those, are legal", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, theirCard)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB338In(opp.Graveyard, theirCard) {
		t.Error("the creature card is still in the graveyard")
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want 1 gained", life, me.Life)
	}
}

func TestOldB338WoeReaperTriggersForAnotherWarriorOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB338Main(t, g)
	effectstest.B12Push(g, me.ID, "Mardu Woe-Reaper", "Creature — Human Warrior", oldB338WoeReaperOracle, 2, 1)
	effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastAndResolveCreature(t, g, "Hill Giant", "Creature — Giant", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Warrior entering triggered the Woe-Reaper")
	}
	effectstest.CastAndResolveCreature(t, g, "Brave Warrior", "Creature — Human Warrior", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("another Warrior entering did not trigger the Woe-Reaper")
	}
}

// --- Rally the Ancestors ---------------------------------------------------------------------

func TestOldB338RallyReturnsCreaturesOfValueXOrLessThenExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB338Main(t, g)
	cheap := effectstest.PushGraveyardPermanent(me, "Cheap", "Creature — Bear", "{1}{G}")
	big := effectstest.PushGraveyardPermanent(me, "Big", "Creature — Giant", "{4}{G}")
	relic := effectstest.PushGraveyardPermanent(me, "Relic", "Artifact", "{1}")
	effectstest.Pr8CastX(t, g, "Rally the Ancestors", "Instant", oldB338RallyOracle, 2, nil)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the cheap creature did not return")
	}
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(relic) {
		t.Error("a creature with a larger mana value, or a noncreature card, returned")
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Rally the Ancestors" {
			t.Error("Rally the Ancestors is in the graveyard instead of exile")
		}
	}
	exiled := false
	for _, c := range g.Exile.Cards {
		exiled = exiled || c.Name == "Rally the Ancestors"
	}
	if !exiled {
		t.Error("Rally the Ancestors is not in exile")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the returned creature was not exiled at the beginning of my next upkeep")
	}
}

// --- Sudden Reclamation ---------------------------------------------------------------------------

func TestOldB338SuddenReclamationMillsThenReturnsACreatureAndALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB338Main(t, g)
	me.Library.Cards = nil
	effectstest.PushLibraryCard(me, game.Card{Name: "Milled Land", TypeLine: "Basic Land — Forest"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	effectstest.PushLibraryCard(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Filler Two", TypeLine: "Sorcery"})
	oldCreature := effectstest.PushGraveyardPermanent(me, "Old Creature", "Creature — Bear", "{1}{G}")
	oldLand := effectstest.PushGraveyardPermanent(me, "Old Land", "Basic Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Sudden Reclamation", "Instant", oldB338ReclamationOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	creature := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if creature == nil || len(creature.ChooseCards) != 2 {
		t.Fatalf("creature prompt = %+v, want the milled and the old creature card", creature)
	}
	if err := g.ResolveChooseCards(creature.ID, me.ID, []uuid.UUID{oldCreature}); err != nil {
		t.Fatalf("ResolveChooseCards (creature): %v", err)
	}
	land := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if land == nil || len(land.ChooseCards) != 2 {
		t.Fatalf("land prompt = %+v, want the milled and the old land card", land)
	}
	if err := g.ResolveChooseCards(land.ID, me.ID, []uuid.UUID{oldLand}); err != nil {
		t.Fatalf("ResolveChooseCards (land): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB338In(me.Hand, oldCreature) || !oldB338In(me.Hand, oldLand) {
		t.Error("the chosen creature and land cards are not in hand")
	}
	if me.Library.Size() != 0 {
		t.Errorf("library has %d cards, want the four milled", me.Library.Size())
	}
}
