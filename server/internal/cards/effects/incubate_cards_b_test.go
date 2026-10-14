package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incubate_cards_b_test.go — #3568 slice inc-b: the plain incubate
// cards (ADR 0185, #2417).

const (
	giftOfCompleationOracle   = "c00a13cc-3c2b-473f-801c-3cfcca642ac0"
	ichorDrinkerOracle        = "470a8ee9-1870-44d8-8ade-64ec8084b9f7"
	infectedDefectorOracle    = "cff8dd65-e531-4496-aa49-e8d4350b9620"
	injectorCrocodileOracle   = "619c7d51-51d1-4467-a58b-4f275ff9df0f"
	maraudingDreadshipOracle  = "12172602-89db-4c8b-b752-4f8380896589"
	phyrexianAwakeningOracle  = "0566ab5d-c804-4ac0-add2-87fc352b4c7e"
	sculptedPerfectionOracle  = "90fa1303-ad1c-46da-a4d6-4b29ea49b7ca"
	searingBarbOracle         = "9a1622a5-c982-407b-a158-7b8ad9238aac"
	tangledSkylineOracle      = "2965bae9-0d2b-42e1-b735-01ea57a72409"
	tillerOfFleshOracle       = "b6d2f392-5637-49ec-b6cb-603edd89e788"
	incubBLightningBoltOracle = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestGiftOfCompleationIncubatesThenSurveilsWhenAPhyrexianDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top Card", "Next Card")
	castCatalogSpell(t, g, "Gift of Compleation", "Enchantment", giftOfCompleationOracle, nil)
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Fatalf("incubate counters = %d, want 3", got)
	}
	// A non-Phyrexian dying does nothing.
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Fatal("a non-Phyrexian dying prompted a surveil")
	}
	// A Phyrexian does.
	phy := pushDiesCreatureForTest(g, me.ID, "Phyrexian Test", "", "Creature — Phyrexian", 2, 2)
	killCreature(t, g, me.ID, phy)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("want a surveil 1 prompt, got %+v", c)
	}
}

func TestIchorDrinkerExilesItselfFromTheGraveyardToIncubateTwo(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Ichor Drinker", "Creature — Phyrexian Vampire",
		ichorDrinkerOracle, 1, 1, "{B}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatal("the cost did not exile the card from the graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("incubate counters = %d, want 2", got)
	}
}

func TestInfectedDefectorIncubatesThreeWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Infected Defector", "Creature — Phyrexian Knight", infectedDefectorOracle, false)
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("incubate counters = %d, want 3", got)
	}
}

func TestInjectorCrocodileIncubatesWhenItDiesAndSwampcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Injector Crocodile", "Creature — Phyrexian Crocodile", injectorCrocodileOracle, false)
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("incubate counters = %d, want 3", got)
	}

	g2 := newCatalogGame(t)
	swamp := pushLibraryCardForTest(g2.Seats[g2.Turn.ActiveSeat], game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	pushLibraryCardForTest(g2.Seats[g2.Turn.ActiveSeat], game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	cid, p := cycleFromHand(t, g2, "Injector Crocodile", "Creature — Phyrexian Crocodile", injectorCrocodileOracle, "{C}{C}")
	if !p.Graveyard.Contains(cid) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g2)
	c := searchChoiceFor(g2, p.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != swamp {
		t.Fatalf("want a search offering only the Swamp, got %+v", c)
	}
	if err := g2.ResolveSearchLibrary(c.ID, p.ID, []uuid.UUID{swamp}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !p.Hand.Contains(swamp) {
		t.Error("swampcycling did not fetch the Swamp")
	}
}

func TestMaraudingDreadshipHasHasteAndIncubatesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Marauding Dreadship", "Artifact — Vehicle", maraudingDreadshipOracle, nil)
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("incubate counters = %d, want 2", got)
	}
	if !containsString(effectiveAbilities(t, g, id), "haste") {
		t.Error("the Vehicle does not have haste")
	}
}

func TestPhyrexianAwakeningIncubatesFourAndGivesPhyrexiansVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phy := pushDiesCreatureForTest(g, me.ID, "Phyrexian Test", "", "Creature — Phyrexian", 2, 2)
	theirs := pushDiesCreatureForTest(g, opp.ID, "Their Phyrexian", "", "Creature — Phyrexian", 2, 2)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Phyrexian Awakening", "Enchantment", phyrexianAwakeningOracle, nil)
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 4 {
		t.Errorf("incubate counters = %d, want 4", got)
	}
	if !containsString(effectiveAbilities(t, g, phy), "vigilance") {
		t.Error("your Phyrexian lacks vigilance")
	}
	if containsString(effectiveAbilities(t, g, theirs), "vigilance") || containsString(effectiveAbilities(t, g, bear), "vigilance") {
		t.Error("vigilance leaked to an opponent's Phyrexian or a non-Phyrexian")
	}
}

func TestSculptedPerfectionIncubatesTwoAndPumpsYourPhyrexians(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phy := pushDiesCreatureForTest(g, me.ID, "Phyrexian Test", "", "Creature — Phyrexian", 2, 2)
	theirs := pushDiesCreatureForTest(g, opp.ID, "Their Phyrexian", "", "Creature — Phyrexian", 2, 2)
	castCatalogSpell(t, g, "Sculpted Perfection", "Enchantment", sculptedPerfectionOracle, nil)
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("incubate counters = %d, want 2", got)
	}
	if got := effectivePower(t, g, phy); got != 3 {
		t.Errorf("your Phyrexian power = %d, want 3", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("their Phyrexian power = %d, want 2", got)
	}
	// A transformed Incubator is a Phyrexian Token and gets the bonus.
	inc := theIncubator(t, g, me.ID)
	transformIncubator(t, g, me, inc.InstanceID)
	if got := effectivePowerWithCounters(t, g, inc.InstanceID); got != 3 {
		t.Errorf("transformed Incubator power = %d, want 2 counters + 1", got)
	}
}

func TestSearingBarbDamagesStopsBlockingAndIncubatesOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushVanillaCreature(g, opp.ID, "Big Blocker", 3, 3)
	castCatalogSpell(t, g, "Searing Barb", "Sorcery", searingBarbOracle, []game.TargetRef{pbCard(big)})
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(big)
	if !ok {
		t.Fatal("the 3/3 died to 2 damage")
	}
	if !game.Restricted(&c, game.CantBlock) {
		t.Error("the damaged creature can still block")
	}
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 1 {
		t.Errorf("incubate counters = %d, want 1", got)
	}
}

func TestSearingBarbAtAPlayerStillIncubates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	castCatalogSpell(t, g, "Searing Barb", "Sorcery", searingBarbOracle, []game.TargetRef{pbPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d -> %d, want -2", before, opp.Life)
	}
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 1 {
		t.Errorf("incubate counters = %d, want 1", got)
	}
}

func TestTangledSkylineGainsFiveIncubatesFiveAndGivesReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phy := pushDiesCreatureForTest(g, me.ID, "Phyrexian Test", "", "Creature — Phyrexian", 2, 2)
	before := me.Life
	castCatalogSpell(t, g, "Tangled Skyline", "Enchantment", tangledSkylineOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("life %d -> %d, want +5", before, me.Life)
	}
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 5 {
		t.Errorf("incubate counters = %d, want 5", got)
	}
	if !containsString(effectiveAbilities(t, g, phy), "reach") {
		t.Error("your Phyrexian lacks reach")
	}
}

func TestTillerOfFleshIncubatesOnlyWhenASpellTargetsAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Tiller of Flesh", "Creature — Phyrexian Knight", tillerOfFleshOracle, false)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 5, 5)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", incubBLightningBoltOracle, []game.TargetRef{pbPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	if n := len(incubatorsOf(g, me.ID)); n != 0 {
		t.Fatalf("a spell targeting a player made %d Incubators", n)
	}

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", incubBLightningBoltOracle, []game.TargetRef{pbCard(victim)})
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("incubate counters = %d, want 2", got)
	}
}
