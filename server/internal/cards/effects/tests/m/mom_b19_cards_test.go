package testsm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mom_b19_cards_test.go — March of the Machine and its Commander set, S82
// batch 19 (#5477).

const (
	momB19ElspethOracle  = "0b2481d6-97a0-4f92-b7e5-3451863fd7fb"
	momB19ChandraOracle  = "53231b31-fb3c-4b30-bfe9-0ee8e754247b"
	momB19ThiefOracle    = "9495b2b4-21bf-488b-a64e-74592bb277d9"
	momB19OrthionOracle  = "31895957-5fa8-46dc-a76e-37bbdc6072ad"
	momB19MirrorOracle   = "6a871d70-a46d-4357-8c7d-876f7a004eb7"
	momB19DragsterOracle = "56ba8fe5-8693-4783-8406-9ced99de0b2a"
	momB19SwordOracle    = "1e0cf860-6be6-4b4b-9a67-0c19b6018ae1"
	momB19VrynOracle     = "73c7fdb3-4742-4d08-8658-daa4435c49b1"
)

func momB19Walker(t *testing.T, name, oracle string, loyalty int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, name, "Legendary Planeswalker", oracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	effectstest.AdvanceToMain(t, g)
	return g, me, id
}

// --- Archangel Elspeth ---

func TestArchangelElspethPlusOneMakesALifelinkSoldier(t *testing.T) {
	g, me, w := momB19Walker(t, "Archangel Elspeth", momB19ElspethOracle, 4)
	effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var soldier uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Soldier" {
			soldier = c.InstanceID
		}
	}
	if soldier == uuid.Nil {
		t.Fatal("no Soldier token")
	}
	effectstest.AssertKeywords(t, g, soldier, "lifelink")
	if got := effectstest.CountersOn(g, w, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestArchangelElspethMinusTwoCountersAngelAndFlying(t *testing.T) {
	g, me, w := momB19Walker(t, "Archangel Elspeth", momB19ElspethOracle, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d +1/+1 counters, want 2", got)
	}
	effectstest.AssertKeywords(t, g, bear, "flying")
	c := effectstest.BattlefieldCardPtr(t, g, bear)
	if !c.HasSubtype("Angel") || !c.HasSubtype("Bear") {
		t.Errorf("type line %q: want Angel in addition to Bear", c.TypeLine)
	}
	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err == nil {
		t.Error("the -2 targeted a player")
	}
}

func TestArchangelElspethMinusSixReturnsCheapNonlandPermanents(t *testing.T) {
	g, me, w := momB19Walker(t, "Archangel Elspeth", momB19ElspethOracle, 6)
	cheap := effectstest.AvB24Graveyard(me, game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	rock := effectstest.AvB24Graveyard(me, game.Card{Name: "Cheap Rock", TypeLine: "Artifact", ManaCost: "{3}"})
	pricey := effectstest.AvB24Graveyard(me, game.Card{Name: "Pricey Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{R}", Power: 5, Toughness: 5})
	land := effectstest.AvB24Graveyard(me, game.Card{Name: "Old Forest", TypeLine: "Land"})
	spell := effectstest.AvB24Graveyard(me, game.Card{Name: "Old Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	theirs := effectstest.AvB24Graveyard(g.Seats[1], game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 1, Toughness: 1})
	effectstest.B16Activate(t, g, me.ID, w, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{cheap, rock} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not return", id)
		}
	}
	for _, id := range []uuid.UUID{pricey, land, spell, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s returned but should not have", id)
		}
	}
}

// --- Orthion, Hero of Lavabrink ---

func TestOrthionCopiesAnotherCreatureWithHasteAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	orthion := effectstest.B12Push(g, me.ID, "Orthion, Hero of Lavabrink", "Legendary Creature — Human Soldier", momB19OrthionOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	// Not itself: the target must be another creature.
	if err := g.ActivateCatalogAbility(me.ID, orthion, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: orthion}}}); err == nil {
		t.Fatal("Orthion copied itself")
	}
	effectstest.B16Activate(t, g, me.ID, orthion, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	var copies []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Grizzly Bear" {
			copies = append(copies, c.InstanceID)
		}
	}
	if len(copies) != 1 {
		t.Fatalf("%d Bear copies, want 1", len(copies))
	}
	effectstest.AssertKeywords(t, g, copies[0], "haste")
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(copies[0]) {
		t.Error("the copy was not sacrificed at the end step")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the original left")
	}
}

func TestOrthionBigAbilityMakesFiveCopiesAndIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	orthion := effectstest.B12Push(g, me.ID, "Orthion, Hero of Lavabrink", "Legendary Creature — Human Soldier", momB19OrthionOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRRRRRRRR")
	effectstest.B16Activate(t, g, me.ID, orthion, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Grizzly Bear" {
			n++
			effectstest.AssertKeywords(t, g, c.InstanceID, "haste")
		}
	}
	if n != 5 {
		t.Fatalf("%d copies, want 5", n)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Grizzly Bear" {
			t.Fatal("a copy survived the end step")
		}
	}
}

func TestOrthionCannotActivateOutsideSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	orthion := effectstest.B12Push(g, me.ID, "Orthion, Hero of Lavabrink", "Legendary Creature — Human Soldier", momB19OrthionOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, orthion, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("activated in combat")
	}
}

// --- Schema Thief ---

func TestSchemaThiefCopiesAnArtifactOfTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	thief := effectstest.B12Push(g, me.ID, "Schema Thief", "Creature — Vedalken Rogue Artificer", momB19ThiefOracle, 3, 3)
	effectstest.AssertKeywords(t, g, thief, "flying")
	trinket := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Opp Trinket", TypeLine: "Artifact", ManaCost: "{2}"})
	elsewhere := effectstest.PushPermanent(g, other.ID, game.Card{Name: "Other Trinket", TypeLine: "Artifact", ManaCost: "{2}"})
	effectstest.AttackWith(t, g, opp.ID, thief)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: elsewhere}); err == nil {
			t.Error("an artifact of another player was accepted")
		}
		effectstest.AnswerPickTarget(t, g, trinket)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Opp Trinket" && c.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d token copies under my control, want 1", n)
	}
}

// --- Deluxe Dragster ---

func TestDeluxeDragsterCastsAnInstantFromTheDamagedPlayersGraveyardAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	drag := effectstest.B12Push(g, me.ID, "Deluxe Dragster", "Artifact Creature — Vehicle", momB19DragsterOracle, 4, 3)
	bolt := effectstest.GyCard(opp, "Opp Instant", "Instant", "{U}", "")
	bear := effectstest.GyCard(opp, "Opp Bear", "Creature — Bear", "{1}{G}", "")
	mine := effectstest.GyCard(me, "My Instant", "Instant", "{U}", "")
	effectstest.AttackWith(t, g, opp.ID, drag)
	ch := effectstest.PickTargetPrompt(t, g)
	if !slices.Contains(ch.PickTargetCards, bolt) || slices.Contains(ch.PickTargetCards, bear) || slices.Contains(ch.PickTargetCards, mine) {
		t.Fatalf("only the damaged player's instants and sorceries: %v", ch.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast it free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(bolt); z == nil || z.Kind != game.ZoneExile {
		t.Error("the spell was not exiled instead of going to a graveyard")
	}
}

func TestDeluxeDragsterCanOnlyBeBlockedByVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	drag := effectstest.B12Push(g, me.ID, "Deluxe Dragster", "Artifact Creature — Vehicle", momB19DragsterOracle, 4, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	cart := effectstest.B12Creature(g, opp.ID, "Cart", "Artifact Creature — Vehicle", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(drag, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, drag); err == nil {
		t.Error("a non-Vehicle blocked it")
	}
	if err := g.DeclareBlocker(cart, drag); err != nil {
		t.Errorf("a Vehicle could not block it: %v", err)
	}
}
