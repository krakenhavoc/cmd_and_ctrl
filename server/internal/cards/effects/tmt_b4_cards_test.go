package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b4_cards_test.go — Teenage Mutant Ninja Turtles, batch 4 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB4JitteOracle        = "ab433d41-e801-405c-81f7-06fa6640797d"
	tmtB4HenchbotsOracle    = "1f097c91-6c2e-410f-8c8f-c021f39491af"
	tmtB4AceOracle          = "84a56740-8563-4ce5-8560-d44998d7f80d"
	tmtB4KittyOracle        = "f8131dca-ba6e-4ca4-943e-a3d26c75b8b8"
	tmtB4BusinessOracle     = "01ee70d4-6527-45ae-9600-bd3d16b3037d"
	tmtB4ArsenalOracle      = "4f02aa99-481c-4089-8f7c-cfe7eaa7dda0"
	tmtB4ExterminatorOracle = "3d081701-514b-4611-b5a1-239b8f335fb6"
	tmtB4JennikaOracle      = "b55da4d3-2cad-4c89-ad3f-7d8397b64d52"
	tmtB4LessonsOracle      = "51df0203-578d-47e1-a9f7-a13445f597d7"
	tmtB4DreggOracle        = "2db77e35-a54a-4e9c-8081-c97822515879"
	tmtB4MissileOracle      = "c596b847-4e05-4279-8675-b4e39cb3dc91"
	tmtB4CavalryOracle      = "2db8dda3-8136-4cf2-ba2d-ecc344a36725"
	tmtB4MetalheadOracle    = "7df45a2a-75bc-4343-8dad-3a642efd13ee"
	tmtB4MichelangeloOracle = "594c6b1f-ff0e-4122-8b29-852632a22a79"
)

func TestTmtB4EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB4JitteOracle:        "Hard-Won Jitte",
		tmtB4HenchbotsOracle:    "Henchbots",
		tmtB4AceOracle:          "High-Flying Ace",
		tmtB4KittyOracle:        "Ice Cream Kitty",
		tmtB4BusinessOracle:     "Illegitimate Business",
		tmtB4ArsenalOracle:      "Improvised Arsenal",
		tmtB4ExterminatorOracle: "Insectoid Exterminator",
		tmtB4JennikaOracle:      "Jennika, Bad Apple Big Sister",
		tmtB4LessonsOracle:      "Lessons from Life",
		tmtB4DreggOracle:        "Lord Dregg, Insect Invader",
		tmtB4MissileOracle:      "Manhole Missile",
		tmtB4CavalryOracle:      "Mechanized Ninja Cavalry",
		tmtB4MetalheadOracle:    "Metalhead",
		tmtB4MichelangeloOracle: "Michelangelo, Game Master",
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

// tmtB4Token makes one token for `owner` and returns its id.
func tmtB4Token(t *testing.T, g *game.Game, owner uuid.UUID, tmpl game.Card) uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			before[c.InstanceID] = true
		}
	})
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, tmpl, 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if !before[c.InstanceID] {
				id = c.InstanceID
			}
		}
	})
	if id == uuid.Nil {
		t.Fatal("no token was created")
	}
	return id
}

// tmtB4Tap taps a battlefield permanent.
func tmtB4Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

// --- Hard-Won Jitte -------------------------------------------------------------

func TestTmtB4HardWonJitteGivesTheEquippedCreatureDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	other := seedBear(g, me.ID)
	jitte := seedEquipment(g, me.ID, "Hard-Won Jitte", tmtB4JitteOracle)
	if effectiveAbilitiesContain(t, g, bear, "double strike") {
		t.Fatal("the bear has double strike before it is equipped")
	}
	equipTo(t, g, me.ID, jitte, bear)
	assertKeywords(t, g, bear, "double strike")
	if effectiveAbilitiesContain(t, g, other, "double strike") {
		t.Error("an unequipped creature has double strike")
	}
}

// --- Henchbots -------------------------------------------------------------------

func TestTmtB4HenchbotsExilesATappedCreatureUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := b12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	tmtB4Tap(g, tapped)

	bots := castCatalogSpell(t, g, "Henchbots", "Artifact Creature — Robot", tmtB4HenchbotsOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no Henchbots target prompt")
	}
	for _, id := range pick.PickTargetCards {
		if id == untapped {
			t.Error("an untapped creature was offered")
		}
	}
	pickCard(t, g, me.ID, tapped)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Fatal("the tapped creature was not exiled")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature left")
	}

	b27Kill(g, bots)
	g.RunStateChecksForTest()
	if tmtB1Count(g, opp.ID, "Tapped Bear") != 1 {
		t.Error("the exiled creature did not return when Henchbots left")
	}
}

func TestTmtB4HenchbotsHasNoTargetWithoutATappedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Tapped Bear", "Creature — Bear", 2, 2)
	tmtB4Tap(g, mine)
	castCatalogSpell(t, g, "Henchbots", "Artifact Creature — Robot", tmtB4HenchbotsOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("Henchbots asked for a target with no tapped creature an opponent controls")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own tapped creature was exiled")
	}
}

// --- High-Flying Ace -------------------------------------------------------------

func TestTmtB4HighFlyingAceGivesFlyingToACreatureWithoutItAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ace := b27Push(g, me.ID, "High-Flying Ace", "Creature — Bird Mutant", tmtB4AceOracle, "{2}{W}", 2, 3, "W")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, ace, "flying")

	spec, _ := Lookup(tmtB4AceOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Fatal("the activation is not timed as a sorcery")
	}
	advanceToMain(t, g)
	if err := hobB4Activate(t, g, me.ID, ace, 0, "{C}{C}{C}{W}", cardRefs(ace)); err == nil {
		t.Fatal("High-Flying Ace targeted a creature with flying")
	}
	if err := hobB4Activate(t, g, me.ID, ace, 0, "", cardRefs(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, bear, "flying")
}

// --- Ice Cream Kitty -------------------------------------------------------------

func TestTmtB4IceCreamKittySacrificesAnotherCreatureOrTokenToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	kitty := b27Push(g, me.ID, "Ice Cream Kitty", "Artifact Creature — Food Cat Mutant", tmtB4KittyOracle, "{1}{B/G}", 1, 3)
	rock := b12Permanent(g, me.ID, "Nontoken Rock", "Artifact")
	treasure := tmtB4Token(t, g, me.ID, TreasureToken())
	activate := func(sac uuid.UUID, mana string) error {
		if mana != "" {
			if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
		}
		return g.ActivateCatalogAbility(me.ID, kitty, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{sac}, Strict: true})
	}
	if err := activate(kitty, "{C}{C}"); err == nil {
		t.Fatal("Ice Cream Kitty sacrificed itself; the cost is another creature or token")
	}
	if err := activate(rock, ""); err == nil {
		t.Fatal("a nontoken noncreature artifact paid the cost")
	}
	hand := me.Hand.Size()
	if err := activate(treasure, ""); err != nil {
		t.Fatalf("a noncreature token could not pay the cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(treasure) {
		t.Error("the token was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

func TestTmtB4IceCreamKittyIsAFoodForThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kitty := b27Push(g, me.ID, "Ice Cream Kitty", "Artifact Creature — Food Cat Mutant", tmtB4KittyOracle, "{1}{B/G}", 1, 3)
	before := me.Life
	if err := hobB4Activate(t, g, me.ID, kitty, 1, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(kitty) {
		t.Error("the Kitty was not sacrificed")
	}
	if me.Life != before+3 {
		t.Errorf("life %d -> %d, want +3", before, me.Life)
	}
}

// --- Illegitimate Business ---------------------------------------------------------

func TestTmtB4IllegitimateBusinessIsAGolgariGainLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	id := playLandFromHand(t, g, "Illegitimate Business", tmtB4BusinessOracle)
	top100AssertEnteredTapped(t, g, id, "Illegitimate Business")
	if triggerOnStack(g, id) == nil {
		t.Error("the life is a trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d -> %d, want +1", before, me.Life)
	}
	spec, _ := Lookup(tmtB4BusinessOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B|G}" {
		t.Errorf("mana abilities %+v, want one producing {B|G}", spec.ManaAbilities)
	}
}

// --- Improvised Arsenal ------------------------------------------------------------

func TestTmtB4ImprovisedArsenalCountsArtifactsAndCopiesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	arsenal := seedEquipment(g, me.ID, "Improvised Arsenal", tmtB4ArsenalOracle)
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	base := effectivePower(t, g, bear)
	equipTo(t, g, me.ID, arsenal, bear)
	// The Arsenal and my Rock: +2/+0. The opponent's artifact does not count.
	if got := effectivePower(t, g, bear); got != base+2 {
		t.Fatalf("equipped power %d, want %d", got, base+2)
	}

	if err := hobB4Activate(t, g, me.ID, arsenal, 0, "{C}{C}{C}{C}{R}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	copies := tmtB1Named(g, "Improvised Arsenal")
	if len(copies) != 2 {
		t.Fatalf("%d Improvised Arsenals, want the original and a token copy", len(copies))
	}
	for _, id := range copies {
		if id == arsenal {
			continue
		}
		if c := tmtB1Card(t, g, id); !c.IsToken() || c.Controller != me.ID || c.AttachedTo.Kind == game.TargetCard {
			t.Errorf("the copy: token %v, controller %s, attached %v; want an unattached token of mine", c.IsToken(), c.Controller, c.AttachedTo)
		}
	}
	// The token is another artifact: +3/+0 now.
	if got := effectivePower(t, g, bear); got != base+3 {
		t.Errorf("power after the copy %d, want %d", got, base+3)
	}
}

// --- Insectoid Exterminator ----------------------------------------------------------

func TestTmtB4InsectoidExterminatorScriesOneWithDisappear(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bug := pushCatalogPermanent(g, me.ID, "Insectoid Exterminator", "Creature — Insect Mutant", tmtB4ExterminatorOracle, false)
	assertKeywords(t, g, bug, "flying")
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil || triggerOnStack(g, bug) != nil {
		t.Fatal("scried with no permanent gone this turn")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushCatalogPermanent(g2, me2.ID, "Insectoid Exterminator", "Creature — Insect Mutant", tmtB4ExterminatorOracle, false)
	revoltFor(t, g2, me2.ID)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if scryChoiceFor(g2, me2.ID) == nil {
		t.Error("no scry 1 at the end step after a permanent left")
	}
}

// --- Jennika, Bad Apple Big Sister -------------------------------------------------------

func TestTmtB4JennikaMakesAMutantAndPlainscycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Jennika, Bad Apple Big Sister", "Legendary Creature — Mutant Ninja Turtle", tmtB4JennikaOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Mutant"); got != 1 {
		t.Fatalf("%d Mutant tokens, want 1", got)
	}
	mutant := tmtB1Named(g, "Mutant")[0]
	if c := tmtB1Card(t, g, mutant); !c.IsToken() || c.Power != 2 || c.Toughness != 2 || len(c.Colors) != 1 || c.Colors[0] != "R" {
		t.Errorf("Mutant = %+v, want a 2/2 red token", c)
	}

	spec, _ := Lookup(tmtB4JennikaOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Zones == nil || spec.Activated[0].Zones[0] != game.ZoneHand {
		t.Fatalf("Jennika's activations %+v, want one plainscycling from the hand", spec.Activated)
	}
	if got := spec.Activated[0].Label; got[:len("Plainscycling {2}")] != "Plainscycling {2}" {
		t.Errorf("label %q, want Plainscycling {2}", got)
	}
}

// --- Lessons from Life ---------------------------------------------------------------

func TestTmtB4LessonsFromLifeDrawsThreeThenPutsALandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	castCatalogSpell(t, g, "Lessons from Life", "Sorcery", tmtB4LessonsOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land offer")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d -> %d before the land, want three cards drawn", hand, got)
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not enter")
	}
	if c := tmtB1Card(t, g, land); !c.Tapped {
		t.Error("the land entered untapped")
	}
}

// --- Lord Dregg, Insect Invader --------------------------------------------------------

func TestTmtB4LordDreggMakesAnInsectWarriorWithDisappear(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Lord Dregg, Insect Invader", "Legendary Creature — Insect Warrior", tmtB4DreggOracle, false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Insect Warrior"); got != 0 {
		t.Fatalf("%d Insect Warriors with no permanent gone, want 0", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushCatalogPermanent(g2, me2.ID, "Lord Dregg, Insect Invader", "Legendary Creature — Insect Warrior", tmtB4DreggOracle, false)
	revoltFor(t, g2, me2.ID)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	ids := tmtB1Named(g2, "Insect Warrior")
	if len(ids) != 1 {
		t.Fatalf("%d Insect Warriors after a permanent left, want 1", len(ids))
	}
	c := tmtB1Card(t, g2, ids[0])
	if !c.IsToken() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" {
		t.Errorf("token = %+v, want a 1/1 black token", c)
	}
	assertKeywords(t, g2, ids[0], "flying")
}

func TestTmtB4LordDreggSacrificesATokenToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dregg := pushCatalogPermanent(g, me.ID, "Lord Dregg, Insect Invader", "Legendary Creature — Insect Warrior", tmtB4DreggOracle, false)
	bear := b12Creature(g, me.ID, "Nontoken Bear", "Creature — Bear", 2, 2)
	treasure := tmtB4Token(t, g, me.ID, TreasureToken())
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, dregg, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Strict: true}); err == nil {
		t.Fatal("a nontoken creature paid \"Sacrifice a token\"")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, dregg, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(treasure) || me.Hand.Size() != hand+1 {
		t.Errorf("token still there %v, hand %d -> %d; want it gone and one card drawn", g.Battlefield.Contains(treasure), hand, me.Hand.Size())
	}
}

// --- Manhole Missile -----------------------------------------------------------------

func TestTmtB4ManholeMissileDealsThreeThenMayCycleACardFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	tuck := handCardForTest(me, "Spare Card", "Instant", "")
	castCatalogSpell(t, g, "Manhole Missile", "Instant", tmtB4MissileOracle, cardRefs(victim))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no offer to put a card on the bottom")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, tuck)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived 3 damage")
	}
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != tuck {
		t.Error("the card is not on the bottom of the library")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want one card out and one drawn", hand, got)
	}
}

func TestTmtB4ManholeMissileDrawsNothingWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	handCardForTest(me, "Spare Card", "Instant", "")
	castCatalogSpell(t, g, "Manhole Missile", "Instant", tmtB4MissileOracle, cardRefs(victim))
	hand := me.Hand.Size()
	library := me.Library.Size()
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	if me.Hand.Size() != hand || me.Library.Size() != library {
		t.Errorf("hand %d -> %d, library %d -> %d; want no change on a decline", hand, me.Hand.Size(), library, me.Library.Size())
	}
}

// --- Mechanized Ninja Cavalry ------------------------------------------------------------

func TestTmtB4MechanizedNinjaCavalryMakesARobot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Mechanized Ninja Cavalry", "Artifact Creature — Robot Ninja", tmtB4CavalryOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	ids := tmtB1Named(g, "Robot")
	if len(ids) != 1 {
		t.Fatalf("%d Robots, want 1", len(ids))
	}
	c := tmtB1Card(t, g, ids[0])
	if c.Controller != me.ID || !c.IsToken() || !c.IsArtifact() || !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 {
		t.Errorf("Robot = %+v, want my 1/1 colorless artifact creature token", c)
	}
}

// --- Metalhead -------------------------------------------------------------------------

func TestTmtB4MetalheadBouncesUpToOneOtherArtifactOrCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	enchantment := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	head := castCatalogSpell(t, g, "Metalhead", "Legendary Artifact Creature — Robot Turtle", tmtB4MetalheadOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no Metalhead target prompt")
	}
	for _, id := range pick.PickTargetCards {
		if id == head || id == enchantment {
			t.Errorf("%s was offered; the target is another artifact or creature", id)
		}
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(rock)...)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
}

func TestTmtB4MetalheadSacrificesAnotherArtifactToGrow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	head := b27Push(g, me.ID, "Metalhead", "Legendary Artifact Creature — Robot Turtle", tmtB4MetalheadOracle, "{4}{U}", 4, 4, "U")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, head, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{head}, Strict: true}); err == nil {
		t.Fatal("Metalhead sacrificed himself; the cost is another artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, head, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := plusOnes(g, head); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	assertKeywords(t, g, head, "menace", "haste")
}

// --- Michelangelo, Game Master -----------------------------------------------------------

func TestTmtB4MichelangeloGameMasterGrowsWithDisappear(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mikey := pushCatalogPermanent(g, me.ID, "Michelangelo, Game Master", "Legendary Creature — Mutant Ninja Turtle", tmtB4MichelangeloOracle, false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, mikey); got != 0 {
		t.Fatalf("%d counters with no permanent gone, want 0", got)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	mikey2 := pushCatalogPermanent(g2, me2.ID, "Michelangelo, Game Master", "Legendary Creature — Mutant Ninja Turtle", tmtB4MichelangeloOracle, false)
	revoltFor(t, g2, opp2.ID)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if got := plusOnes(g2, mikey2); got != 0 {
		t.Fatalf("an opponent's permanent leaving gave %d counters, want 0", got)
	}

	g3 := newCatalogGame(t)
	me3 := g3.Seats[0]
	mikey3 := pushCatalogPermanent(g3, me3.ID, "Michelangelo, Game Master", "Legendary Creature — Mutant Ninja Turtle", tmtB4MichelangeloOracle, false)
	revoltFor(t, g3, me3.ID)
	advanceTo(t, g3, game.StepEnd)
	passPriorityAroundTable(t, g3)
	if got := plusOnes(g3, mikey3); got != 1 {
		t.Errorf("%d counters after my permanent left, want 1", got)
	}
}
