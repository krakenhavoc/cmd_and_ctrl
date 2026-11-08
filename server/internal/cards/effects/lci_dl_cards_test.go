package effects

import (
	"testing"

	"github.com/google/uuid"

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

// --- fixtures ------------------------------------------------------------------

// lciDLFlankerModes casts Kutzil's Flanker, resolves the spell and picks
// the enters trigger's bullet.
func lciDLFlankerModes(t *testing.T, g *game.Game, me *game.Player, mode int) uuid.UUID {
	t.Helper()
	flanker := castAndResolveCreature(t, g, "Kutzil's Flanker", "Creature — Cat Warrior", lciDLFlankerOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aspirant := castHoldCreature(t, g, "Ironpaw Aspirant", "Creature — Cat Warrior", lciDLIronpawOracle, 1, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no enters-trigger target prompt")
	}
	if !hasID(pick.PickTargetCards, theirs) || !hasID(pick.PickTargetCards, aspirant) {
		t.Errorf("offered %v; want any creature", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the target, want 1", got)
	}
	if got := countersOn(g, aspirant, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters on the Aspirant, want 0", got)
	}
}

// --- Kinjalli's Dawnrunner -------------------------------------------------------

func TestLciDLKinjallisDawnrunnerExploresOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	top := putOnTopForTest(me, "Some Sorcery", "Sorcery")
	runner := msc17Play(t, g, me, game.Card{Name: "Kinjalli's Dawnrunner", TypeLine: "Creature — Human Scout", OracleID: lciDLDawnrunnerOracle, Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, runner, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after revealing a nonland, want 1", got)
	}
	answerExploreQuestion(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the revealed card was not put into the graveyard")
	}
	if !hasEffectiveKeyword(t, g, runner, "double strike") {
		t.Error("Kinjalli's Dawnrunner has no double strike")
	}
}

// --- Hermitic Nautilus -----------------------------------------------------------

func TestLciDLHermiticNautilusTradesToughnessForPower(t *testing.T) {
	g, me, _ := spendTable(t)
	naut := b12Push(g, me.ID, "Hermitic Nautilus", "Artifact Creature — Nautilus", lciDLNautilusOracle, 1, 4)
	if !hasEffectiveKeyword(t, g, naut, "vigilance") {
		t.Error("Hermitic Nautilus has no vigilance")
	}
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, naut, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the {U}")
	}
	apaMana(me, "U")
	p7Activate(t, g, me, naut, 0, game.ActivateAbilityParams{Strict: true})
	if p, tough := effectivePTOf(t, g, naut); p != 4 || tough != 1 {
		t.Errorf("after one activation the Nautilus is %d/%d, want 4/1", p, tough)
	}
	apaMana(me, "C", "U")
	p7Activate(t, g, me, naut, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(naut) {
		t.Error("a 7/-2 Nautilus is still on the battlefield")
	}
}

// --- Hotfoot Gnome ---------------------------------------------------------------

func TestLciDLHotfootGnomeGivesAnotherCreatureHaste(t *testing.T) {
	g, me, _ := spendTable(t)
	gnome := b12Push(g, me.ID, "Hotfoot Gnome", "Artifact Creature — Gnome", lciDLHotfootOracle, 3, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, gnome, "haste") {
		t.Error("Hotfoot Gnome has no haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, gnome, 0, game.ActivateAbilityParams{Targets: rfCardTarget(gnome)}); err == nil {
		t.Fatal("the Gnome targeted itself; the text says another")
	}
	p7Activate(t, g, me, gnome, 0, game.ActivateAbilityParams{Targets: rfCardTarget(bear)})
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the targeted creature did not gain haste")
	}
	if !tappedNow(g, gnome) {
		t.Error("the Gnome did not tap for its ability")
	}
}

// --- Hoverstone Pilgrim ----------------------------------------------------------

func TestLciDLHoverstonePilgrimTucksAGraveyardCardUnderItsOwnersLibrary(t *testing.T) {
	g, me, opp := spendTable(t)
	pilgrim := b12Push(g, me.ID, "Hoverstone Pilgrim", "Artifact Creature — Golem", lciDLPilgrimOracle, 2, 5)
	if !hasEffectiveKeyword(t, g, pilgrim, "flying") {
		t.Error("Hoverstone Pilgrim has no flying")
	}
	dead := pushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
	putOnTopForTest(opp, "Their Top Card", "Sorcery")
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, pilgrim, 0, game.ActivateAbilityParams{Targets: rfCardTarget(dead), Strict: true}); err == nil {
		t.Fatal("activated with one mana for {2}")
	}
	apaMana(me, "C")
	p7Activate(t, g, me, pilgrim, 0, game.ActivateAbilityParams{Targets: rfCardTarget(dead), Strict: true})
	if opp.Graveyard.Contains(dead) {
		t.Fatal("the card is still in the graveyard")
	}
	if opp.Library.Size() == 0 || opp.Library.Cards[0].InstanceID != dead {
		t.Error("the card is not on the bottom of its owner's library")
	}
	if tappedNow(g, pilgrim) {
		t.Error("the Pilgrim tapped; the cost has no {T}")
	}
}

// --- Goldfury Strider ------------------------------------------------------------

func TestLciDLGoldfuryStriderTapsTwoToPumpACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	strider := b12Push(g, me.ID, "Goldfury Strider", "Artifact Creature — Golem", lciDLStriderOracle, 3, 5)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	if !hasEffectiveKeyword(t, g, strider, "trample") {
		t.Error("Goldfury Strider has no trample")
	}
	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{strider, land}, Targets: rfCardTarget(bear), Strict: true}); err == nil {
		t.Fatal("a land paid for \"artifacts and/or creatures\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{strider, relic}, Targets: rfCardTarget(bear), Strict: true}); err != nil {
		t.Fatalf("tapping the Strider and an artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
}

func TestLciDLGoldfuryStriderIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	strider := b12Push(g, me.ID, "Goldfury Strider", "Artifact Creature — Golem", lciDLStriderOracle, 3, 5)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{strider, relic}, Targets: rfCardTarget(strider), Strict: true}); err == nil {
		t.Error("the ability was activated at the beginning of combat")
	}
}

// --- Gargantuan Leech ------------------------------------------------------------

// Two Caves on the battlefield and one Cave card in the graveyard take
// {3} off; a land card that is not a Cave takes nothing.
func TestLciDLGargantuanLeechCostsLessForEachCave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushLandFor(g, me.ID, "Hidden Cave", "Land — Cave")
	pushLandFor(g, me.ID, "Other Cave", "Land — Cave")
	pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	pushGraveyardCardTyped(me, "Old Cave", "Land — Cave")
	pushGraveyardCardTyped(me, "Old Forest", "Basic Land — Forest")
	leech := avB24Hand(g, me, game.Card{Name: "Gargantuan Leech", TypeLine: "Creature — Leech", OracleID: lciDLLeechOracle, ManaCost: "{7}{B}", Power: 5, Toughness: 5})
	floatMana(t, g, me, "{B}{C}{C}{C}")
	if err := g.CastSpell(me.ID, leech, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for {3}{B}; three Caves take off only {3}")
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, leech, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Gargantuan Leech for {4}{B} with three Caves: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, leech, "lifelink") {
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 12, 12)
			lciAC2SeedGraveyard(g, me, tc.permanents)
			castCatalogSpell(t, g, "Join the Dead", "Instant", lciDLJoinOracle, rfCardTarget(giant))
			passPriorityAroundTable(t, g)
			if p, tough := effectivePTOf(t, g, giant); p != tc.want || tough != tc.want {
				t.Errorf("the giant is %d/%d, want %d/%d", p, tough, tc.want, tc.want)
			}
		})
	}
}

// --- Huatli's Final Strike -------------------------------------------------------

// The +1/+0 lands before the damage, so a 3/3 hits for 4.
func TestLciDLHuatlisFinalStrikePumpsThenBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	castCatalogSpell(t, g, "Huatli's Final Strike", "Instant", lciDLFinalStrikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 4/4 survived 4 damage")
	}
	if p, tough := effectivePTOf(t, g, mine); p != 4 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 4/3", p, tough)
	}
}

func TestLciDLHuatlisFinalStrikeNeedsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	other := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Huatli's Final Strike", "Instant", lciDLFinalStrikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: other, Slot: 1},
	}); err == nil {
		t.Error("the second target was a creature you control")
	}
}

// --- Hunter's Blowgun ------------------------------------------------------------

func TestLciDLHuntersBlowgunDeathtouchOnYourTurnReachOtherwise(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	gun := seedEquipment(g, me.ID, "Hunter's Blowgun", lciDLBlowgunOracle)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, gun, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped P/T %d/%d, want 3/3", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "deathtouch") || effectiveAbilitiesContain(t, g, bear, "reach") {
		t.Error("on your turn the equipped creature should have deathtouch and not reach")
	}
	advanceToNextSeatsTurn(t, g)
	if effectiveAbilitiesContain(t, g, bear, "deathtouch") || !effectiveAbilitiesContain(t, g, bear, "reach") {
		t.Error("on an opponent's turn the equipped creature should have reach and not deathtouch")
	}
}

// --- Diamond Pick-Axe ------------------------------------------------------------

func TestLciDLDiamondPickAxeMakesATreasureWhenItsCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	bear := seedBear(g, me.ID)
	axe := seedEquipment(g, me.ID, "Diamond Pick-Axe", lciDLPickAxeOracle)
	if !hasEffectiveKeyword(t, g, axe, "indestructible") {
		t.Error("Diamond Pick-Axe is not indestructible")
	}
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, axe, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped P/T %d/%d, want 3/3", p, tough)
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := len(trekB33Tokens(g, me.ID, "Treasure")); got != 1 {
		t.Errorf("%d Treasure tokens, want 1", got)
	}
}

// --- Fungal Fortitude ------------------------------------------------------------

func TestLciDLFungalFortitudePumpsAndReturnsTheCreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Fungal Fortitude", lciDLFungalOracle, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("enchanted P/T %d/%d, want 4/2", p, tough)
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	ids := battlefieldIDsNamed(g, "Bear")
	if len(ids) != 1 {
		t.Fatalf("%d Bears on the battlefield, want 1", len(ids))
	}
	if !b16Tapped(t, g, ids[0]) {
		t.Error("the Bear returned untapped")
	}
}

// --- Dusk Rose Reliquary ---------------------------------------------------------

func TestLciDLDuskRoseReliquaryExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Creature(g, me.ID, "My Fodder", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	reliquary := castWithSacrifice(t, g, "Dusk Rose Reliquary", "Artifact", lciDLReliquaryOracle, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed as the spell was cast")
	}
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, theirs)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the opponent's creature was not exiled")
	}
	killCreature(t, g, me.ID, reliquary)
	passPriorityAroundTable(t, g)
	if !onBattlefieldByName(g, "Their Bear") {
		t.Error("the exiled creature did not return when the Reliquary left")
	}
}

func TestLciDLDuskRoseReliquaryNeedsSomethingToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	if err := castCatalogSpellErr(t, g, "Dusk Rose Reliquary", "Artifact", lciDLReliquaryOracle, nil); err == nil {
		t.Error("cast with no artifact or creature to sacrifice")
	}
}

// --- Kutzil's Flanker ------------------------------------------------------------

// Two of your creatures left (one to the graveyard, one to exile); an
// opponent's creature dying does not count.
func TestLciDLKutzilsFlankerCountsYourCreaturesThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	avB9Sacrifice(t, g, b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2))
	exileCreature(t, g, b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2))
	avB9Sacrifice(t, g, b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	flanker := lciDLFlankerModes(t, g, me, 0)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, flanker, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
}

func TestLciDLKutzilsFlankerGainsTwoAndScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Card A", "Card B", "Card C")
	life := me.Life
	lciDLFlankerModes(t, g, me, 1)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if sc := scryChoiceFor(g, me.ID); sc == nil {
		t.Error("no scry prompt")
	}
}

func TestLciDLKutzilsFlankerExilesTargetPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
	mine := pushGraveyardCardTyped(me, "My Dead Bear", "Creature — Bear")
	lciDLFlankerModes(t, g, me, 2)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || opp.Graveyard.Size() != 0 {
		t.Error("the target player's graveyard was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("your own graveyard was touched")
	}
}

// --- Itzquinth, Firstborn of Gishath ---------------------------------------------

func TestLciDLItzquinthPaysForADinosaurBite(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := b12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 5, 5)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	itz := castRingCreature(t, g, "Itzquinth, Firstborn of Gishath", "Legendary Creature — Dinosaur", lciDLItzquinthOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, itz, "haste") {
		t.Error("Itzquinth has no haste")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for before the payment was made")
	}
	b06AddMana(me, "C", "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if hasID(pick.PickTargetCards, bear) || hasID(pick.PickTargetCards, theirs) || !hasID(pick.PickTargetCards, dino) {
		t.Errorf("first clause offered %v; want Dinosaurs you control only", pick.PickTargetCards)
	}
	pickTriggerTarget(t, g, me.ID, dino)
	pick = latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the second target")
	}
	if hasID(pick.PickTargetCards, dino) {
		t.Errorf("second clause offered the biting Dinosaur: %v", pick.PickTargetCards)
	}
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 4/4 survived 5 damage")
	}
}

func TestLciDLItzquinthDecliningDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 5, 5)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	castRingCreature(t, g, "Itzquinth, Firstborn of Gishath", "Legendary Creature — Dinosaur", lciDLItzquinthOracle, 2, 3)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for after declining")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature was damaged without the payment")
	}
}
