package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsfl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/fl"
	cardsg "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/g"
	cardsgr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/gr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_dg_cards_test.go — Wilds of Eldraine, the D–G batch (#4471). Each
// test plays the real catalog entry and asserts what its text says.

const (
	woeDGPackOracle       = "ef702a1a-4ace-43e0-8e86-dcceb0b1e4d5"
	woeDGCauldronOracle   = "21153db6-552a-4956-a53a-a9dfb6968629"
	woeDGFoodComaOracle   = "aab3774c-4ec3-4946-a09b-af0fab7cacb7"
	woeDGFreezeOracle     = "8bf90bb0-f9c1-49c8-905e-32747e921e92"
	woeDGFamiliarOracle   = cardsfl.FrolickingFamiliarOracleID
	woeDGFrostbridgeOracl = "515c1604-59f0-45b4-91be-d2f20fdd3e1e"
	woeDGPieWielderOracle = "fba2fa87-3f2a-4a9f-a19e-a1f9dada4118"
	woeDGHunterOracle     = cardsg.GingerbreadHunterOracleID
	woeDGCasketOracle     = "d9167046-7908-4d16-8cc1-3eb02d1ba547"
	woeDGGrabbyOracle     = cardsgr.GrabbyGiantOracleID
	woeDGBallGuestOracle  = "0b0fdb28-f7b1-4c30-9f2c-14a9074da22d"
	woeDGGretaOracle      = "00173df7-a584-410c-af1d-ada9c791056a"
)

func TestWoeDGEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeDGPackOracle:            "Edgewall Pack",
		woeDGCauldronOracle:        "Feed the Cauldron",
		woeDGFoodComaOracle:        "Food Coma",
		woeDGFreezeOracle:          "Freeze in Place",
		woeDGFamiliarOracle:        "Frolicking Familiar",
		woeDGFamiliarOracle + "#1": "Blow Off Steam",
		woeDGFrostbridgeOracl:      "Frostbridge Guard",
		woeDGPieWielderOracle:      "Gallant Pie-Wielder",
		woeDGHunterOracle:          "Gingerbread Hunter",
		woeDGHunterOracle + "#1":   "Puny Snack",
		woeDGCasketOracle:          "Glass Casket",
		woeDGGrabbyOracle:          "Grabby Giant",
		woeDGGrabbyOracle + "#1":   "That's Mine",
		woeDGBallGuestOracle:       "Grand Ball Guest",
		woeDGGretaOracle:           "Greta, Sweettooth Scourge",
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

// woeDGCastAfterAPass puts a card in p's hand and casts it after the
// active player passes priority in their precombat main phase: an
// instant cast by a player whose turn it is not.
func woeDGCastAfterAPass(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// woeDGCelebrate makes celebration hold for `you` this turn: tokens
// enter until two or more nonland permanents have.
func woeDGCelebrate(t *testing.T, g *game.Game, you uuid.UUID) {
	t.Helper()
	if n := celebrationCountFor(g, you); n < 2 {
		effectstest.CelebrationTokens(t, g, you, 2-n)
	}
}

// --- Edgewall Pack ---------------------------------------------------------------

func TestWoeDGEdgewallPackHasMenaceAndMakesARatThatCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	pack := effectstest.CastCatalogSpell(t, g, "Edgewall Pack", "Creature — Dog", woeDGPackOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pack) {
		t.Fatal("Edgewall Pack did not resolve")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rat tokens after it entered, want 1", got)
	}
	effectstest.AssertKeywords(t, g, pack, "menace")
	rat := effectstest.FindBattlefieldByName(g, "Rat")
	if c, _ := effectstest.BattlefieldCard(g, rat); c.Power != 1 || c.Toughness != 1 {
		t.Errorf("the Rat is %d/%d, want 1/1", c.Power, c.Toughness)
	}
	effectstest.AssertRestrictions(t, g, rat, game.CantBlock)
}

// --- Feed the Cauldron -----------------------------------------------------------

func TestWoeDGFeedTheCauldronDestroysAndMakesAFoodOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	small := effectstest.WoeB1CreatureOfCost(g, opp.ID, "Small", "{1}{G}{G}")
	effectstest.CastCatalogSpell(t, g, "Feed the Cauldron", "Instant", woeDGCauldronOracle, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Fatal("the mana value 3 creature was not destroyed")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food on your turn, want 1", got)
	}
}

func TestWoeDGFeedTheCauldronMakesNoFoodOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active, other := effectstest.WoeB1Seats(g)
	bear := effectstest.WoeB1CreatureOfCost(g, active.ID, "Bear", "{1}{G}")
	woeDGCastAfterAPass(t, g, other, "Feed the Cauldron", "Instant", woeDGCauldronOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not destroyed")
	}
	if got := effectstest.B27CountNamed(g, other.ID, "Food"); got != 0 {
		t.Errorf("%d Food when cast in another player's turn, want 0", got)
	}
}

func TestWoeDGFeedTheCauldronRefusesAManaValueFourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.WoeB1Seats(g)
	big := effectstest.WoeB1CreatureOfCost(g, opp.ID, "Big", "{2}{G}{G}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Feed the Cauldron", "Instant", woeDGCauldronOracle, effectstest.CardRefs(big)); err == nil {
		t.Fatal("a mana value 4 creature was a legal target")
	}
}

// --- Food Coma -------------------------------------------------------------------

func TestWoeDGFoodComaExilesAnOpponentsCreatureUntilItLeavesAndMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Their Other Bear", woeMPBearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	coma := effectstest.CastCatalogSpell(t, g, "Food Coma", "Enchantment", woeDGFoodComaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("Food Coma targeted your own creature")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || g.Battlefield.Contains(theirs) {
		t.Fatal("the opponent's creature was not exiled")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the other creature left too")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after the trigger, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(coma) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after Food Coma left, %d of the exiled creature under its owner's control, want 1", n)
	}
}

// --- Freeze in Place -------------------------------------------------------------

func TestWoeDGFreezeInPlaceTapsStunsThreeAndScriesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Freeze in Place", "Sorcery", woeDGFreezeOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if n := effectstest.CountersOn(g, bear, game.CounterStun); n != 3 {
		t.Errorf("%d stun counters, want 3", n)
	}
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestWoeDGFreezeInPlaceRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Freeze in Place", "Sorcery", woeDGFreezeOracle, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("Freeze in Place targeted your own creature")
	}
}

// --- Frolicking Familiar // Blow Off Steam ---------------------------------------

func TestWoeDGFrolickingFamiliarFliesAndGrowsOnInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	familiar := effectstest.B12Push(g, me.ID, "Frolicking Familiar", "Creature — Otter Wizard", woeDGFamiliarOracle, 2, 2)
	effectstest.AssertKeywords(t, g, familiar, "flying")
	effectstest.CastCatalogSpell(t, g, "Some Creature", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, familiar); p != 2 || tough != 2 {
		t.Fatalf("a creature spell pumped it to %d/%d", p, tough)
	}
	effectstest.CastCatalogSpell(t, g, "Some Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, familiar); p != 4 || tough != 4 {
		t.Errorf("after an instant and a sorcery it is %d/%d, want 4/4", p, tough)
	}
	effectstest.EndTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, familiar); p != 2 || tough != 2 {
		t.Errorf("the pump outlasted the turn: %d/%d", p, tough)
	}
}

func TestWoeDGBlowOffSteamDealsOneToAnyTargetAndGoesOnAnAdventure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	before := opp.Life
	card := woeB1Adventure(me, woeDGFamiliarOracle,
		game.Face{Name: "Frolicking Familiar", TypeLine: "Creature — Otter Wizard", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 2, Toughness: 2},
		game.Face{Name: "Blow Off Steam", TypeLine: "Instant — Adventure", ManaCost: "{R}", Colors: []string{"R"}})
	woeB1CastAdventure(t, g, me, card, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Errorf("the opponent is at %d, want %d", got, before-1)
	}
}

// --- Frostbridge Guard -----------------------------------------------------------

func TestWoeDGFrostbridgeGuardTapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	guard := effectstest.B12Push(g, me.ID, "Frostbridge Guard", "Creature — Elemental Soldier", woeDGFrostbridgeOracl, 2, 2)
	woeMPReady(g, guard)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("the tap ability activated without {2}{W}")
	}
	effectstest.FloatForTest(g, me, "WCC")
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("{2}{W}, {T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the target was not tapped")
	}
	if c, _ := effectstest.BattlefieldCard(g, guard); !c.Tapped {
		t.Error("the Guard did not tap to pay its cost")
	}
}

func TestWoeDGFrostbridgeGuardCantTapWhileSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	guard := effectstest.PushCatalogPermanent(g, me.ID, "Frostbridge Guard", "Creature — Elemental Soldier", woeDGFrostbridgeOracl, true)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a summoning-sick Guard paid {T}")
	}
}

// --- Gallant Pie-Wielder ---------------------------------------------------------

func TestWoeDGGallantPieWielderHasDoubleStrikeOnlyWhileCelebrating(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pie := effectstest.PushCatalogPermanent(g, me.ID, "Gallant Pie-Wielder", "Creature — Dwarf Knight", woeDGPieWielderOracle, false)
	if celebrationCountFor(g, me.ID) >= 2 {
		t.Fatal("setup: celebration before anything entered")
	}
	effectstest.AssertKeywords(t, g, pie, "first strike")
	if effectstest.HasEffectiveKeyword(t, g, pie, "double strike") {
		t.Fatal("double strike without celebration")
	}
	woeDGCelebrate(t, g, me.ID)
	if !effectstest.HasEffectiveKeyword(t, g, pie, "double strike") {
		t.Error("no double strike while celebrating")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, pie, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

// --- Gingerbread Hunter // Puny Snack --------------------------------------------

func woeDGHunter(p *game.Player) game.Card {
	return woeB1Adventure(p, woeDGHunterOracle,
		game.Face{Name: "Gingerbread Hunter", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Colors: []string{"G"}, Power: 5, Toughness: 5},
		game.Face{Name: "Puny Snack", TypeLine: "Instant — Adventure", ManaCost: "{2}{B}", Colors: []string{"B"}})
}

func TestWoeDGGingerbreadHunterMakesAFoodOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	card := woeDGHunter(me)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Fatal("Gingerbread Hunter did not resolve")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after it entered, want 1", got)
	}
}

func TestWoeDGPunySnackShrinksTwoTwoAndGoesOnAnAdventure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	big := effectstest.B12Creature(g, opp.ID, "Their Big Bear", woeMPBearType, 4, 4)
	woeB1CastAdventure(t, g, me, woeDGHunter(me), effectstest.CardRefs(big))
	if p, tough := effectstest.EffectivePTOf(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, tough)
	}
	effectstest.EndTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, big); p != 4 || tough != 4 {
		t.Errorf("the shrink outlasted the turn: %d/%d", p, tough)
	}
}

// --- Glass Casket ----------------------------------------------------------------

func TestWoeDGGlassCasketExilesASmallOpposingCreatureUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	small := effectstest.WoeB1CreatureOfCost(g, opp.ID, "Small", "{1}{G}{G}")
	tiny := effectstest.WoeB1CreatureOfCost(g, opp.ID, "Tiny", "{G}")
	big := effectstest.WoeB1CreatureOfCost(g, opp.ID, "Big", "{2}{G}{G}")
	mine := effectstest.WoeB1CreatureOfCost(g, me.ID, "Mine", "{G}")
	casket := effectstest.CastCatalogSpell(t, g, "Glass Casket", "Artifact", woeDGCasketOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	for name, id := range map[string]uuid.UUID{"a mana value 4 creature": big, "your own creature": mine} {
		if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(id)); err == nil {
			t.Fatalf("Glass Casket targeted %s", name)
		}
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(small) || g.Battlefield.Contains(small) {
		t.Fatal("the mana value 3 creature was not exiled")
	}
	if !g.Battlefield.Contains(tiny) {
		t.Error("the other creature left too")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(casket) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Small", opp.ID); n != 1 {
		t.Errorf("after the Casket left, %d of the exiled creature under its owner's control, want 1", n)
	}
}

// --- Grabby Giant // That's Mine -------------------------------------------------

func TestWoeDGGrabbyGiantHasReachAndSacrificesAnArtifactOrLandToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Grabby Giant", "Creature — Giant", woeDGGrabbyOracle, true)
	effectstest.AssertKeywords(t, g, giant, "reach")
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature paid the artifact-or-land sacrifice")
	}
	hand := effectstest.HandSize(me)
	effectstest.FloatForTest(g, me, "RCC")
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("{2}{R}, sacrifice a land (while summoning sick): %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand %d after the draw, want %d", got, hand+1)
	}
}

func TestWoeDGThatsMineMakesATreasureAndGoesOnAnAdventure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	card := woeB1Adventure(me, woeDGGrabbyOracle,
		game.Face{Name: "Grabby Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{R}", Colors: []string{"R"}, Power: 4, Toughness: 3},
		game.Face{Name: "That's Mine", TypeLine: "Instant — Adventure", ManaCost: "{1}{R}", Colors: []string{"R"}})
	woeB1CastAdventure(t, g, me, card, nil)
	if got := effectstest.B27CountNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasure, want 1", got)
	}
}

// --- Grand Ball Guest ------------------------------------------------------------

func TestWoeDGGrandBallGuestGetsPlusOneAndTrampleWhileCelebrating(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guest := effectstest.B12Push(g, me.ID, "Grand Ball Guest", "Creature — Human Peasant", woeDGBallGuestOracle, 2, 2)
	if celebrationCountFor(g, me.ID) >= 2 {
		t.Fatal("setup: celebration before anything entered")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, guest); p != 2 || tough != 2 || effectstest.HasEffectiveKeyword(t, g, guest, "trample") {
		t.Fatalf("without celebration it is %d/%d (trample %v), want a plain 2/2", p, tough, effectstest.HasEffectiveKeyword(t, g, guest, "trample"))
	}
	woeDGCelebrate(t, g, me.ID)
	if p, tough := effectstest.EffectivePTOf(t, g, guest); p != 3 || tough != 3 {
		t.Errorf("while celebrating it is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, guest, "trample") {
		t.Error("no trample while celebrating")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, guest); p != 2 || tough != 2 || effectstest.HasEffectiveKeyword(t, g, guest, "trample") {
		t.Errorf("on the next turn it is %d/%d (trample %v), want a plain 2/2", p, tough, effectstest.HasEffectiveKeyword(t, g, guest, "trample"))
	}
}

// --- Greta, Sweettooth Scourge ---------------------------------------------------

func TestWoeDGGretaMakesAFoodAndEatsFoodForACounterOrACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	greta := effectstest.CastCatalogSpell(t, g, "Greta, Sweettooth Scourge", "Legendary Creature — Human Warrior", woeDGGretaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("%d Food after Greta entered, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1) })
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	notFood := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")

	if err := g.ActivateCatalogAbility(me.ID, greta, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{notFood}}); err == nil {
		t.Fatal("an artifact that is not a Food paid the cost")
	}
	food := effectstest.FindBattlefieldByName(g, "Food")
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, greta, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("{G}, sacrifice a Food (while summoning sick): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters on the target, want 1", n)
	}

	life, hand := me.Life, effectstest.HandSize(me)
	food = effectstest.FindBattlefieldByName(g, "Food")
	effectstest.FloatForTest(g, me, "BC")
	if err := g.ActivateCatalogAbility(me.ID, greta, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("{1}{B}, sacrifice a Food: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand %d after the draw, want %d", got, hand+1)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d, want %d", got, life-1)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food left, want 0", got)
	}
}

func TestWoeDGGretasCounterAbilityIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	greta := effectstest.PushCatalogPermanent(g, me.ID, "Greta, Sweettooth Scourge", "Legendary Creature — Human Warrior", woeDGGretaOracle, false)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1) })
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	food := effectstest.FindBattlefieldByName(g, "Food")
	if err := g.ActivateCatalogAbility(me.ID, greta, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{food}}); err == nil {
		t.Fatal("the counter ability activated in the beginning of combat step")
	}
}
