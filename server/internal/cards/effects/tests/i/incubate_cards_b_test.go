package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestGiftOfCompleationIncubatesThenSurveilsWhenAPhyrexianDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top Card", "Next Card")
	effectstest.CastCatalogSpell(t, g, "Gift of Compleation", "Enchantment", effectstest.GiftOfCompleationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Fatalf("incubate counters = %d, want 3", got)
	}
	// A non-Phyrexian dying does nothing.
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Fatal("a non-Phyrexian dying prompted a surveil")
	}
	// A Phyrexian does.
	phy := effectstest.PushDiesCreatureForTest(g, me.ID, "Phyrexian Test", "", "Creature — Phyrexian", 2, 2)
	effectstest.KillCreature(t, g, me.ID, phy)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("want a surveil 1 prompt, got %+v", c)
	}
}

func TestIchorDrinkerExilesItselfFromTheGraveyardToIncubateTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Ichor Drinker", "Creature — Phyrexian Vampire",
		effectstest.IchorDrinkerOracle, 1, 1, "{B}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatal("the cost did not exile the card from the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("incubate counters = %d, want 2", got)
	}
}

func TestInfectedDefectorIncubatesThreeWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Infected Defector", "Creature — Phyrexian Knight", effectstest.InfectedDefectorOracle, false)
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("incubate counters = %d, want 3", got)
	}
}

func TestInjectorCrocodileIncubatesWhenItDiesAndSwampcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Injector Crocodile", "Creature — Phyrexian Crocodile", effectstest.InjectorCrocodileOracle, false)
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("incubate counters = %d, want 3", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	swamp := effectstest.PushLibraryCardForTest(g2.Seats[g2.Turn.ActiveSeat], game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	effectstest.PushLibraryCardForTest(g2.Seats[g2.Turn.ActiveSeat], game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	cid, p := effectstest.CycleFromHand(t, g2, "Injector Crocodile", "Creature — Phyrexian Crocodile", effectstest.InjectorCrocodileOracle, "{C}{C}")
	if !p.Graveyard.Contains(cid) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g2)
	c := effectstest.SearchChoiceFor(g2, p.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Marauding Dreadship", "Artifact — Vehicle", effectstest.MaraudingDreadshipOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("incubate counters = %d, want 2", got)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "haste") {
		t.Error("the Vehicle does not have haste")
	}
}

func TestPhyrexianAwakeningIncubatesFourAndGivesPhyrexiansVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phy := effectstest.PushDiesCreatureForTest(g, me.ID, "Phyrexian Test", "", "Creature — Phyrexian", 2, 2)
	theirs := effectstest.PushDiesCreatureForTest(g, opp.ID, "Their Phyrexian", "", "Creature — Phyrexian", 2, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Phyrexian Awakening", "Enchantment", effectstest.PhyrexianAwakeningOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 4 {
		t.Errorf("incubate counters = %d, want 4", got)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, phy), "vigilance") {
		t.Error("your Phyrexian lacks vigilance")
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, theirs), "vigilance") || effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "vigilance") {
		t.Error("vigilance leaked to an opponent's Phyrexian or a non-Phyrexian")
	}
}

func TestSculptedPerfectionIncubatesTwoAndPumpsYourPhyrexians(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phy := effectstest.PushDiesCreatureForTest(g, me.ID, "Phyrexian Test", "", "Creature — Phyrexian", 2, 2)
	theirs := effectstest.PushDiesCreatureForTest(g, opp.ID, "Their Phyrexian", "", "Creature — Phyrexian", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sculpted Perfection", "Enchantment", effectstest.SculptedPerfectionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("incubate counters = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, phy); got != 3 {
		t.Errorf("your Phyrexian power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their Phyrexian power = %d, want 2", got)
	}
	// A transformed Incubator is a Phyrexian Token and gets the bonus.
	inc := theIncubator(t, g, me.ID)
	transformIncubator(t, g, me, inc.InstanceID)
	if got := effectstest.EffectivePowerWithCounters(t, g, inc.InstanceID); got != 3 {
		t.Errorf("transformed Incubator power = %d, want 2 counters + 1", got)
	}
}

func TestSearingBarbDamagesStopsBlockingAndIncubatesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big Blocker", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Searing Barb", "Sorcery", effectstest.SearingBarbOracle, []game.TargetRef{effectstest.PbCard(big)})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Searing Barb", "Sorcery", effectstest.SearingBarbOracle, []game.TargetRef{effectstest.PbPlayer(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d -> %d, want -2", before, opp.Life)
	}
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 1 {
		t.Errorf("incubate counters = %d, want 1", got)
	}
}

func TestTangledSkylineGainsFiveIncubatesFiveAndGivesReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	phy := effectstest.PushDiesCreatureForTest(g, me.ID, "Phyrexian Test", "", "Creature — Phyrexian", 2, 2)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Tangled Skyline", "Enchantment", effectstest.TangledSkylineOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("life %d -> %d, want +5", before, me.Life)
	}
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 5 {
		t.Errorf("incubate counters = %d, want 5", got)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, phy), "reach") {
		t.Error("your Phyrexian lacks reach")
	}
}

func TestTillerOfFleshIncubatesOnlyWhenASpellTargetsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Tiller of Flesh", "Creature — Phyrexian Knight", effectstest.TillerOfFleshOracle, false)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 5, 5)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.IncubBLightningBoltOracle, []game.TargetRef{effectstest.PbPlayer(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.IncubatorsOf(g, me.ID)); n != 0 {
		t.Fatalf("a spell targeting a player made %d Incubators", n)
	}

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.IncubBLightningBoltOracle, []game.TargetRef{effectstest.PbCard(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("incubate counters = %d, want 2", got)
	}
}
