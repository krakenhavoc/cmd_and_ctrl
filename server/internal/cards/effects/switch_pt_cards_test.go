package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// switch_pt_cards_test.go — #3996's "switch <creature>'s power and
// toughness until end of turn" (CR 613.4d, layer 7d) on the cards that
// print it. The engine half's layer cases are in
// game/switch_pt_test.go.

const (
	swAboutFaceOracle       = "8a810017-06ac-4824-bc2f-a7d303d58ad7"
	swAeromoebaOracle       = "1e276998-9c54-4d68-a901-657853312d06"
	swAquamoebaOracle       = "55de434c-0fb9-4127-81b6-47c1c0b088fe"
	swCalciteSnapperOracle  = "36593200-e53c-46e4-a04e-8deec8273c65"
	swCragPucaOracle        = "b51d2c70-204b-4321-a589-f00a3e49fab4"
	swCrookclawOracle       = "53501112-bcdb-4bb5-9fd3-674cdba91951"
	swDwarvenThaumOracle    = "019d0013-52fd-4261-9cc8-7efcee5a4616"
	swFlatmanOracle         = "3a7e3423-4df3-4e10-9861-ac2f403bc49d"
	swFluxchargerOracle     = "99bba323-b5a4-4ff3-8ad1-2a19ed119069"
	swInsideOutOracle       = "623d9db1-a3af-4a1c-9643-975de364e073"
	swInversionOracle       = "fedcb737-723b-4a38-b86c-dad6ac7bec7b"
	swInvertInventOracle    = "9a378964-2c04-4d60-a905-c819d37ed4c3"
	swMannichiOracle        = "3aafbf63-35c0-4f62-9281-a4d460e8d6f3"
	swMerfolkThaumOracle    = "1220e11f-aab0-459b-ae3f-b095aad37294"
	swMyrQuadropodOracle    = "f53732d3-d317-4042-a0ad-1a99c7ffc94f"
	swPhantasmalFiendOracle = "821ce349-5379-4415-89e2-224e21dda24b"
	swReversePolarityOracle = "4499d8cd-d090-4acc-a1d9-3d9d9b6ef768"
	swTransmutationOracle   = "72a75113-4ddf-460d-8d9b-03931c2788da"
	swTurtleshellOracle     = "87f8b403-765c-4a2a-8b21-edaaaa889234"
	swTwistedImageOracle    = "4a40cee1-fc09-4f97-aef8-440778d8c0ca"
	swTwistedReflOracle     = "57fdf496-c31e-4a3b-a8a6-4f053b808267"
	swValakutFireboarOracle = "126bf98a-64d0-4de2-9bf4-f59d36ef6e25"
	swWanderingFumOracle    = "741c51f1-cfbe-4c29-ac8f-ca6bcd2652f9"
	swWindreaverOracle      = "d6b6c27e-8028-49b1-90c6-ae0907159f85"
)

func TestSwitchPTCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		swAboutFaceOracle:           "About Face",
		swAeromoebaOracle:           "Aeromoeba",
		swAquamoebaOracle:           "Aquamoeba",
		swCalciteSnapperOracle:      "Calcite Snapper",
		swCragPucaOracle:            "Crag Puca",
		swCrookclawOracle:           "Crookclaw Transmuter",
		swDwarvenThaumOracle:        "Dwarven Thaumaturgist",
		swFlatmanOracle:             "Flatman",
		swFluxchargerOracle:         "Fluxcharger",
		swInsideOutOracle:           "Inside Out",
		swInversionOracle:           "Inversion Behemoth",
		swInvertInventOracle:        "Invert",
		swInvertInventOracle + "#1": "Invent",
		swMannichiOracle:            "Mannichi, the Fevered Dream",
		swMerfolkThaumOracle:        "Merfolk Thaumaturgist",
		swMyrQuadropodOracle:        "Myr Quadropod",
		swPhantasmalFiendOracle:     "Phantasmal Fiend",
		swReversePolarityOracle:     "Reverse the Polarity",
		swTransmutationOracle:       "Transmutation",
		swTurtleshellOracle:         "Turtleshell Changeling",
		swTwistedImageOracle:        "Twisted Image",
		swTwistedReflOracle:         "Twisted Reflection",
		swValakutFireboarOracle:     "Valakut Fireboar",
		swWanderingFumOracle:        "Wandering Fumarole",
		swWindreaverOracle:          "Windreaver",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: Lookup = %q, %v, completeness %v", name, spec.Name, ok, spec.Completeness)
		}
	}
}

// swPush parks a catalog creature with its printed power and toughness,
// not summoning sick, announcing the entry so its abilities register.
func swPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

// swPT is the creature's power and toughness, counters included.
func swPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := gaLayered(t, g, id)
	p, _, tough := c.PowerToughness()
	return p, tough
}

func swWant(t *testing.T, g *game.Game, id uuid.UUID, power, toughness int, when string) {
	t.Helper()
	if p, tough := swPT(t, g, id); p != power || tough != toughness {
		t.Errorf("%s: P/T = %d/%d, want %d/%d", when, p, tough, power, toughness)
	}
}

func swOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

func swCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// swActivate activates row idx of source with `mana` in the pool and
// settles the stack.
func swActivate(t *testing.T, g *game.Game, who *game.Player, source uuid.UUID, idx int, mana string, targets []game.TargetRef) {
	t.Helper()
	who.ManaPool.EmptyPool()
	if mana != "" {
		if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("add mana: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(who.ID, source, idx, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("activate row %d: %v", idx, err)
	}
	passPriorityAroundTable(t, g)
}

// --- Twisted Reflection (the issue's card) ----------------------------

// Entwined on one 6/6: the -6/-0 is layer 7c and the switch 7d, so the
// creature is 0/6 switched to 6/0 whatever order the bullets resolve
// in, and it dies.
func TestTwistedReflectionEntwinedOnOneCreatureKillsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	giant := pushVanillaCreature(g, opp.ID, "Giant", 6, 6)
	if _, err := castEntwineSpell(g, me, "Twisted Reflection", "Instant", "{1}{U}", swTwistedReflOracle, []string{"U"},
		"{C}{U}{B}", []int{0, 1}, true, []game.TargetRef{ewaTarget(giant, 0), ewaTarget(giant, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if swOnBattlefield(g, giant) {
		p, tough := swPT(t, g, giant)
		t.Fatalf("a 6/6 given -6/-0 and switched is 6/0 and dies; it is %d/%d", p, tough)
	}
}

func TestTwistedReflectionSwitchBulletAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 5)
	if _, err := castEntwineSpell(g, me, "Twisted Reflection", "Instant", "{1}{U}", swTwistedReflOracle, []string{"U"},
		"{C}{U}", []int{1}, false, []game.TargetRef{ewaTarget(wall, 0)}); err != nil {
		t.Fatalf("switch bullet: %v", err)
	}
	passPriorityAroundTable(t, g)
	if swOnBattlefield(g, wall) {
		t.Fatal("a 0/5 switched is 5/0 and dies")
	}
}

func TestTwistedReflectionShrinkBulletAloneDoesNotSwitch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	giant := pushVanillaCreature(g, opp.ID, "Giant", 7, 7)
	if _, err := castEntwineSpell(g, me, "Twisted Reflection", "Instant", "{1}{U}", swTwistedReflOracle, []string{"U"},
		"{C}{U}", []int{0}, false, []game.TargetRef{ewaTarget(giant, 0)}); err != nil {
		t.Fatalf("shrink bullet: %v", err)
	}
	passPriorityAroundTable(t, g)
	swWant(t, g, giant, 1, 7, "-6/-0 alone")
}

// --- the spells -----------------------------------------------------

// A +1/+1 counter is applied before the switch (CR 613.4c, 613.4d), and
// the switch ends at cleanup.
func TestAboutFaceSwitchesCountersTooAndEndsAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 1, 4)
	addCountersForTest(g, bear, game.CounterPlusOne, 1)
	castCatalogSpell(t, g, "About Face", "Instant", swAboutFaceOracle, swCard(bear))
	passPriorityAroundTable(t, g)
	swWant(t, g, bear, 5, 2, "switched 2/5")
	// A counter put on after the switch still lands before it.
	addCountersForTest(g, bear, game.CounterPlusOne, 1)
	swWant(t, g, bear, 6, 3, "switched 3/6")
	advanceToNextSeatsTurn(t, g)
	swWant(t, g, bear, 3, 6, "next turn")
}

func TestTwistedImageSwitchesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Twisted Image", "Instant", swTwistedImageOracle, swCard(wall))
	passPriorityAroundTable(t, g)
	if swOnBattlefield(g, wall) {
		t.Error("a 0/4 switched is 4/0 and dies")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestInvertSwitchesUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	a := pushVanillaCreature(g, me.ID, "A", 1, 3)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 5)
	c := pushVanillaCreature(g, opp.ID, "C", 1, 2)
	castCatalogSpell(t, g, "Invert", "Instant", swInvertInventOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	swWant(t, g, a, 3, 1, "A")
	swWant(t, g, b, 5, 2, "B")
	swWant(t, g, c, 1, 2, "C, not targeted")
}

// Mannichi's and Reverse the Polarity's "each creature" is the set on
// the battlefield as it resolves (CR 611.2c).
func TestMannichiSwitchesEachCreatureOnceLocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mannichi := swPush(g, me.ID, "Mannichi, the Fevered Dream", "Legendary Creature — Spirit", swMannichiOracle, 1, 2)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 3)
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 3)
	swActivate(t, g, me, mannichi, 0, "{C}{R}", nil)
	swWant(t, g, mannichi, 2, 1, "Mannichi")
	swWant(t, g, bear, 3, 2, "Bear")
	if swOnBattlefield(g, wall) {
		t.Error("a 0/3 switched is 3/0 and dies")
	}
	later := pushVanillaCreature(g, opp.ID, "Later", 1, 4)
	swWant(t, g, later, 1, 4, "a creature that arrived later")
}

func TestReverseThePolaritySwitchMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	a := pushVanillaCreature(g, me.ID, "A", 1, 3)
	b := pushVanillaCreature(g, opp.ID, "B", 4, 6)
	castCatalogSpellWithModes(t, g, "Reverse the Polarity", "Instant", swReversePolarityOracle, []int{1})
	passPriorityAroundTable(t, g)
	swWant(t, g, a, 3, 1, "A")
	swWant(t, g, b, 6, 4, "B")
}

// "Counter all other spells" counters what is under it and resolves.
func TestReverseThePolarityCountersAllOtherSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	image := castCatalogSpell(t, g, "Twisted Image", "Instant", swTwistedImageOracle, swCard(wall))
	polarity := castCatalogSpellWithModes(t, g, "Reverse the Polarity", "Instant", swReversePolarityOracle, []int{0})
	passPriorityAroundTable(t, g)
	if !swOnBattlefield(g, wall) {
		t.Error("Twisted Image was countered: the wall is not switched")
	}
	if !me.Graveyard.Contains(image) || !me.Graveyard.Contains(polarity) {
		t.Error("both spells end in their owner's graveyard")
	}
}

// --- the activated abilities ----------------------------------------

// Two switches undo each other.
func TestCragPucaSwitchesAndSwitchesBack(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	puca := swPush(g, me.ID, "Crag Puca", "Creature — Shapeshifter", swCragPucaOracle, 2, 4)
	swActivate(t, g, me, puca, 0, "{U}", nil)
	swWant(t, g, puca, 4, 2, "once")
	swActivate(t, g, me, puca, 0, "{R}", nil)
	swWant(t, g, puca, 2, 4, "twice")
}

// A pump activated after the switch still applies before it (7c before
// 7d): a 1/5 switched and then given +1/-1 is 2/4 switched, 4/2.
func TestPhantasmalFiendPumpAfterTheSwitchAppliesFirst(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	fiend := swPush(g, me.ID, "Phantasmal Fiend", "Creature — Illusion", swPhantasmalFiendOracle, 1, 5)
	swActivate(t, g, me, fiend, 1, "{C}{U}", nil)
	swWant(t, g, fiend, 5, 1, "switched")
	swActivate(t, g, me, fiend, 0, "{B}", nil)
	swWant(t, g, fiend, 4, 2, "switched, then +1/-1")
}

func TestMerfolkThaumaturgistSwitchesItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	thaum := swPush(g, me.ID, "Merfolk Thaumaturgist", "Creature — Merfolk Wizard", swMerfolkThaumOracle, 1, 2)
	bear := pushVanillaCreature(g, opp.ID, "Bear", 3, 1)
	swActivate(t, g, me, thaum, 0, "", swCard(bear))
	swWant(t, g, bear, 1, 3, "Bear")
	swWant(t, g, thaum, 1, 2, "the Thaumaturgist itself")
}

func TestAquamoebaDiscardsToSwitch(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	moeba := swPush(g, me.ID, "Aquamoeba", "Creature — Elemental Beast", swAquamoebaOracle, 1, 3)
	if me.Hand.Size() == 0 {
		t.Fatal("the harness hand is empty")
	}
	hand := me.Hand.Size()
	discard := me.Hand.Cards[0].InstanceID
	if err := g.ActivateCatalogAbility(me.ID, moeba, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand-1)
	}
	swWant(t, g, moeba, 3, 1, "switched")
}

// The land becomes a 1/4 with "{0}: Switch this creature's power and
// toughness": the 7b base comes before the switch, so a +1/+1 counter
// makes it 2/5 switched to 5/2.
func TestWanderingFumaroleAnimatesAndSwitches(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	land := swPush(g, me.ID, "Wandering Fumarole", "Land", swWanderingFumOracle, 0, 0)
	swActivate(t, g, me, land, 0, "{C}{C}{U}{R}", nil)
	swWant(t, g, land, 1, 4, "animated")
	addCountersForTest(g, land, game.CounterPlusOne, 1)
	idx := -1
	var ref string
	abs, origins := game.ActivatedAbilitiesWithOrigins(gaLayered(t, g, land))
	for i, ab := range abs {
		if origins.At(i).Granted() && strings.HasPrefix(ab.Label, "{0}: Switch") {
			idx, ref = i, origins.Ref(i)
		}
	}
	if idx < 0 {
		t.Fatal("the animated land has no granted switch row")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Strict: true, Ref: ref}); err != nil {
		t.Fatalf("activate the switch: %v", err)
	}
	passPriorityAroundTable(t, g)
	swWant(t, g, land, 5, 2, "switched")
}

// --- the triggers ---------------------------------------------------

func TestValakutFireboarSwitchesWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	boar := swPush(g, me.ID, "Valakut Fireboar", "Creature — Elemental Boar", swValakutFireboarOracle, 1, 7)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(boar, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)
	swWant(t, g, boar, 7, 1, "attacking")
}

func TestCrookclawTransmuterSwitchesOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 2)
	id := handCardFull(me, "Crookclaw Transmuter", "Creature — Bird Wizard", "{3}{U}", swCrookclawOracle, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, wall)
	passPriorityAroundTable(t, g)
	if swOnBattlefield(g, wall) {
		t.Error("a 0/2 switched is 2/0 and dies")
	}
}
