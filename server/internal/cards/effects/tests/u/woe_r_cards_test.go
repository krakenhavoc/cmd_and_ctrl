package testsu

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/r"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_r_cards_test.go — Wilds of Eldraine, the R batch (#4471). Each test
// plays the real catalog entry and asserts what its text says.

const (
	woeRMouseOracle      = "9ffe4d90-b46c-4606-89cb-1b724104f6ea"
	woeRTraineeOracle    = cardsr.RatcatcherTraineeOracleID
	woeRRatOutOracle     = "37043e1e-4556-45f7-9484-15dd8ed5272b"
	woeRHellkiteOracle   = "50ace222-57f1-477f-9e06-155d7b403512"
	woeRThiefOracle      = "e055bc6d-d395-4f85-8e15-3bce65b7fcca"
	woeRBunnicornOracle  = "efd93428-88a5-4f7b-95f3-1d6eec069a4c"
	woeRBivouacOracle    = "b3c7b46f-c9ab-40ca-b50b-a4e0d0bd9be8"
	woeRCottageOracle    = "7e16595f-bdeb-422e-b99a-bfc0ed52e9f8"
	woeRFortressOracle   = "8b3726f1-20b8-42ec-8f9b-b361515c3f05"
	woeRVinestalkOracle  = "0935faa2-fb90-48db-8a92-906ba0f374c7"
	woeRReturnOracle     = "541adb13-1277-4c9e-ae1e-976f7aac1c72"
	woeRReindeerOracle   = "66dc3237-3543-4fc2-96f1-c8d1eff7c04c"
	woeRFaunOracle       = "87bedc9b-b4d2-41b6-b9f1-e1aeb8e75ec3"
	woeRResearchOracle   = "1a33a5c2-5aa1-48e5-95a8-d52cdce980de"
	woeRRubyOracle       = "5d1b0eee-3a7a-4f22-a40d-7658a368962a"
	woeRBearType         = "Creature — Bear"
	woeRResearchTypeLine = "Instant"
)

func TestWoeREveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeRMouseOracle:          "Raging Battle Mouse",
		woeRTraineeOracle:        "Ratcatcher Trainee",
		woeRTraineeOracle + "#1": "Pest Problem",
		woeRRatOutOracle:         "Rat Out",
		woeRHellkiteOracle:       "Realm-Scorcher Hellkite",
		woeRThiefOracle:          "Redcap Thief",
		woeRBunnicornOracle:      "Regal Bunnicorn",
		woeRBivouacOracle:        "Restless Bivouac",
		woeRCottageOracle:        "Restless Cottage",
		woeRFortressOracle:       "Restless Fortress",
		woeRVinestalkOracle:      "Restless Vinestalk",
		woeRReturnOracle:         "Return from the Wilds",
		woeRReindeerOracle:       "Rimefur Reindeer",
		woeRFaunOracle:           "Rootrider Faun",
		woeRResearchOracle:       "Rowdy Research",
		woeRRubyOracle:           "Ruby, Daring Tracker",
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

// woeRAssertAnimated checks a Restless land's animated characteristics,
// its colours in any order.
func woeRAssertAnimated(t *testing.T, g *game.Game, land uuid.UUID, subtype string, colours []string, power, toughness int) {
	t.Helper()
	c := effectstest.S58p6View(t, g, land)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	if !slices.Contains(c.Effective().Subtypes, subtype) {
		t.Errorf("subtypes %v, want %s", c.Effective().Subtypes, subtype)
	}
	got, want := slices.Clone(c.EffectiveColors()), slices.Clone(colours)
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("colours %v, want %v", got, want)
	}
	if c.CurrentPower() != power || c.CurrentToughness() != toughness {
		t.Errorf("%d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), power, toughness)
	}
}

// woeRRestlessEntersTapped plays a Restless land from hand and checks it
// enters tapped.
func woeRRestlessEntersTapped(t *testing.T, name, oracle string) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	if id := effectstest.B12PlayFromHand(t, g, name, "Land", oracle, game.CastSpellParams{}); !effectstest.IsTapped(g, id) {
		t.Errorf("%s entered untapped", name)
	}
}

// woeRPassUntilManaPick passes priority until chooser owes a colour pick,
// at most three times.
func woeRPassUntilManaPick(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 3 && effectstest.RiderLatestManaPick(g, chooser) == nil; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
}

// --- Raging Battle Mouse -----------------------------------------------------------

func TestWoeRRagingBattleMouseDiscountsOnlyTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Raging Battle Mouse", "Creature — Mouse", woeRMouseOracle, false)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("first spell: %d, want 3 (undiscounted)", got)
	}
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 2 {
		t.Errorf("second spell: %d, want 2 ({1} less)", got)
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("third spell: %d, want 3 (undiscounted)", got)
	}
}

func TestWoeRRagingBattleMousePumpsATargetWhileCelebrating(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Raging Battle Mouse", "Creature — Mouse", woeRMouseOracle, false)
	effectstest.B12Creature(g, me.ID, "Grizzly Bears", woeRBearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggerOnStackFrom(g, "Raging Battle Mouse") {
		t.Fatal("the trigger fired with no celebration")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Raging Battle Mouse", "Creature — Mouse", woeRMouseOracle, false)
	bear := effectstest.B12Creature(g2, me2.ID, "Grizzly Bears", woeRBearType, 2, 2)
	effectstest.CelebrationTokens(t, g2, me2.ID, 2)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PickCard(t, g2, me2.ID, bear)
	effectstest.PassPriorityAroundTable(t, g2)
	if p, tough := effectstest.EffectivePTOf(t, g2, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g2)
	if p, tough := effectstest.EffectivePTOf(t, g2, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear is %d/%d on the next turn, want 2/2", p, tough)
	}
}

// --- Ratcatcher Trainee // Pest Problem --------------------------------------------

func TestWoeRRatcatcherTraineeHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	trainee := effectstest.B12Push(g, me.ID, "Ratcatcher Trainee", "Creature — Human Peasant", woeRTraineeOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, trainee, "first strike") {
		t.Error("no first strike on your turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, trainee, "first strike") {
		t.Error("first strike on an opponent's turn")
	}
}

func TestWoeRPestProblemMakesTwoRatsThatCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	card := woeB1Adventure(me, woeRTraineeOracle,
		game.Face{Name: "Ratcatcher Trainee", TypeLine: "Creature — Human Peasant", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 2, Toughness: 1},
		game.Face{Name: "Pest Problem", TypeLine: "Instant — Adventure", ManaCost: "{2}{R}", Colors: []string{"R"}})
	woeB1CastAdventure(t, g, me, card, nil)
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 2 {
		t.Fatalf("%d Rats, want 2", got)
	}
	var rats []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rat" && c.Controller == me.ID {
				rats = append(rats, c.InstanceID)
			}
		}
	})
	for _, rat := range rats {
		effectstest.AssertRestrictions(t, g, rat, game.CantBlock)
	}
}

// --- Rat Out ---------------------------------------------------------------------

func TestWoeRRatOutShrinksATargetAndMakesARat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rat Out", "Instant", woeRRatOutOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Errorf("%d Rats, want 1", got)
	}
}

func TestWoeRRatOutWithNoTargetStillMakesARat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.CastCatalogSpell(t, g, "Rat Out", "Instant", woeRRatOutOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Errorf("%d Rats, want 1", got)
	}
}

// --- Realm-Scorcher Hellkite --------------------------------------------------------

func TestWoeRRealmScorcherHellkiteBargainedAddsFourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	token := woeILToken(g, me.ID)
	effectstest.P1g2Cast(t, g, "Realm-Scorcher Hellkite", "Creature — Dragon", woeRHellkiteOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	woeRPassUntilManaPick(t, g, me.ID)
	if n := effectstest.RiderAnswerManaPicks(t, g, me.ID, "R"); n != 4 {
		t.Fatalf("answered %d colour picks, want 4", n)
	}
	if me.ManaPool.Total() != 4 {
		t.Errorf("pool has %d mana, want 4", me.ManaPool.Total())
	}
	hk := effectstest.FindBattlefieldByName(g, "Realm-Scorcher Hellkite")
	if hk == uuid.Nil {
		t.Fatal("the Hellkite is not on the battlefield")
	}
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, hk, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestWoeRRealmScorcherHellkiteUnbargainedAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.CastAndResolveCreature(t, g, "Realm-Scorcher Hellkite", "Creature — Dragon", woeRHellkiteOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RiderLatestManaPick(g, me.ID) != nil || me.ManaPool.Total() != 0 {
		t.Errorf("an unbargained Hellkite added mana (pool %d)", me.ManaPool.Total())
	}
}

func TestWoeRRealmScorcherHellkitePingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	hk := effectstest.B12Push(g, me.ID, "Realm-Scorcher Hellkite", "Creature — Dragon", woeRHellkiteOracle, 4, 6)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, hk, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
}

// --- Redcap Thief ------------------------------------------------------------------

func TestWoeRRedcapThiefMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.CastAndResolveCreature(t, g, "Redcap Thief", "Creature — Goblin Rogue", woeRThiefOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Regal Bunnicorn ---------------------------------------------------------------

func TestWoeRRegalBunnicornCountsNonlandPermanentsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	bunny := effectstest.B12Push(g, me.ID, "Regal Bunnicorn", "Creature — Rabbit Unicorn", woeRBunnicornOracle, 0, 0)
	if p, tough := effectstest.EffectivePTOf(t, g, bunny); p != 1 || tough != 1 {
		t.Errorf("alone it is %d/%d, want 1/1", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	effectstest.B12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	effectstest.B12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, bunny); p != 2 || tough != 2 {
		t.Errorf("with an artifact and a land it is %d/%d, want 2/2", p, tough)
	}
}

// --- Restless Bivouac --------------------------------------------------------------

func TestWoeRRestlessBivouacBecomesAnOx(t *testing.T) {
	woeRRestlessEntersTapped(t, "Restless Bivouac", woeRBivouacOracle)
	g := effectstest.NewCatalogGame(t)
	land := effectstest.LciQRUZRestless(t, g, "Restless Bivouac", woeRBivouacOracle, 1, "R", "W")
	woeRAssertAnimated(t, g, land, "Ox", []string{"R", "W"}, 2, 2)
}

func TestWoeRRestlessBivouacAttackPutsACounterOnYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeRBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	land := effectstest.LciQRUZRestless(t, g, "Restless Bivouac", woeRBivouacOracle, 1, "R", "W")
	effectstest.DeclareAttack(t, g, opp.ID, land)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, theirs) || !effectstest.HasID(pick.PickTargetCards, bear) {
		t.Errorf("offered %v; want creatures you control only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters on the target, want 1", got)
	}
}

// --- Restless Cottage --------------------------------------------------------------

func TestWoeRRestlessCottageBecomesAHorror(t *testing.T) {
	woeRRestlessEntersTapped(t, "Restless Cottage", woeRCottageOracle)
	g := effectstest.NewCatalogGame(t)
	land := effectstest.LciQRUZRestless(t, g, "Restless Cottage", woeRCottageOracle, 2, "B", "G")
	woeRAssertAnimated(t, g, land, "Horror", []string{"B", "G"}, 4, 4)
}

func TestWoeRRestlessCottageAttackMakesFoodAndExilesAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	dead := effectstest.PushGraveyardCard(g, opp, "Old Bear")
	land := effectstest.LciQRUZRestless(t, g, "Restless Cottage", woeRCottageOracle, 2, "B", "G")
	effectstest.DeclareAttack(t, g, opp.ID, land)
	effectstest.PickTriggerTarget(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
	if opp.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the graveyard card was not exiled")
	}
}

// --- Restless Fortress -------------------------------------------------------------

func TestWoeRRestlessFortressBecomesANightmare(t *testing.T) {
	woeRRestlessEntersTapped(t, "Restless Fortress", woeRFortressOracle)
	g := effectstest.NewCatalogGame(t)
	land := effectstest.LciQRUZRestless(t, g, "Restless Fortress", woeRFortressOracle, 2, "W", "B")
	woeRAssertAnimated(t, g, land, "Nightmare", []string{"W", "B"}, 1, 4)
}

func TestWoeRRestlessFortressAttackDrainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	land := effectstest.LciQRUZRestless(t, g, "Restless Fortress", woeRFortressOracle, 2, "W", "B")
	mine, theirs := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-2 || me.Life != mine+2 {
		t.Errorf("life you %d → %d, defender %d → %d; want +2 and -2", mine, me.Life, theirs, opp.Life)
	}
}

// --- Restless Vinestalk ------------------------------------------------------------

func TestWoeRRestlessVinestalkBecomesAPlantWithTrample(t *testing.T) {
	woeRRestlessEntersTapped(t, "Restless Vinestalk", woeRVinestalkOracle)
	g := effectstest.NewCatalogGame(t)
	land := effectstest.LciQRUZRestless(t, g, "Restless Vinestalk", woeRVinestalkOracle, 3, "G", "U")
	woeRAssertAnimated(t, g, land, "Plant", []string{"G", "U"}, 5, 5)
	if !effectstest.HasEffectiveKeyword(t, g, land, "trample") {
		t.Error("the animated Vinestalk has no trample")
	}
	effectstest.EndTurn(t, g)
	if c := effectstest.S58p6View(t, g, land); c.IsCreature() || effectstest.HasEffectiveKeyword(t, g, land, "trample") {
		t.Error("the animation outlasted the turn")
	}
}

func TestWoeRRestlessVinestalkAttackMakesAnotherCreatureBaseThreeThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	land := effectstest.LciQRUZRestless(t, g, "Restless Vinestalk", woeRVinestalkOracle, 3, "G", "U")
	effectstest.DeclareAttack(t, g, opp.ID, land)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, land) || !effectstest.HasID(pick.PickTargetCards, big) {
		t.Errorf("offered %v; want another creature only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, big); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, big); p != 6 || tough != 6 {
		t.Errorf("the target is %d/%d on the next turn, want 6/6", p, tough)
	}
}

// --- Return from the Wilds ---------------------------------------------------------

func TestWoeRReturnFromTheWildsFetchesALandAndMakesAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	forest := effectstest.StackLibrary(me, "Forest", "Basic Land — Forest", "")
	effectstest.CastCatalogSpellWithModes(t, g, "Return from the Wilds", "Sorcery", woeRReturnOracle, []int{0, 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, forest); !ok || !c.Tapped {
		t.Errorf("the Forest is not on the battlefield tapped (on=%v)", ok)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Human"); got != 1 {
		t.Errorf("%d Humans, want 1", got)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food from an unchosen mode", got)
	}
}

func TestWoeRReturnFromTheWildsMakesAHumanAndAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.CastCatalogSpellWithModes(t, g, "Return from the Wilds", "Sorcery", woeRReturnOracle, []int{1, 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Human"); got != 1 {
		t.Errorf("%d Humans, want 1", got)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

// --- Rimefur Reindeer --------------------------------------------------------------

func TestWoeRRimefurReindeerTapsAnOpposingCreatureWhenYourEnchantmentEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rimefur Reindeer", "Creature — Elk", woeRReindeerOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", woeRBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	ench := effectstest.B12Permanent(g, me.ID, "My Anthem", "Enchantment")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: ench, Actor: me.ID})
	})
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt when your enchantment entered")
	}
	if effectstest.HasID(pick.PickTargetCards, mine) || !effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Errorf("offered %v; want creatures an opponent controls only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, theirs) {
		t.Error("the opposing creature was not tapped")
	}
}

func TestWoeRRimefurReindeerIgnoresAnOpponentsEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rimefur Reindeer", "Creature — Elk", woeRReindeerOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Anthem", "Enchantment")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: ench, Actor: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TappedNow(g, theirs) {
		t.Error("an opponent's enchantment triggered the Reindeer")
	}
}

// --- Rootrider Faun ----------------------------------------------------------------

func TestWoeRRootriderFaunTapsForGreenOrFiltersToAnyColor(t *testing.T) {
	spec, _ := effects.Lookup(woeRFaunOracle)
	if len(spec.ManaAbilities) != 2 ||
		spec.ManaAbilities[0].Produced != "{G}" || !spec.ManaAbilities[0].Cost.Tap ||
		spec.ManaAbilities[1].Produced != "{W|U|B|R|G}" || spec.ManaAbilities[1].Cost.Mana != "{1}" || !spec.ManaAbilities[1].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {G} and {1}, {T}: any color", spec.ManaAbilities)
	}
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	faun := effectstest.PushCatalogPermanent(g, me.ID, "Rootrider Faun", "Creature — Satyr Scout", woeRFaunOracle, false)
	effectstest.AdvanceToMain(t, g)
	before := me.ManaPool.CountColor("G")
	effectstest.ActivateManaFor(t, g, me.ID, faun, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("G"); got != before+1 {
		t.Errorf("green in the pool %d → %d, want one more", before, got)
	}
}

// --- Rowdy Research ----------------------------------------------------------------

func TestWoeRRowdyResearchCostsOneLessForEachCreatureThatAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	if got := woeILPrice(t, g, me.ID, "Rowdy Research", woeRResearchTypeLine, "{6}{U}", woeRResearchOracle, nil); got != 7 {
		t.Errorf("price with no attackers %d, want 7", got)
	}
	a := effectstest.B12Creature(g, me.ID, "A", woeRBearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", woeRBearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if got := woeILPrice(t, g, me.ID, "Rowdy Research", woeRResearchTypeLine, "{6}{U}", woeRResearchOracle, nil); got != 5 {
		t.Errorf("price after two attackers %d, want 5", got)
	}
}

func TestWoeRRowdyResearchDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Rowdy Research", woeRResearchTypeLine, woeRResearchOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want three cards drawn", hand, got)
	}
}

// --- Ruby, Daring Tracker ----------------------------------------------------------

func TestWoeRRubyGetsPlusTwoWhenAttackingBesideABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	ruby := effectstest.B12Push(g, me.ID, "Ruby, Daring Tracker", "Legendary Creature — Human Scout", woeRRubyOracle, 1, 2)
	effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, ruby, "haste") {
		t.Error("Ruby has no haste")
	}
	effectstest.DeclareAttack(t, g, opp.ID, ruby)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, ruby); p != 3 || tough != 4 {
		t.Errorf("Ruby is %d/%d, want 3/4", p, tough)
	}
}

func TestWoeRRubyGetsNothingWithoutABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	ruby := effectstest.B12Push(g, me.ID, "Ruby, Daring Tracker", "Legendary Creature — Human Scout", woeRRubyOracle, 1, 2)
	effectstest.B12Creature(g, me.ID, "My Bear", woeRBearType, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, ruby)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, ruby); p != 1 || tough != 2 {
		t.Errorf("Ruby is %d/%d, want 1/2", p, tough)
	}
}

func TestWoeRRubyTapsForRedOrGreen(t *testing.T) {
	spec, _ := effects.Lookup(woeRRubyOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {R} or {G}", spec.ManaAbilities)
	}
}
