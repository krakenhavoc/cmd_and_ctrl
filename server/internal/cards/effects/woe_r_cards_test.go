package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_r_cards_test.go — Wilds of Eldraine, the R batch (#4471). Each test
// plays the real catalog entry and asserts what its text says.

const (
	woeRMouseOracle      = "9ffe4d90-b46c-4606-89cb-1b724104f6ea"
	woeRTraineeOracle    = ratcatcherTraineeOracleID
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

// woeRAssertAnimated checks a Restless land's animated characteristics,
// its colours in any order.
func woeRAssertAnimated(t *testing.T, g *game.Game, land uuid.UUID, subtype string, colours []string, power, toughness int) {
	t.Helper()
	c := s58p6View(t, g, land)
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
	g := newCatalogGame(t)
	if id := b12PlayFromHand(t, g, name, "Land", oracle, game.CastSpellParams{}); !isTapped(g, id) {
		t.Errorf("%s entered untapped", name)
	}
}

// woeRPassUntilManaPick passes priority until chooser owes a colour pick,
// at most three times.
func woeRPassUntilManaPick(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 3 && riderLatestManaPick(g, chooser) == nil; i++ {
		passPriorityAroundTable(t, g)
	}
}

// --- Raging Battle Mouse -----------------------------------------------------------

func TestWoeRRagingBattleMouseDiscountsOnlyTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Raging Battle Mouse", "Creature — Mouse", woeRMouseOracle, false)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("first spell: %d, want 3 (undiscounted)", got)
	}
	b24CastNoncreature(t, g, "First", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 2 {
		t.Errorf("second spell: %d, want 2 ({1} less)", got)
	}
	b24CastNoncreature(t, g, "Second", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("third spell: %d, want 3 (undiscounted)", got)
	}
}

func TestWoeRRagingBattleMousePumpsATargetWhileCelebrating(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Raging Battle Mouse", "Creature — Mouse", woeRMouseOracle, false)
	b12Creature(g, me.ID, "Grizzly Bears", woeRBearType, 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	if latestPickTarget(g, me.ID) != nil || triggerOnStackFrom(g, "Raging Battle Mouse") {
		t.Fatal("the trigger fired with no celebration")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushCatalogPermanent(g2, me2.ID, "Raging Battle Mouse", "Creature — Mouse", woeRMouseOracle, false)
	bear := b12Creature(g2, me2.ID, "Grizzly Bears", woeRBearType, 2, 2)
	celebrationTokens(t, g2, me2.ID, 2)
	advanceTo(t, g2, game.StepBeginCombat)
	pickCard(t, g2, me2.ID, bear)
	passPriorityAroundTable(t, g2)
	if p, tough := effectivePTOf(t, g2, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	advanceToNextSeatsTurn(t, g2)
	if p, tough := effectivePTOf(t, g2, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear is %d/%d on the next turn, want 2/2", p, tough)
	}
}

// --- Ratcatcher Trainee // Pest Problem --------------------------------------------

func TestWoeRRatcatcherTraineeHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	trainee := b12Push(g, me.ID, "Ratcatcher Trainee", "Creature — Human Peasant", woeRTraineeOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, trainee, "first strike") {
		t.Error("no first strike on your turn")
	}
	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, trainee, "first strike") {
		t.Error("first strike on an opponent's turn")
	}
}

func TestWoeRPestProblemMakesTwoRatsThatCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	card := woeB1Adventure(me, woeRTraineeOracle,
		game.Face{Name: "Ratcatcher Trainee", TypeLine: "Creature — Human Peasant", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 2, Toughness: 1},
		game.Face{Name: "Pest Problem", TypeLine: "Instant — Adventure", ManaCost: "{2}{R}", Colors: []string{"R"}})
	woeB1CastAdventure(t, g, me, card, nil)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 2 {
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
		assertRestrictions(t, g, rat, game.CantBlock)
	}
}

// --- Rat Out ---------------------------------------------------------------------

func TestWoeRRatOutShrinksATargetAndMakesARat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	castCatalogSpell(t, g, "Rat Out", "Instant", woeRRatOutOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Errorf("%d Rats, want 1", got)
	}
}

func TestWoeRRatOutWithNoTargetStillMakesARat(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	castCatalogSpell(t, g, "Rat Out", "Instant", woeRRatOutOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Errorf("%d Rats, want 1", got)
	}
}

// --- Realm-Scorcher Hellkite --------------------------------------------------------

func TestWoeRRealmScorcherHellkiteBargainedAddsFourMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	token := woeILToken(g, me.ID)
	p1g2Cast(t, g, "Realm-Scorcher Hellkite", "Creature — Dragon", woeRHellkiteOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	woeRPassUntilManaPick(t, g, me.ID)
	if n := riderAnswerManaPicks(t, g, me.ID, "R"); n != 4 {
		t.Fatalf("answered %d colour picks, want 4", n)
	}
	if me.ManaPool.Total() != 4 {
		t.Errorf("pool has %d mana, want 4", me.ManaPool.Total())
	}
	hk := findBattlefieldByName(g, "Realm-Scorcher Hellkite")
	if hk == uuid.Nil {
		t.Fatal("the Hellkite is not on the battlefield")
	}
	for _, kw := range []string{"flying", "haste"} {
		if !hasEffectiveKeyword(t, g, hk, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestWoeRRealmScorcherHellkiteUnbargainedAddsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	castAndResolveCreature(t, g, "Realm-Scorcher Hellkite", "Creature — Dragon", woeRHellkiteOracle)
	passPriorityAroundTable(t, g)
	if riderLatestManaPick(g, me.ID) != nil || me.ManaPool.Total() != 0 {
		t.Errorf("an unbargained Hellkite added mana (pool %d)", me.ManaPool.Total())
	}
}

func TestWoeRRealmScorcherHellkitePingsAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	advanceToMain(t, g)
	hk := b12Push(g, me.ID, "Realm-Scorcher Hellkite", "Creature — Dragon", woeRHellkiteOracle, 4, 6)
	floatMana(t, g, me, "{R}{R}")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, hk, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
}

// --- Redcap Thief ------------------------------------------------------------------

func TestWoeRRedcapThiefMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	castAndResolveCreature(t, g, "Redcap Thief", "Creature — Goblin Rogue", woeRThiefOracle)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Regal Bunnicorn ---------------------------------------------------------------

func TestWoeRRegalBunnicornCountsNonlandPermanentsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bunny := b12Push(g, me.ID, "Regal Bunnicorn", "Creature — Rabbit Unicorn", woeRBunnicornOracle, 0, 0)
	if p, tough := effectivePTOf(t, g, bunny); p != 1 || tough != 1 {
		t.Errorf("alone it is %d/%d, want 1/1", p, tough)
	}
	b12Permanent(g, me.ID, "My Relic", "Artifact")
	b12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	b12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	if p, tough := effectivePTOf(t, g, bunny); p != 2 || tough != 2 {
		t.Errorf("with an artifact and a land it is %d/%d, want 2/2", p, tough)
	}
}

// --- Restless Bivouac --------------------------------------------------------------

func TestWoeRRestlessBivouacBecomesAnOx(t *testing.T) {
	woeRRestlessEntersTapped(t, "Restless Bivouac", woeRBivouacOracle)
	g := newCatalogGame(t)
	land := lciQRUZRestless(t, g, "Restless Bivouac", woeRBivouacOracle, 1, "R", "W")
	woeRAssertAnimated(t, g, land, "Ox", []string{"R", "W"}, 2, 2)
}

func TestWoeRRestlessBivouacAttackPutsACounterOnYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeRBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	land := lciQRUZRestless(t, g, "Restless Bivouac", woeRBivouacOracle, 1, "R", "W")
	declareAttack(t, g, opp.ID, land)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, theirs) || !hasID(pick.PickTargetCards, bear) {
		t.Errorf("offered %v; want creatures you control only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters on the target, want 1", got)
	}
}

// --- Restless Cottage --------------------------------------------------------------

func TestWoeRRestlessCottageBecomesAHorror(t *testing.T) {
	woeRRestlessEntersTapped(t, "Restless Cottage", woeRCottageOracle)
	g := newCatalogGame(t)
	land := lciQRUZRestless(t, g, "Restless Cottage", woeRCottageOracle, 2, "B", "G")
	woeRAssertAnimated(t, g, land, "Horror", []string{"B", "G"}, 4, 4)
}

func TestWoeRRestlessCottageAttackMakesFoodAndExilesAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	dead := pushGraveyardCard(g, opp, "Old Bear")
	land := lciQRUZRestless(t, g, "Restless Cottage", woeRCottageOracle, 2, "B", "G")
	declareAttack(t, g, opp.ID, land)
	pickTriggerTarget(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
	if opp.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the graveyard card was not exiled")
	}
}

// --- Restless Fortress -------------------------------------------------------------

func TestWoeRRestlessFortressBecomesANightmare(t *testing.T) {
	woeRRestlessEntersTapped(t, "Restless Fortress", woeRFortressOracle)
	g := newCatalogGame(t)
	land := lciQRUZRestless(t, g, "Restless Fortress", woeRFortressOracle, 2, "W", "B")
	woeRAssertAnimated(t, g, land, "Nightmare", []string{"W", "B"}, 1, 4)
}

func TestWoeRRestlessFortressAttackDrainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	land := lciQRUZRestless(t, g, "Restless Fortress", woeRFortressOracle, 2, "W", "B")
	mine, theirs := me.Life, opp.Life
	declareAttack(t, g, opp.ID, land)
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-2 || me.Life != mine+2 {
		t.Errorf("life you %d → %d, defender %d → %d; want +2 and -2", mine, me.Life, theirs, opp.Life)
	}
}

// --- Restless Vinestalk ------------------------------------------------------------

func TestWoeRRestlessVinestalkBecomesAPlantWithTrample(t *testing.T) {
	woeRRestlessEntersTapped(t, "Restless Vinestalk", woeRVinestalkOracle)
	g := newCatalogGame(t)
	land := lciQRUZRestless(t, g, "Restless Vinestalk", woeRVinestalkOracle, 3, "G", "U")
	woeRAssertAnimated(t, g, land, "Plant", []string{"G", "U"}, 5, 5)
	if !hasEffectiveKeyword(t, g, land, "trample") {
		t.Error("the animated Vinestalk has no trample")
	}
	endTurn(t, g)
	if c := s58p6View(t, g, land); c.IsCreature() || hasEffectiveKeyword(t, g, land, "trample") {
		t.Error("the animation outlasted the turn")
	}
}

func TestWoeRRestlessVinestalkAttackMakesAnotherCreatureBaseThreeThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	land := lciQRUZRestless(t, g, "Restless Vinestalk", woeRVinestalkOracle, 3, "G", "U")
	declareAttack(t, g, opp.ID, land)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, land) || !hasID(pick.PickTargetCards, big) {
		t.Errorf("offered %v; want another creature only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, big); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
	advanceToNextSeatsTurn(t, g)
	if p, tough := effectivePTOf(t, g, big); p != 6 || tough != 6 {
		t.Errorf("the target is %d/%d on the next turn, want 6/6", p, tough)
	}
}

// --- Return from the Wilds ---------------------------------------------------------

func TestWoeRReturnFromTheWildsFetchesALandAndMakesAHuman(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	forest := stackLibrary(me, "Forest", "Basic Land — Forest", "")
	castCatalogSpellWithModes(t, g, "Return from the Wilds", "Sorcery", woeRReturnOracle, []int{0, 1})
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, forest); !ok || !c.Tapped {
		t.Errorf("the Forest is not on the battlefield tapped (on=%v)", ok)
	}
	if got := b27CountNamed(g, me.ID, "Human"); got != 1 {
		t.Errorf("%d Humans, want 1", got)
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food from an unchosen mode", got)
	}
}

func TestWoeRReturnFromTheWildsMakesAHumanAndAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	castCatalogSpellWithModes(t, g, "Return from the Wilds", "Sorcery", woeRReturnOracle, []int{1, 2})
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Human"); got != 1 {
		t.Errorf("%d Humans, want 1", got)
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

// --- Rimefur Reindeer --------------------------------------------------------------

func TestWoeRRimefurReindeerTapsAnOpposingCreatureWhenYourEnchantmentEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Rimefur Reindeer", "Creature — Elk", woeRReindeerOracle, false)
	mine := b12Creature(g, me.ID, "My Bear", woeRBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	ench := b12Permanent(g, me.ID, "My Anthem", "Enchantment")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: ench, Actor: me.ID})
	})
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt when your enchantment entered")
	}
	if hasID(pick.PickTargetCards, mine) || !hasID(pick.PickTargetCards, theirs) {
		t.Errorf("offered %v; want creatures an opponent controls only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, theirs) {
		t.Error("the opposing creature was not tapped")
	}
}

func TestWoeRRimefurReindeerIgnoresAnOpponentsEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Rimefur Reindeer", "Creature — Elk", woeRReindeerOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeRBearType, 2, 2)
	ench := b12Permanent(g, opp.ID, "Their Anthem", "Enchantment")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: ench, Actor: opp.ID})
	})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || tappedNow(g, theirs) {
		t.Error("an opponent's enchantment triggered the Reindeer")
	}
}

// --- Rootrider Faun ----------------------------------------------------------------

func TestWoeRRootriderFaunTapsForGreenOrFiltersToAnyColor(t *testing.T) {
	spec, _ := Lookup(woeRFaunOracle)
	if len(spec.ManaAbilities) != 2 ||
		spec.ManaAbilities[0].Produced != "{G}" || !spec.ManaAbilities[0].Cost.Tap ||
		spec.ManaAbilities[1].Produced != "{W|U|B|R|G}" || spec.ManaAbilities[1].Cost.Mana != "{1}" || !spec.ManaAbilities[1].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {G} and {1}, {T}: any color", spec.ManaAbilities)
	}
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	faun := pushCatalogPermanent(g, me.ID, "Rootrider Faun", "Creature — Satyr Scout", woeRFaunOracle, false)
	advanceToMain(t, g)
	before := me.ManaPool.CountColor("G")
	activateManaFor(t, g, me.ID, faun, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("G"); got != before+1 {
		t.Errorf("green in the pool %d → %d, want one more", before, got)
	}
}

// --- Rowdy Research ----------------------------------------------------------------

func TestWoeRRowdyResearchCostsOneLessForEachCreatureThatAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	if got := woeILPrice(t, g, me.ID, "Rowdy Research", woeRResearchTypeLine, "{6}{U}", woeRResearchOracle, nil); got != 7 {
		t.Errorf("price with no attackers %d, want 7", got)
	}
	a := b12Creature(g, me.ID, "A", woeRBearType, 2, 2)
	b := b12Creature(g, me.ID, "B", woeRBearType, 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepPostcombatMain)
	if got := woeILPrice(t, g, me.ID, "Rowdy Research", woeRResearchTypeLine, "{6}{U}", woeRResearchOracle, nil); got != 5 {
		t.Errorf("price after two attackers %d, want 5", got)
	}
}

func TestWoeRRowdyResearchDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Rowdy Research", woeRResearchTypeLine, woeRResearchOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want three cards drawn", hand, got)
	}
}

// --- Ruby, Daring Tracker ----------------------------------------------------------

func TestWoeRRubyGetsPlusTwoWhenAttackingBesideABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	ruby := b12Push(g, me.ID, "Ruby, Daring Tracker", "Legendary Creature — Human Scout", woeRRubyOracle, 1, 2)
	b12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)
	if !hasEffectiveKeyword(t, g, ruby, "haste") {
		t.Error("Ruby has no haste")
	}
	declareAttack(t, g, opp.ID, ruby)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, ruby); p != 3 || tough != 4 {
		t.Errorf("Ruby is %d/%d, want 3/4", p, tough)
	}
}

func TestWoeRRubyGetsNothingWithoutABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	ruby := b12Push(g, me.ID, "Ruby, Daring Tracker", "Legendary Creature — Human Scout", woeRRubyOracle, 1, 2)
	b12Creature(g, me.ID, "My Bear", woeRBearType, 3, 3)
	declareAttack(t, g, opp.ID, ruby)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, ruby); p != 1 || tough != 2 {
		t.Errorf("Ruby is %d/%d, want 1/2", p, tough)
	}
}

func TestWoeRRubyTapsForRedOrGreen(t *testing.T) {
	spec, _ := Lookup(woeRRubyOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {R} or {G}", spec.ManaAbilities)
	}
}
