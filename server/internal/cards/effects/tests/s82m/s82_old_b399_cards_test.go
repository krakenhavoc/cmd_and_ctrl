package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b399_cards_test.go — Magic 2014 Auras, S82 older sets batch 399
// (#5563): Awaken the Ancient, Blessing, Indestructibility, Artificer's Hex,
// Shiv's Embrace and Illusionary Armor, through the real cast, activation and
// trigger paths.

const (
	m14b399Awaken  = "0fd27f59-36ec-4235-8e1f-4a4a14972f4a"
	m14b399Bless   = "5c84d8da-2bfb-4618-89a0-7d9ed604e854"
	m14b399Indes   = "9aaa0a28-e342-48bd-8f59-0cd289491f98"
	m14b399Hex     = "d14d9f44-9cff-4a11-be86-0cea5fe9b551"
	m14b399Shiv    = "e41162bf-1a4f-4e69-870a-faf9836c56d5"
	m14b399Illusry = "f36603d4-571d-4d03-a9cb-d1f2af3eb3c1"
)

func TestS82OldB399EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m14b399Awaken: "Awaken the Ancient", m14b399Bless: "Blessing", m14b399Indes: "Indestructibility",
		m14b399Hex: "Artificer's Hex", m14b399Shiv: "Shiv's Embrace", m14b399Illusry: "Illusionary Armor",
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

// aura399 seeds an Aura attached to host, the shape a resolved cast leaves.
func aura399(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	return effectstest.PushAuraOnLand(t, g, owner, name, oracle, host)
}

// --- Awaken the Ancient ---

func TestS82OldB399AwakenTheAncientMakesTheMountainASevenSevenHasteGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mountain := effectstest.B27Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Awaken the Ancient", "Enchantment — Aura", m14b399Awaken, effectstest.CardTarget(mountain)); err != nil {
		t.Fatalf("cast on a Mountain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, mountain)
	if !c.IsCreature() || !c.IsLand() {
		t.Errorf("type line %q, want a creature that is still a land", c.TypeLine)
	}
	if c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Errorf("%d/%d, want 7/7", c.CurrentPower(), c.CurrentToughness())
	}
	if !c.HasSubtype("Giant") {
		t.Error("not a Giant")
	}
	if cols := c.EffectiveColors(); len(cols) != 1 || cols[0] != "R" {
		t.Errorf("colors %v, want red", cols)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mountain, "haste") {
		t.Error("no haste")
	}
}

func TestS82OldB399AwakenTheAncientEnchantsOnlyAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.B27Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, host := range []uuid.UUID{forest, bear} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Awaken the Ancient", "Enchantment — Aura", m14b399Awaken, effectstest.CardTarget(host)); err == nil {
			t.Errorf("cast on %v, which is not a Mountain", host)
		}
	}
}

func TestS82OldB399AwakenTheAncientEndsWhenTheAuraLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mountain := effectstest.B27Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	aura := aura399(t, g, me.ID, "Awaken the Ancient", m14b399Awaken, mountain)
	if !effectstest.LayeredCard(t, g, mountain).IsCreature() {
		t.Fatal("the Mountain is not a creature while enchanted")
	}
	effectstest.B27Kill(g, aura)
	if effectstest.LayeredCard(t, g, mountain).IsCreature() {
		t.Error("the Mountain is still a creature without the Aura")
	}
}

// --- Blessing ---

func TestS82OldB399BlessingPumpsTheEnchantedCreatureOncePerWhiteMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := aura399(t, g, me.ID, "Blessing", m14b399Bless, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "WW")
	for i := 0; i < 2; i++ {
		effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d after two activations, want 4", p)
	}
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentToughness() != 4 {
		t.Errorf("toughness %d, want 4", c.CurrentToughness())
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no white mana left")
	}
}

// --- Indestructibility ---

func TestS82OldB399IndestructibilityProtectsAnyEnchantedPermanentFromDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B27Push(g, me.ID, "Rock", "Artifact", "", "{2}", 0, 0)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	aura399(t, g, me.ID, "Indestructibility", m14b399Indes, bear)
	aura399(t, g, me.ID, "Indestructibility", m14b399Indes, rock)
	for _, id := range []uuid.UUID{bear, rock, bare} {
		effectstest.B27Kill(g, id)
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("an enchanted permanent was destroyed")
	}
	if g.Battlefield.Contains(bare) {
		t.Error("the unenchanted creature survived")
	}
}

func TestS82OldB399IndestructibilityCanEnchantAnyPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B27Push(g, me.ID, "Rock", "Artifact", "", "{2}", 0, 0)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Indestructibility", "Enchantment — Aura", m14b399Indes, effectstest.CardTarget(rock)); err != nil {
		t.Errorf("cast on an artifact: %v", err)
	}
}

// --- Artificer's Hex ---

func TestS82OldB399ArtificersHexDestroysTheCreatureTheEnchantedEquipmentIsAttachedTo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sword := effectstest.B27Push(g, opp.ID, "Sword", "Artifact — Equipment", "", "{1}", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.AttachForTest(t, g, sword, bear)
	aura399(t, g, me.ID, "Artificer's Hex", m14b399Hex, sword)
	effectstest.AdvanceToMain(t, g)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the equipped creature survived")
	}
	if !g.Battlefield.Contains(bystander) || !g.Battlefield.Contains(sword) {
		t.Error("something else was destroyed")
	}
}

func TestS82OldB399ArtificersHexDoesNothingForAnUnattachedEquipmentOrOtherUpkeeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loose := effectstest.B27Push(g, opp.ID, "Loose Sword", "Artifact — Equipment", "", "{1}", 0, 0)
	aura399(t, g, me.ID, "Artificer's Hex", m14b399Hex, loose)
	sword := effectstest.B27Push(g, opp.ID, "Sword", "Artifact — Equipment", "", "{1}", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttachForTest(t, g, sword, bear)
	aura399(t, g, me.ID, "Artificer's Hex", m14b399Hex, sword)
	effectstest.AdvanceToMain(t, g)
	effectstest.BeginUpkeep(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was destroyed on someone else's upkeep")
	}
}

func TestS82OldB399ArtificersHexEnchantsOnlyEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B27Push(g, me.ID, "Rock", "Artifact", "", "{2}", 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, host := range []uuid.UUID{rock, bear} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Artificer's Hex", "Enchantment — Aura", m14b399Hex, effectstest.CardTarget(host)); err == nil {
			t.Errorf("cast on %v, which is not an Equipment", host)
		}
	}
}

// --- Shiv's Embrace ---

func TestS82OldB399ShivsEmbraceGivesPlusTwoAndFlyingAndFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := aura399(t, g, me.ID, "Shiv's Embrace", m14b399Shiv, bear)
	effectstest.AdvanceToMain(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("no flying")
	}
	effectstest.FloatForTest(g, me, "RR")
	for i := 0; i < 2; i++ {
		effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	c := effectstest.LayeredCard(t, g, bear)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 4 {
		t.Errorf("%d/%d after two firebreaths, want 6/4", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Illusionary Armor ---

func TestS82OldB399IllusionaryArmorGivesPlusFourAndIsSacrificedWhenTheCreatureIsTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := aura399(t, g, me.ID, "Illusionary Armor", m14b399Illusry, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 6 {
		t.Errorf("power %d, want 6", p)
	}
	effectstest.CastCatalogSpell(t, g, "Targeting Spell", "Instant", "", effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("power %d after the Aura left, want 2", p)
	}
}

func TestS82OldB399IllusionaryArmorStaysWhenAnotherCreatureIsTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	aura := aura399(t, g, me.ID, "Illusionary Armor", m14b399Illusry, bear)
	effectstest.CastCatalogSpell(t, g, "Targeting Spell", "Instant", "", effectstest.CardTarget(other))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Error("the Aura was sacrificed for a spell that targeted another creature")
	}
}
