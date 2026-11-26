package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b306_cards_test.go — Battle for Zendikar Devoid cards, batch 306
// (#5547). Each test plays the real catalog entry and asserts what its text
// says.

const (
	oldB306Aggregate  = "4a6b254e-76fb-4712-ac34-5e26c311b2e4"
	oldB306Sifter     = "4c2c5f3b-61f3-4772-9ade-f5be80719e9f"
	oldB306Stalker    = "69cd25d2-206c-4572-95e7-d66973c94358"
	oldB306Oracle     = "6e5f50b2-0acc-42be-b9fd-32ea390c07fc"
	oldB306CallScions = "731a70e8-40d6-47b9-b2b3-b16c818e1935"
	oldB306Surge      = "77b98632-e103-44fa-a940-2c3b2897b60f"
	oldB306Adverse    = "7a5a81f7-1597-4972-8565-fd2a5667df99"
	oldB306Sentinel   = "86ae05d6-f1ba-456d-bbc0-f35cd0d0537d"
	oldB306Skimmer    = "92198967-e89b-4213-a0fa-71361c8420dc"
	oldB306Nursery    = "92abb79d-3c60-48f7-87a0-13c25a5e1128"
	oldB306Transgres  = "9a9bc316-9652-46d9-9448-57b8902f5d6f"
)

func TestOldB306CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB306Aggregate: "Vile Aggregate", oldB306Sifter: "Catacomb Sifter", oldB306Stalker: "Dust Stalker",
		oldB306Oracle: "Oracle of Dust", oldB306CallScions: "Call the Scions", oldB306Surge: "Swarm Surge",
		oldB306Adverse: "Adverse Conditions", oldB306Sentinel: "Kozilek's Sentinel", oldB306Skimmer: "Silent Skimmer",
		oldB306Nursery: "Molten Nursery", oldB306Transgres: "Transgress the Mind",
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

func oldB306Colored(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: "{G}", Colors: []string{"G"},
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func oldB306Scions(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsToken() && c.HasSubtype("Scion") {
			n++
		}
	}
	return n
}

func TestOldB306VileAggregatePowerCountsColorlessCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	agg := effectstest.B12Push(g, me.ID, "Vile Aggregate", "Creature — Eldrazi Drone", oldB306Aggregate, 0, 5)
	if p := effectstest.EffectivePower(t, g, agg); p != 1 {
		t.Errorf("alone: power %d, want 1 (itself)", p)
	}
	effectstest.B12Creature(g, me.ID, "Grey Golem", "Artifact Creature — Golem", 2, 2)
	oldB306Colored(g, me.ID, "Green Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	if p := effectstest.EffectivePower(t, g, agg); p != 2 {
		t.Errorf("with a colorless Golem, a green Bear and an opposing Golem: power %d, want 2", p)
	}
	if tough := effectstest.EffectiveToughness(t, g, agg); tough != 5 {
		t.Errorf("toughness %d, want 5", tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, agg, "trample") {
		t.Error("no trample")
	}
}

func TestOldB306IngestExilesTheTopCardOfTheDamagedPlayersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	agg := effectstest.B12Push(g, me.ID, "Vile Aggregate", "Creature — Eldrazi Drone", oldB306Aggregate, 0, 5)
	effectstest.SeedLibrary(opp, "a", "b", "c")
	lib := opp.Library.Size()
	exile := g.Exile.Size()
	effectstest.DeclareAttack(t, g, opp.ID, agg)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-1 || g.Exile.Size() != exile+1 {
		t.Errorf("library %d -> %d, exile %d -> %d, want one card exiled", lib, opp.Library.Size(), exile, g.Exile.Size())
	}
}

func TestOldB306CatacombSifterMakesAScionAndScriesWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := oldB306Scions(g, me.ID)
	effectstest.CastAndResolveCreature(t, g, "Catacomb Sifter", "Creature — Eldrazi Drone", oldB306Sifter)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB306Scions(g, me.ID) != before+1 {
		t.Fatalf("%d Scions, want %d", oldB306Scions(g, me.ID), before+1)
	}
	effectstest.SeedLibrary(me, "x", "y", "z")
	scion := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Scion") {
			scion = c.InstanceID
		}
	}
	effectstest.KillCreature(t, g, me.ID, scion)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("another creature of mine died but there was no scry prompt")
	}
}

func TestOldB306DustStalkerReturnsOnlyWithNoOtherColorlessCreature(t *testing.T) {
	for _, other := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		stalker := effectstest.B12Push(g, me.ID, "Dust Stalker", "Creature — Eldrazi", oldB306Stalker, 5, 3)
		if other {
			effectstest.B12Creature(g, me.ID, "Grey Golem", "Artifact Creature — Golem", 2, 2)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		inHand := me.Hand.Contains(stalker)
		if other && (inHand || !g.Battlefield.Contains(stalker)) {
			t.Error("returned to hand although another colorless creature is out")
		}
		if !other && (!inHand || g.Battlefield.Contains(stalker)) {
			t.Error("did not return to hand at the end step with no other colorless creature")
		}
	}
}

func TestOldB306OracleOfDustLootsByProcessingACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oracle := effectstest.B12Push(g, me.ID, "Oracle of Dust", "Creature — Eldrazi Processor", oldB306Oracle, 3, 5)
	effectstest.SeedLibrary(me, "x", "y")
	junk := effectstest.B13HandCard(me, "Junk", "Sorcery", "")
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no card an opponent owns in exile")
	}
	exiled := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: exiled, Name: "Exiled Card", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{exiled}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !opp.Graveyard.Contains(exiled) {
		t.Error("the exiled card did not go to its owner's graveyard as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		effectstest.AnswerDiscard(t, g, me.ID, junk)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want a draw and a discard (no change)", hand, me.Hand.Size())
	}
}

func TestOldB306CallTheScionsMakesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Call the Scions", "Sorcery", oldB306CallScions, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB306Scions(g, me.ID) != 2 {
		t.Errorf("%d Scions, want 2", oldB306Scions(g, me.ID))
	}
}

func TestOldB306SwarmSurgePumpsAllAndGivesColorlessFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.B12Creature(g, me.ID, "Grey Golem", "Artifact Creature — Golem", 2, 2)
	green := oldB306Colored(g, me.ID, "Green Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Swarm Surge", "Sorcery", oldB306Surge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, golem) != 4 || effectstest.EffectivePower(t, g, green) != 4 {
		t.Errorf("my creatures: %d and %d power, want 4", effectstest.EffectivePower(t, g, golem), effectstest.EffectivePower(t, g, green))
	}
	if effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("an opposing creature was pumped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, golem, "first strike") {
		t.Error("the colorless creature has no first strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, green, "first strike") {
		t.Error("the green creature gained first strike")
	}
}

func TestOldB306AdverseConditionsTapsFreezesAndMakesAScion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Adverse Conditions", "Instant", oldB306Adverse, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, a) || !effectstest.IsTapped(g, b) {
		t.Error("a target was not tapped")
	}
	if oldB306Scions(g, me.ID) != 1 {
		t.Errorf("%d Scions, want 1", oldB306Scions(g, me.ID))
	}
	// their next untap step: both stay tapped
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.IsTapped(g, a) || !effectstest.IsTapped(g, b) {
		t.Error("a frozen creature untapped during its controller's next untap step")
	}
}

func TestOldB306KozileksSentinelGrowsOnlyForColorlessSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s := effectstest.B12Push(g, me.ID, "Kozilek's Sentinel", "Creature — Eldrazi Drone", oldB306Sentinel, 1, 4)
	effectstest.PutInHand(me, game.Card{Name: "Green Spell", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Grey Spell", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, s); p != 2 {
		t.Errorf("after a colorless spell: power %d, want 2", p)
	}
}

func TestOldB306SilentSkimmerDrainsTheDefendingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	skim := effectstest.B12Push(g, me.ID, "Silent Skimmer", "Creature — Eldrazi Drone", oldB306Skimmer, 0, 4)
	life, other := opp.Life, g.Seats[2].Life
	effectstest.DeclareAttack(t, g, opp.ID, skim)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 || g.Seats[2].Life != other {
		t.Errorf("defender life %d -> %d, bystander %d -> %d, want -2 and no change", life, opp.Life, other, g.Seats[2].Life)
	}
}

func TestOldB306MoltenNurseryPingsOnAColorlessSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Molten Nursery", "Enchantment", oldB306Nursery, 0, 0)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Grey Spell", "Artifact", "", nil)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 damage", life, opp.Life)
	}
}

func TestOldB306TransgressTheMindExilesAThreeOrMoreCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := uuid.New()
	big := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: cheap, Name: "Cheap", TypeLine: "Sorcery", ManaCost: "{1}{B}", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Bear", ManaCost: "{3}{B}", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Transgress the Mind", "Sorcery", oldB306Transgres, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(prompt.ID, me.ID, []uuid.UUID{cheap}); err == nil {
		t.Error("the mana value 2 card was accepted")
	}
	if err := g.ResolvePendingChoice(prompt.ID, me.ID, []uuid.UUID{big}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(big) || opp.Hand.Contains(big) {
		t.Error("the card was not exiled from their hand")
	}
	if opp.Graveyard.Contains(big) {
		t.Error("the card was discarded rather than exiled")
	}
}
