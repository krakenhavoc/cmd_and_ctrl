package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_dl_cards_test.go — The Lost Caverns of Ixalan, cards D–L, second
// batch (#4470). Each test plays the real catalog entry and asserts what
// its text says.

const (
	lciDLIronpawOracle     = "fab5f694-4a9c-4e72-b164-f236e8212ff7"
	lciDLDawnrunnerOracle  = "b6565da9-773c-49b1-85f5-e34f07ba84cb"
	lciDLNautilusOracle    = "6bda307b-1508-4582-952b-c690db85e045"
	lciDLHotfootOracle     = "3efb64a4-04f6-4b78-be1c-5b6320c02a71"
	lciDLPilgrimOracle     = "64399663-2c1e-447e-a3ed-c49ebd9f2231"
	lciDLStriderOracle     = "7534d0b3-4211-4add-8e8d-370c033518d3"
	lciDLLeechOracle       = "073f12d1-eaa1-4d08-bc1d-b216a3dc0c67"
	lciDLJoinOracle        = "b153fc70-0312-4856-befd-bb6b9a04a26e"
	lciDLFinalStrikeOracle = "20acb8df-2ea0-48b0-8dcf-54dd0e1f8db8"
	lciDLBlowgunOracle     = "a0292e0b-7bb3-46d3-a29f-b54bce5b20f0"
	lciDLPickAxeOracle     = "542e7d60-3fc8-46e2-8786-f7c8b3d4c979"
	lciDLFungalOracle      = "cf73d2e7-584f-41d0-8fc7-50d886dcf9fe"
	lciDLReliquaryOracle   = "dca7102e-c020-4fdc-9b59-f186e3e615a5"
	lciDLFlankerOracle     = "ba557e19-a409-4c02-a9ac-a8fb77762e43"
	lciDLItzquinthOracle   = "f976a244-1240-4430-9182-bb8d6d932ea0"
)

func TestLciDLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciDLIronpawOracle:     "Ironpaw Aspirant",
		lciDLDawnrunnerOracle:  "Kinjalli's Dawnrunner",
		lciDLNautilusOracle:    "Hermitic Nautilus",
		lciDLHotfootOracle:     "Hotfoot Gnome",
		lciDLPilgrimOracle:     "Hoverstone Pilgrim",
		lciDLStriderOracle:     "Goldfury Strider",
		lciDLLeechOracle:       "Gargantuan Leech",
		lciDLJoinOracle:        "Join the Dead",
		lciDLFinalStrikeOracle: "Huatli's Final Strike",
		lciDLBlowgunOracle:     "Hunter's Blowgun",
		lciDLPickAxeOracle:     "Diamond Pick-Axe",
		lciDLFungalOracle:      "Fungal Fortitude",
		lciDLReliquaryOracle:   "Dusk Rose Reliquary",
		lciDLFlankerOracle:     "Kutzil's Flanker",
		lciDLItzquinthOracle:   "Itzquinth, Firstborn of Gishath",
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

// --- fixtures ------------------------------------------------------------------

// lciDLFlankerModes casts Kutzil's Flanker, resolves the spell and picks
// the enters trigger's bullet.
func lciDLFlankerModes(t *testing.T, g *game.Game, me *game.Player, mode int) uuid.UUID {
	t.Helper()
	flanker := effectstest.CastAndResolveCreature(t, g, "Kutzil's Flanker", "Creature — Cat Warrior", lciDLFlankerOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode pick from the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
	return flanker
}

// --- Ironpaw Aspirant ------------------------------------------------------------

func TestLciDLIronpawAspirantPutsACounterOnTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aspirant := effectstest.CastHoldCreature(t, g, "Ironpaw Aspirant", "Creature — Cat Warrior", lciDLIronpawOracle, 1, 2)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no enters-trigger target prompt")
	}
	if !effectstest.HasID(pick.PickTargetCards, theirs) || !effectstest.HasID(pick.PickTargetCards, aspirant) {
		t.Errorf("offered %v; want any creature", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the target, want 1", got)
	}
	if got := effectstest.CountersOn(g, aspirant, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters on the Aspirant, want 0", got)
	}
}

// --- Kinjalli's Dawnrunner -------------------------------------------------------

func TestLciDLKinjallisDawnrunnerExploresOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	top := effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
	runner := effectstest.Msc17Play(t, g, me, game.Card{Name: "Kinjalli's Dawnrunner", TypeLine: "Creature — Human Scout", OracleID: lciDLDawnrunnerOracle, Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, runner, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after revealing a nonland, want 1", got)
	}
	effectstest.AnswerExploreQuestion(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the revealed card was not put into the graveyard")
	}
	if !effectstest.HasEffectiveKeyword(t, g, runner, "double strike") {
		t.Error("Kinjalli's Dawnrunner has no double strike")
	}
}

// --- Hermitic Nautilus -----------------------------------------------------------

func TestLciDLHermiticNautilusTradesToughnessForPower(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	naut := effectstest.B12Push(g, me.ID, "Hermitic Nautilus", "Artifact Creature — Nautilus", lciDLNautilusOracle, 1, 4)
	if !effectstest.HasEffectiveKeyword(t, g, naut, "vigilance") {
		t.Error("Hermitic Nautilus has no vigilance")
	}
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, naut, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the {U}")
	}
	effectstest.ApaMana(me, "U")
	effectstest.P7Activate(t, g, me, naut, 0, game.ActivateAbilityParams{Strict: true})
	if p, tough := effectstest.EffectivePTOf(t, g, naut); p != 4 || tough != 1 {
		t.Errorf("after one activation the Nautilus is %d/%d, want 4/1", p, tough)
	}
	effectstest.ApaMana(me, "C", "U")
	effectstest.P7Activate(t, g, me, naut, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(naut) {
		t.Error("a 7/-2 Nautilus is still on the battlefield")
	}
}

// --- Hotfoot Gnome ---------------------------------------------------------------

func TestLciDLHotfootGnomeGivesAnotherCreatureHaste(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	gnome := effectstest.B12Push(g, me.ID, "Hotfoot Gnome", "Artifact Creature — Gnome", lciDLHotfootOracle, 3, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, gnome, "haste") {
		t.Error("Hotfoot Gnome has no haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, gnome, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(gnome)}); err == nil {
		t.Fatal("the Gnome targeted itself; the text says another")
	}
	effectstest.P7Activate(t, g, me, gnome, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(bear)})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the targeted creature did not gain haste")
	}
	if !effectstest.TappedNow(g, gnome) {
		t.Error("the Gnome did not tap for its ability")
	}
}

// --- Hoverstone Pilgrim ----------------------------------------------------------

func TestLciDLHoverstonePilgrimTucksAGraveyardCardUnderItsOwnersLibrary(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	pilgrim := effectstest.B12Push(g, me.ID, "Hoverstone Pilgrim", "Artifact Creature — Golem", lciDLPilgrimOracle, 2, 5)
	if !effectstest.HasEffectiveKeyword(t, g, pilgrim, "flying") {
		t.Error("Hoverstone Pilgrim has no flying")
	}
	dead := effectstest.PushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
	effectstest.PutOnTopForTest(opp, "Their Top Card", "Sorcery")
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, pilgrim, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(dead), Strict: true}); err == nil {
		t.Fatal("activated with one mana for {2}")
	}
	effectstest.ApaMana(me, "C")
	effectstest.P7Activate(t, g, me, pilgrim, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(dead), Strict: true})
	if opp.Graveyard.Contains(dead) {
		t.Fatal("the card is still in the graveyard")
	}
	if opp.Library.Size() == 0 || opp.Library.Cards[0].InstanceID != dead {
		t.Error("the card is not on the bottom of its owner's library")
	}
	if effectstest.TappedNow(g, pilgrim) {
		t.Error("the Pilgrim tapped; the cost has no {T}")
	}
}

// --- Goldfury Strider ------------------------------------------------------------

func TestLciDLGoldfuryStriderTapsTwoToPumpACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	strider := effectstest.B12Push(g, me.ID, "Goldfury Strider", "Artifact Creature — Golem", lciDLStriderOracle, 3, 5)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	if !effectstest.HasEffectiveKeyword(t, g, strider, "trample") {
		t.Error("Goldfury Strider has no trample")
	}
	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{strider, land}, Targets: effectstest.RfCardTarget(bear), Strict: true}); err == nil {
		t.Fatal("a land paid for \"artifacts and/or creatures\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{strider, relic}, Targets: effectstest.RfCardTarget(bear), Strict: true}); err != nil {
		t.Fatalf("tapping the Strider and an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
}

func TestLciDLGoldfuryStriderIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	strider := effectstest.B12Push(g, me.ID, "Goldfury Strider", "Artifact Creature — Golem", lciDLStriderOracle, 3, 5)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{strider, relic}, Targets: effectstest.RfCardTarget(strider), Strict: true}); err == nil {
		t.Error("the ability was activated at the beginning of combat")
	}
}

// --- Gargantuan Leech ------------------------------------------------------------

// Two Caves on the battlefield and one Cave card in the graveyard take
// {3} off; a land card that is not a Cave takes nothing.
func TestLciDLGargantuanLeechCostsLessForEachCave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushLandFor(g, me.ID, "Hidden Cave", "Land — Cave")
	effectstest.PushLandFor(g, me.ID, "Other Cave", "Land — Cave")
	effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.PushGraveyardCardTyped(me, "Old Cave", "Land — Cave")
	effectstest.PushGraveyardCardTyped(me, "Old Forest", "Basic Land — Forest")
	leech := effectstest.AvB24Hand(g, me, game.Card{Name: "Gargantuan Leech", TypeLine: "Creature — Leech", OracleID: lciDLLeechOracle, ManaCost: "{7}{B}", Power: 5, Toughness: 5})
	effectstest.FloatMana(t, g, me, "{B}{C}{C}{C}")
	if err := g.CastSpell(me.ID, leech, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for {3}{B}; three Caves take off only {3}")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, leech, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Gargantuan Leech for {4}{B} with three Caves: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, leech, "lifelink") {
		t.Error("Gargantuan Leech has no lifelink")
	}
}

// --- Join the Dead ---------------------------------------------------------------

func TestLciDLJoinTheDeadShrinksMoreWithDescendFour(t *testing.T) {
	for _, tc := range []struct {
		name       string
		permanents int
		want       int
	}{
		{"three permanent cards", 3, 7},
		{"four permanent cards", 4, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			giant := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 12, 12)
			lciAC2SeedGraveyard(g, me, tc.permanents)
			effectstest.CastCatalogSpell(t, g, "Join the Dead", "Instant", lciDLJoinOracle, effectstest.RfCardTarget(giant))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.EffectivePTOf(t, g, giant); p != tc.want || tough != tc.want {
				t.Errorf("the giant is %d/%d, want %d/%d", p, tough, tc.want, tc.want)
			}
		})
	}
}

// --- Huatli's Final Strike -------------------------------------------------------

// The +1/+0 lands before the damage, so a 3/3 hits for 4.
func TestLciDLHuatlisFinalStrikePumpsThenBites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Huatli's Final Strike", "Instant", lciDLFinalStrikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 4/4 survived 4 damage")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 4 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 4/3", p, tough)
	}
}

func TestLciDLHuatlisFinalStrikeNeedsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Huatli's Final Strike", "Instant", lciDLFinalStrikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: other, Slot: 1},
	}); err == nil {
		t.Error("the second target was a creature you control")
	}
}

// --- Hunter's Blowgun ------------------------------------------------------------

func TestLciDLHuntersBlowgunDeathtouchOnYourTurnReachOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	gun := effectstest.SeedEquipment(g, me.ID, "Hunter's Blowgun", lciDLBlowgunOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, gun, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped P/T %d/%d, want 3/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") || effectstest.EffectiveAbilitiesContain(t, g, bear, "reach") {
		t.Error("on your turn the equipped creature should have deathtouch and not reach")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") || !effectstest.EffectiveAbilitiesContain(t, g, bear, "reach") {
		t.Error("on an opponent's turn the equipped creature should have reach and not deathtouch")
	}
}

// --- Diamond Pick-Axe ------------------------------------------------------------

func TestLciDLDiamondPickAxeMakesATreasureWhenItsCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	axe := effectstest.SeedEquipment(g, me.ID, "Diamond Pick-Axe", lciDLPickAxeOracle)
	if !effectstest.HasEffectiveKeyword(t, g, axe, "indestructible") {
		t.Error("Diamond Pick-Axe is not indestructible")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, axe, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped P/T %d/%d, want 3/3", p, tough)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.TrekB33Tokens(g, me.ID, "Treasure")); got != 1 {
		t.Errorf("%d Treasure tokens, want 1", got)
	}
}

// --- Fungal Fortitude ------------------------------------------------------------

func TestLciDLFungalFortitudePumpsAndReturnsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Fungal Fortitude", lciDLFungalOracle, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("enchanted P/T %d/%d, want 4/2", p, tough)
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	ids := effectstest.BattlefieldIDsNamed(g, "Bear")
	if len(ids) != 1 {
		t.Fatalf("%d Bears on the battlefield, want 1", len(ids))
	}
	if !effectstest.B16Tapped(t, g, ids[0]) {
		t.Error("the Bear returned untapped")
	}
}

// --- Dusk Rose Reliquary ---------------------------------------------------------

func TestLciDLDuskRoseReliquaryExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "My Fodder", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	reliquary := effectstest.CastWithSacrifice(t, g, "Dusk Rose Reliquary", "Artifact", lciDLReliquaryOracle, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed as the spell was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the opponent's creature was not exiled")
	}
	effectstest.KillCreature(t, g, me.ID, reliquary)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Their Bear") {
		t.Error("the exiled creature did not return when the Reliquary left")
	}
}

func TestLciDLDuskRoseReliquaryNeedsSomethingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.CastCatalogSpellErr(t, g, "Dusk Rose Reliquary", "Artifact", lciDLReliquaryOracle, nil); err == nil {
		t.Error("cast with no artifact or creature to sacrifice")
	}
}

// --- Kutzil's Flanker ------------------------------------------------------------

// Two of your creatures left (one to the graveyard, one to exile); an
// opponent's creature dying does not count.
func TestLciDLKutzilsFlankerCountsYourCreaturesThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2))
	effectstest.ExileCreature(t, g, effectstest.B12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2))
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	flanker := lciDLFlankerModes(t, g, me, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, flanker, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
}

func TestLciDLKutzilsFlankerGainsTwoAndScriesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Card A", "Card B", "Card C")
	life := me.Life
	lciDLFlankerModes(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if sc := effectstest.ScryChoiceFor(g, me.ID); sc == nil {
		t.Error("no scry prompt")
	}
}

func TestLciDLKutzilsFlankerExilesTargetPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
	mine := effectstest.PushGraveyardCardTyped(me, "My Dead Bear", "Creature — Bear")
	lciDLFlankerModes(t, g, me, 2)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || opp.Graveyard.Size() != 0 {
		t.Error("the target player's graveyard was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("your own graveyard was touched")
	}
}

// --- Itzquinth, Firstborn of Gishath ---------------------------------------------

func TestLciDLItzquinthPaysForADinosaurBite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := effectstest.B12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 5, 5)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	itz := effectstest.CastRingCreature(t, g, "Itzquinth, Firstborn of Gishath", "Legendary Creature — Dinosaur", lciDLItzquinthOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, itz, "haste") {
		t.Error("Itzquinth has no haste")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for before the payment was made")
	}
	effectstest.B06AddMana(me, "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, bear) || effectstest.HasID(pick.PickTargetCards, theirs) || !effectstest.HasID(pick.PickTargetCards, dino) {
		t.Errorf("first clause offered %v; want Dinosaurs you control only", pick.PickTargetCards)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, dino)
	pick = effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the second target")
	}
	if effectstest.HasID(pick.PickTargetCards, dino) {
		t.Errorf("second clause offered the biting Dinosaur: %v", pick.PickTargetCards)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 4/4 survived 5 damage")
	}
}

func TestLciDLItzquinthDecliningDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.CastRingCreature(t, g, "Itzquinth, Firstborn of Gishath", "Legendary Creature — Dinosaur", lciDLItzquinthOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for after declining")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature was damaged without the payment")
	}
}
