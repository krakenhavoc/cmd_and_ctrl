package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	matB04CalixOracle   = "1bcc8ca6-5fd0-4fc9-8f0c-f81d8abf9249"
	matB04DrifterOracle = "727ef984-1643-4a09-8283-a665508e85dd"
	matB04JolraelOracle = "65e4fbd1-4935-495e-9897-bad5a0cfd935"
	matB04SparkOracle   = "82e5642a-0803-432c-85d3-091fca75c82f"
	matB04ScavOracle    = "d27f3338-4900-49eb-9d8b-994afd524abe"
	matB04TazriOracle   = "13a3e61d-c9ff-447d-9016-09a183b7af1d"
	matB04PingerOracle  = "62e3f285-886c-414e-b4ff-403a7c01c23a"
	matB04BirdsOracle   = "d3a0b660-358c-41bd-9cd2-41fbf3491b1a"
)

// matB04PassUntilConfirm passes priority until a yes/no prompt for me opens.
func matB04PassUntilConfirm(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := 0; i < 10; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && (c.Kind == game.PendingChoiceConfirm || c.Kind == game.PendingChoiceTriggerPrompt) && c.Chooser == me.ID {
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("no yes/no prompt for %s: %+v", me.Name, g.PendingChoices)
}

// matB04Answer says yes to the open "you may" prompt.
func matB04Answer(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Chooser != me.ID {
			continue
		}
		switch c.Kind {
		case game.PendingChoiceTriggerPrompt:
			if err := g.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			return
		case game.PendingChoiceConfirm:
			if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			return
		}
	}
	t.Fatalf("no yes/no prompt: %+v", g.PendingChoices)
}

// --- Calix, Guided by Fate -------------------------------------------------

func matB04Calix(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Calix, Guided by Fate", "Legendary Enchantment Creature — Human Druid", matB04CalixOracle, 2, 2)
}

func TestMatB04CalixConstellationPutsACounterOnATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	calix := matB04Calix(g, me)
	bear := effectstest.P1Creature(g, me.ID, 2, 2)
	effectstest.ToMainForCost(t, g)
	effectstest.CastCatalogSpell(t, g, "Spark Rupture", "Enchantment", matB04SparkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	if effectstest.TriggerOrderPrompt(g) != nil {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.EffectiveCardForTest(t, g, bear); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
	if c := effectstest.EffectiveCardForTest(t, g, calix); c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("Calix has %d counters, the target was the bear", c.Counters[game.CounterPlusOne])
	}
}

func TestMatB04CalixOwnEntryTriggersConstellation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	effectstest.ToMainForCost(t, g)
	calix := effectstest.CastAndResolveCreature(t, g, "Calix, Guided by Fate", "Legendary Enchantment Creature — Human Druid", matB04CalixOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, calix)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.EffectiveCardForTest(t, g, calix); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("Calix has %d counters, want 1 from entering", c.Counters[game.CounterPlusOne])
	}
}

func TestMatB04CalixCopiesANonlegendaryEnchantmentOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	calix := matB04Calix(g, me)
	ench := effectstest.PushPermanentForTest(g, me.ID, "Test Enchantment", "", "Enchantment")
	legend := effectstest.PushPermanentForTest(g, me.ID, "Legend Enchantment", "", "Legendary Enchantment")
	bear := effectstest.P1Creature(g, me.ID, 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Aura", TypeLine: "Enchantment — Aura",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})

	effectstest.AttackWith(t, g, opp.ID, calix, bear)
	matB04PassUntilConfirm(t, g, me)
	matB04Answer(t, g, me)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no permanent pick: %+v", g.PendingChoices)
	}
	if effectstest.HasID(pick.ChooseCards, legend) || effectstest.HasID(pick.ChooseCards, calix) {
		t.Fatalf("a legendary enchantment is not on offer: %v", pick.ChooseCards)
	}
	if !effectstest.HasID(pick.ChooseCards, ench) {
		t.Fatalf("the nonlegendary enchantment is on offer: %v", pick.ChooseCards)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{ench}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The enchanted bear connected too: its trigger finds the turn used up.
	if len(effectstest.TokensNamed(g, "Test Enchantment")) != 1 {
		t.Fatalf("%d Test Enchantment tokens, want exactly 1 this turn", len(effectstest.TokensNamed(g, "Test Enchantment")))
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("a second copy was offered in the same turn")
	}
}

// --- Vesuvan Drifter ---------------------------------------------------------

func TestMatB04VesuvanDrifterCopiesARevealedCreatureCardWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	drifter := effectstest.B12Push(g, me.ID, "Vesuvan Drifter", "Creature — Shapeshifter", matB04DrifterOracle, 2, 4)
	effectstest.SeedLibraryTop(me, "Hill Giant", "Creature — Giant", "{3}{R}")
	me.Library.Cards[len(me.Library.Cards)-1].Power = 3
	me.Library.Cards[len(me.Library.Cards)-1].Toughness = 3

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	matB04PassUntilConfirm(t, g, me)
	matB04Answer(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EffectiveCardForTest(t, g, drifter)
	if c.Name != "Hill Giant" || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("Drifter is %q %d/%d, want a 3/3 Hill Giant", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(&c, "flying") {
		t.Error("the copy has flying (the exception)")
	}
}

func TestMatB04VesuvanDrifterNoncreatureOrDeclinedChangesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	drifter := effectstest.B12Push(g, me.ID, "Vesuvan Drifter", "Creature — Shapeshifter", matB04DrifterOracle, 2, 4)
	effectstest.SeedLibraryTop(me, "Divination", "Sorcery", "{2}{U}")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	matB04PassUntilConfirm(t, g, me)
	matB04Answer(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.EffectiveCardForTest(t, g, drifter); c.Name != "Vesuvan Drifter" {
		t.Errorf("a noncreature card copies nothing, Drifter is %q", c.Name)
	}
}

// --- Jolrael, Voice of Zhalfir -----------------------------------------------

func TestMatB04JolraelAnimatesALandToAHandSizedBird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	effectstest.B12Push(g, me.ID, "Jolrael, Voice of Zhalfir", "Legendary Creature — Human Druid", matB04JolraelOracle, 3, 3)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	hand := len(me.Hand.Cards)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for i := 0; i < 6 && effectstest.TriggerOnStackAny(g) == false && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		_ = g.PassPriority()
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, forest)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EffectiveCardForTest(t, g, forest)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Bird") {
		t.Fatalf("the land is a Bird creature and still a land: %q", c.TypeLine)
	}
	if c.CurrentPower() != hand || c.CurrentToughness() != hand {
		t.Errorf("land is %d/%d, want %d/%d (cards in hand)", c.CurrentPower(), c.CurrentToughness(), hand, hand)
	}
	if !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "haste") {
		t.Errorf("flying and haste: %v", c.Keywords)
	}
	if cols := c.EffectiveColors(); len(cols) != 2 {
		t.Errorf("green and blue, got %v", cols)
	}
	// X is locked: drawing afterwards does not grow it.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	if c := effectstest.EffectiveCardForTest(t, g, forest); c.CurrentPower() != hand {
		t.Errorf("X changed to %d after a draw", c.CurrentPower())
	}
}

func TestMatB04JolraelDrawsWhenALandCreatureConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	effectstest.B12Push(g, me.ID, "Jolrael, Voice of Zhalfir", "Legendary Creature — Human Druid", matB04JolraelOracle, 3, 3)
	land := effectstest.B12Push(g, me.ID, "Tarn", "Land Creature — Test", "", 2, 2)
	nonland := effectstest.P1Creature(g, me.ID, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		_ = g.PassPriority()
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("no land target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, land, nonland)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 1 {
		t.Errorf("hand grew by %d, want 1 (only the land creature counts)", me.Hand.Size()-hand)
	}
}

// --- Spark Rupture ---------------------------------------------------------------

func TestMatB04SparkRuptureMakesLoyaltyWalkersCreaturesSizedByTheirCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	effectstest.PushPermanentForTest(g, me.ID, "Spark Rupture", matB04SparkOracle, "Enchantment")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Walker",
		TypeLine: "Legendary Planeswalker — Test", Keywords: []string{"hexproof"}, Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(walker, game.CounterLoyalty, 4); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	c := effectstest.EffectiveCardForTest(t, g, walker)
	if !c.IsCreature() || !c.IsPlaneswalker() {
		t.Fatalf("an opponent's walker is a creature and still a planeswalker: %q", c.TypeLine)
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("walker is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if game.HasKeyword(&c, "hexproof") {
		t.Error("it loses all abilities")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(walker, game.CounterLoyalty, 2); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	if c := effectstest.EffectiveCardForTest(t, g, walker); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("after +2 loyalty the walker is %d/%d, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(walker, game.CounterLoyalty, -5); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	if c := effectstest.EffectiveCardForTest(t, g, walker); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("after -5 loyalty the walker is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestMatB04SparkRuptureLeavesCountersFreeWalkersAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	effectstest.PushPermanentForTest(g, me.ID, "Spark Rupture", matB04SparkOracle, "Enchantment")
	walker := effectstest.PushPermanentForTest(g, me.ID, "Test Walker", "", "Legendary Planeswalker — Test")
	if c := effectstest.EffectiveCardForTest(t, g, walker); c.IsCreature() {
		t.Error("a planeswalker with no loyalty counters is not a creature")
	}
}

func TestMatB04SparkRuptureDrawsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	effectstest.ToMainForCost(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Spark Rupture", "Enchantment", matB04SparkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn", hand, me.Hand.Size())
	}
}

// --- Urborg Scavengers ----------------------------------------------------------------

func matB04Flyer(p *game.Player, kws ...string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Skyline Bird", TypeLine: "Creature — Bird",
		Keywords: kws, Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return id
}

func TestMatB04UrborgScavengersExilesGrowsAndTakesKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	scav := effectstest.B12Push(g, me.ID, "Urborg Scavengers", "Creature — Spirit", matB04ScavOracle, 2, 2)
	bird := matB04Flyer(opp, "flying", "lifelink", "defender")

	matB04Attack(t, g, scav, opp.ID)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bird)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, bird) {
		t.Fatal("the target card is exiled")
	}
	c := effectstest.EffectiveCardForTest(t, g, scav)
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("%d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
	for _, kw := range []string{"flying", "lifelink"} {
		if !game.HasKeyword(&c, kw) {
			t.Errorf("Scavengers has %s from the exiled card", kw)
		}
	}
	if game.HasKeyword(&c, "defender") {
		t.Error("defender is not in the list")
	}
	if game.HasKeyword(&c, "trample") {
		t.Error("it has no trample")
	}
}

func TestMatB04UrborgScavengersLosesTheKeywordWhenTheCardLeavesExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	scav := effectstest.B12Push(g, me.ID, "Urborg Scavengers", "Creature — Spirit", matB04ScavOracle, 2, 2)
	bird := matB04Flyer(opp, "flying")
	matB04Attack(t, g, scav, opp.ID)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bird)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.EffectiveCardForTest(t, g, scav); !game.HasKeyword(&c, "flying") {
		t.Fatal("flying while the card stays exiled")
	}
	_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, bird)
	if c := effectstest.EffectiveCardForTest(t, g, scav); game.HasKeyword(&c, "flying") {
		t.Error("the card left exile: no more flying")
	}
}

func TestMatB04UrborgScavengersTriggersOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := matB04Seats(g)
	bird := matB04Flyer(opp, "reach")
	effectstest.ToMainForCost(t, g)
	scav := effectstest.CastAndResolveCreature(t, g, "Urborg Scavengers", "Creature — Spirit", matB04ScavOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bird)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EffectiveCardForTest(t, g, scav)
	if c.Counters[game.CounterPlusOne] != 1 || !game.HasKeyword(&c, "reach") {
		t.Errorf("counters %d, reach %v", c.Counters[game.CounterPlusOne], game.HasKeyword(&c, "reach"))
	}
}

// --- Tazri, Stalwart Survivor -----------------------------------------------------------

func matB04Colored(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		OracleID: oracle, Colors: colors, Power: 3, Toughness: 3, Owner: owner, Controller: owner})
}

func TestMatB04TazriLetsOnlyCreaturesWithAnotherActivatedAbilityTapForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	tazri := matB04Colored(g, me.ID, "Tazri, Stalwart Survivor", "Legendary Creature — Human Warrior", matB04TazriOracle, "W")
	vanilla := matB04Colored(g, me.ID, "Bear", "Creature — Bear", "", "G")
	if err := effectstest.TapGrantedMana(t, g, me.ID, vanilla, "G"); err == nil {
		t.Fatal("a creature with no other activated ability can't tap for the mana")
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, tazri, "W"); err != nil {
		t.Fatalf("Tazri has another activated ability (her mill) and taps for W: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool holds %d mana, want 1", got)
	}
}

func TestMatB04TazriManaMakesOnlyTheCreaturesColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	matB04Colored(g, me.ID, "Tazri, Stalwart Survivor", "Legendary Creature — Human Warrior", matB04TazriOracle, "W")
	// Any activated ability will do: Cryptolith Rite's own grant is another.
	effectstest.GaPut(g, me.ID, "Cryptolith Rite", "Enchantment", effectstest.GaCryptolithRiteOracle)
	bear := matB04Colored(g, me.ID, "Bear", "Creature — Bear", "", "G", "U")
	var rows int
	_, origins := game.ManaAbilitiesWithOrigins(effectstest.GaLayered(t, g, bear))
	for _, o := range origins {
		if o.Granted() {
			rows++
		}
	}
	if rows != 2 {
		t.Fatalf("the bear has %d granted mana rows, want 2 (Tazri's and Cryptolith Rite's)", rows)
	}
}

func TestMatB04TazriMillsFiveAndTakesCreaturesWithActivatedAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := matB04Seats(g)
	tazri := matB04Colored(g, me.ID, "Tazri, Stalwart Survivor", "Legendary Creature — Human Warrior", matB04TazriOracle, "W")
	// Top five, from the top: a mana-ability-only creature (not taken), a
	// creature with an activated ability (taken), a noncreature with one
	// (not taken), a vanilla creature (not taken), and a land.
	var ids []uuid.UUID
	push := func(c game.Card) {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	push(game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	push(game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	push(game.Card{Name: "Wand", TypeLine: "Artifact", OracleID: "", Keywords: nil})
	push(game.Card{Name: "Abbey Matron", TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 1, OracleID: matB04PingerOracle})
	push(game.Card{Name: "Birds", TypeLine: "Creature — Bird", Power: 0, Toughness: 1, OracleID: matB04BirdsOracle})
	effectstest.FloatForTest(g, me, "WUBRG")
	if err := g.ActivateCatalogAbility(me.ID, tazri, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pinger, birds := ids[3], ids[4]
	if !me.Hand.Contains(pinger) {
		t.Error("a creature with a non-mana activated ability goes to hand")
	}
	if me.Hand.Contains(birds) || !me.Graveyard.Contains(birds) {
		t.Error("a creature with only a mana ability stays in the graveyard")
	}
	for _, id := range []uuid.UUID{ids[0], ids[1], ids[2]} {
		if me.Hand.Contains(id) || !me.Graveyard.Contains(id) {
			t.Errorf("card %s stays in the graveyard", id)
		}
	}
}
