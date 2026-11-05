package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b3_cards_test.go — Teenage Mutant Ninja Turtles, batch 3 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB3DonatelloOracle    = "194af2be-50c7-484b-bdd3-402bad70335b"
	tmtB3BeaversOracle      = "a1cec8e8-cf3c-42a4-b625-ca0cd148664e"
	tmtB3PointSquadOracle   = "2dea1935-c227-46d5-9d98-e7d2cd612f98"
	tmtB3EastWindOracle     = "e2ef355a-2472-47e0-9b2e-987994291921"
	tmtB3FilcherOracle      = "79a9fc1c-a6a9-483a-9ba7-d09fe41760c3"
	tmtB3FootEliteOracle    = "b6244b6c-664d-417a-bd58-ace5c7a995a2"
	tmtB3FootHQOracle       = "063dd25b-d4b4-4e09-acae-8c52dcb40803"
	tmtB3FootMysticOracle   = "2043fbde-48c4-4a77-8911-8991d77de1eb"
	tmtB3FrogButlerOracle   = "5b865275-4114-44d0-9832-ae5d3b4efba4"
	tmtB3TraagOracle        = "ade54bb7-3df8-4578-add8-1f85afee57ab"
	tmtB3GoNinjaGoOracle    = "46f77c8c-6817-4aea-a80a-d91f508ccbd0"
	tmtB3GroundedOracle     = "9a334434-072a-4486-810a-63efb29dea6d"
	tmtB3GuacPizzaOracle    = "d6d83a51-9d57-44a7-aae3-a96d43ec5ac8"
	tmtB3HamatoStanceOracle = "e08b19ca-f9d7-4c33-b37a-a77b37aabf1b"
)

func TestTmtB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB3DonatelloOracle:    "Donatello, Way with Machines",
		tmtB3BeaversOracle:      "Dream Beavers",
		tmtB3PointSquadOracle:   "EPF Point Squad",
		tmtB3EastWindOracle:     "East Wind Avatar",
		tmtB3FilcherOracle:      "Featherbrained Filcher",
		tmtB3FootEliteOracle:    "Foot Elite",
		tmtB3FootHQOracle:       "Foot Headquarters",
		tmtB3FootMysticOracle:   "Foot Mystic",
		tmtB3FrogButlerOracle:   "Frog Butler",
		tmtB3TraagOracle:        "General Traag, Heart of Stone",
		tmtB3GoNinjaGoOracle:    "Go Ninja Go",
		tmtB3GroundedOracle:     "Grounded for Life",
		tmtB3GuacPizzaOracle:    "Guac & Marshmallow Pizza",
		tmtB3HamatoStanceOracle: "Hamato Guardian Stance",
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

// tmtB3MakeToken creates one token for owner from the token table.
func tmtB3MakeToken(g *game.Game, owner uuid.UUID, key string) {
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(owner, TokenCard(key), 1) })
}

// --- Donatello, Way with Machines --------------------------------------------

func TestTmtB3DonatelloGrowsWhenYourArtifactsEnter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	don := b27Push(g, me.ID, "Donatello, Way with Machines", "Legendary Creature — Mutant Ninja Turtle", tmtB3DonatelloOracle, "{2}{U}", 1, 2, "U")
	assertKeywords(t, g, don, "flying")

	tmtB3MakeToken(g, me.ID, "1/1 colorless Robot artifact")
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, FoodToken(), 1) })
	passPriorityAroundTable(t, g)
	tmtB3MakeToken(g, opp.ID, "1/1 colorless Robot artifact")
	passPriorityAroundTable(t, g)
	tmtB3MakeToken(g, me.ID, "2/2 red Mutant")
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, don); got != 2 {
		t.Errorf("+1/+1 counters = %d after two of my artifacts, an opponent's and a nonartifact; want 2", got)
	}
}

// --- Dream Beavers -----------------------------------------------------------

func TestTmtB3DreamBeaversDrainsEachOpponentAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	start := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		start[p.ID] = p.Life
	}
	beavers := castCatalogSpell(t, g, "Dream Beavers", "Creature — Beaver Nightmare", tmtB3BeaversOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, beavers, "flying")
	if got := me.Life - start[me.ID]; got != 1 {
		t.Errorf("gained %d life, want 1", got)
	}
	for _, p := range g.Seats[1:] {
		if got := start[p.ID] - p.Life; got != 1 {
			t.Errorf("opponent %s lost %d life, want 1", p.ID, got)
		}
	}
	resolveAnyScryFor(t, g, me.ID)
}

// --- EPF Point Squad ---------------------------------------------------------

func TestTmtB3EPFPointSquadAllianceCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squad := b27Push(g, me.ID, "EPF Point Squad", "Creature — Human Soldier", tmtB3PointSquadOracle, "{1}{R/W}{R/W}", 2, 1, "R", "W")
	tmtB3MakeToken(g, opp.ID, "2/2 red Mutant")
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, squad); got != 0 {
		t.Fatalf("an opponent's creature entering put %d counters on the Squad", got)
	}
	tmtB3MakeToken(g, me.ID, "2/2 red Mutant")
	passPriorityAroundTable(t, g)
	tmtB3MakeToken(g, me.ID, "1/1 black Ninja")
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, squad); got != 2 {
		t.Errorf("+1/+1 counters = %d after two of my creatures entered, want 2", got)
	}
}

// --- East Wind Avatar --------------------------------------------------------

func TestTmtB3EastWindAvatarAlliancePumpsUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avatar := b27Push(g, me.ID, "East Wind Avatar", "Creature — Bird Spirit Avatar", tmtB3EastWindOracle, "{3}{W}", 2, 4, "W")
	assertKeywords(t, g, avatar, "flying", "vigilance")
	tmtB3MakeToken(g, opp.ID, "2/2 red Mutant")
	passPriorityAroundTable(t, g)
	tmtB3MakeToken(g, me.ID, "2/2 red Mutant")
	passPriorityAroundTable(t, g)
	tmtB3MakeToken(g, me.ID, "2/2 red Mutant")
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, avatar), effectiveToughness(t, g, avatar); p != 4 || tough != 4 {
		t.Fatalf("East Wind Avatar is %d/%d after two of my creatures entered, want 4/4", p, tough)
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if p := effectivePower(t, g, avatar); p != 2 {
		t.Errorf("the pump outlived the turn: power %d, want 2", p)
	}
}

// --- Featherbrained Filcher --------------------------------------------------

func TestTmtB3FeatherbrainedFilcherMakesAFoodWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	filcher := b27Push(g, me.ID, "Featherbrained Filcher", "Creature — Bird Mutant", tmtB3FilcherOracle, "{W}", 0, 2, "W")
	assertKeywords(t, g, filcher, "flying")
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(filcher) })
	passPriorityAroundTable(t, g)
	if got := onBattlefieldNamed(g, "Food"); got != 1 {
		t.Errorf("%d Food tokens after the Filcher left, want 1", got)
	}
}

// --- Foot Elite --------------------------------------------------------------

func TestTmtB3FootEliteAttackPumpsAndProtectsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elite := b27Push(g, me.ID, "Foot Elite", "Creature — Human Ninja", tmtB3FootEliteOracle, "{2}{W/B}", 2, 4, "W", "B")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, elite)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack trigger asks for another creature you control")
	}
	if hasID(p.PickTargetCards, elite) {
		t.Error("\"another\" excludes Foot Elite itself")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bear) != 3 || effectiveToughness(t, g, bear) != 2 || !effectiveAbilitiesContain(t, g, bear, "indestructible") {
		t.Errorf("the target: %d/%d, indestructible %v; want 3/2 and indestructible",
			effectivePower(t, g, bear), effectiveToughness(t, g, bear), effectiveAbilitiesContain(t, g, bear, "indestructible"))
	}
	if effectivePower(t, g, elite) != 2 || effectiveAbilitiesContain(t, g, elite, "indestructible") {
		t.Error("Foot Elite pumped itself")
	}
}

func TestTmtB3FootEliteCannotTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elite := b27Push(g, me.ID, "Foot Elite", "Creature — Human Ninja", tmtB3FootEliteOracle, "{2}{W/B}", 2, 4, "W", "B")
	declareAttack(t, g, opp.ID, elite)
	if triggerOnStack(g, elite) != nil || latestPickTarget(g, me.ID) != nil {
		t.Error("with no other creature, the trigger still went on the stack")
	}
}

// --- Foot Headquarters -------------------------------------------------------

func TestTmtB3FootHeadquartersEntersTappedGainsOneAndTapsForWOrB(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	hq := playLandFromHand(t, g, "Foot Headquarters", tmtB3FootHQOracle)
	top100AssertEnteredTapped(t, g, hq, "Foot Headquarters")
	if triggerOnStack(g, hq) == nil {
		t.Error("the life is a trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	spec, _ := Lookup(tmtB3FootHQOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|B}" {
		t.Errorf("mana abilities %+v, want one producing {W|B}", spec.ManaAbilities)
	}
}

// --- Foot Mystic -------------------------------------------------------------

func TestTmtB3FootMysticMakesANinjaOnlyWithDisappear(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mystic := castAndResolveCreature(t, g, "Foot Mystic", "Creature — Human Ninja Warlock", tmtB3FootMysticOracle)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, mystic, "lifelink")
	if n := onBattlefieldNamed(g, "Ninja"); n != 0 {
		t.Fatalf("no permanent left: %d Ninjas, want 0", n)
	}
	revoltFor(t, g, me.ID)
	castAndResolveCreature(t, g, "Foot Mystic", "Creature — Human Ninja Warlock", tmtB3FootMysticOracle)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Ninja"); n != 1 {
		t.Fatalf("disappear: %d Ninjas, want 1", n)
	}
	for _, id := range tmtB1Named(g, "Ninja") {
		c := tmtB1Card(t, g, id)
		if !c.IsToken() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" || !c.HasSubtype("Ninja") {
			t.Errorf("the token is %+v, want a 1/1 black Ninja creature token", c)
		}
	}
}

// --- Frog Butler -------------------------------------------------------------

func TestTmtB3FrogButlerMakesAnyColorAndBuysReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	frog := b27Push(g, me.ID, "Frog Butler", "Creature — Frog Spirit", tmtB3FrogButlerOracle, "{1}{G}", 1, 1, "G")
	assertKeywords(t, g, frog, "deathtouch")
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, frog, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := pendingOfKind(g, game.PendingChoiceMana); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool = %v, want [U]", got)
	}

	if effectiveAbilitiesContain(t, g, frog, "reach") {
		t.Fatal("Frog Butler has reach before the ability")
	}
	if err := p2Activate(t, g, me, frog, 0, "CC"); err != nil {
		t.Fatalf("activate {2}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, frog, "reach") {
		t.Error("Frog Butler did not gain reach")
	}
}

// --- General Traag, Heart of Stone -------------------------------------------

func castTmtB3Traag(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	traag := castCatalogSpell(t, g, "General Traag, Heart of Stone", "Legendary Artifact Creature — Elemental Soldier", tmtB3TraagOracle, nil)
	passPriorityAroundTable(t, g)
	return traag
}

func TestTmtB3GeneralTraagSacrificesAnArtifactForFourDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Wall", 0, 5)
	traag := castTmtB3Traag(t, g)
	assertKeywords(t, g, traag, "trample")

	answerOwnPermanents(t, g, me.ID, rock)
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact was not sacrificed before the reflexive trigger asked for a target")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := tmtB1Card(t, g, victim).DamageMarked; got != 4 {
		t.Errorf("damage marked = %d, want 4", got)
	}
}

func TestTmtB3GeneralTraagMayDeclineAndNeverOffersItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Wall", 0, 5)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	traag := castTmtB3Traag(t, g)

	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{traag}); err == nil && !g.Battlefield.Contains(traag) {
		t.Fatal("General Traag sacrificed himself as \"another artifact\"")
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		answerOwnPermanents(t, g, me.ID)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || tmtB1Card(t, g, victim).DamageMarked != 0 {
		t.Error("declining sacrificed something or dealt damage")
	}
}

// --- Go Ninja Go -------------------------------------------------------------

func TestTmtB3GoNinjaGoFlickersThenDealsGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	big := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(big, game.CounterPlusOne, 1) })
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)

	b02cCastModalSpell(t, g, "Go Ninja Go", "Sorcery", tmtB3GoNinjaGoOracle, []int{0, 1},
		[]game.TargetRef{p1g3ModeCard(big, 0), p1g3ModeCard(wall, 1)})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(big) {
		t.Fatal("the Ogre was not exiled (it is still the same object)")
	}
	ogres := tmtB1Named(g, "My Ogre")
	if len(ogres) != 1 {
		t.Fatalf("%d Ogres on the battlefield, want 1 returned", len(ogres))
	}
	if plusOnes(g, ogres[0]) != 0 || controllerOf(t, g, ogres[0]) != me.ID {
		t.Error("the returned Ogre kept its counter or came back under the wrong control")
	}
	// The counter left with the old object: the greatest power is now 4, not 5.
	if got := tmtB1Card(t, g, wall).DamageMarked; got != 4 {
		t.Errorf("damage marked = %d, want 4 (the greatest power after the flicker)", got)
	}
}

func TestTmtB3GoNinjaGoDamageAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 7, 7)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	b02cCastModalSpell(t, g, "Go Ninja Go", "Sorcery", tmtB3GoNinjaGoOracle, []int{1},
		[]game.TargetRef{p1g3ModeCard(wall, 0)})
	passPriorityAroundTable(t, g)
	if got := tmtB1Card(t, g, wall).DamageMarked; got != 5 {
		t.Errorf("damage marked = %d, want 5 (my greatest power, not an opponent's)", got)
	}
}

// --- Grounded for Life -------------------------------------------------------

func TestTmtB3GroundedForLifeCostsThreeLessAgainstATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { b16Tap(g, tapped) })
	advanceToMain(t, g)
	card := game.Card{Name: "Grounded for Life", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: tmtB3GroundedOracle}
	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(untapped)}); err == nil {
		t.Fatal("cast for two mana against an untapped creature")
	}
	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {1}{W} against a tapped creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !g.Battlefield.Contains(untapped) {
		t.Error("the tapped creature survived or the untapped one died")
	}
}

// --- Guac & Marshmallow Pizza ------------------------------------------------

func TestTmtB3GuacPizzaPumpsAndUntapsThenFeeds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { b16Tap(g, bear) })
	pizza := castCatalogSpell(t, g, "Guac & Marshmallow Pizza", "Artifact — Food", tmtB3GuacPizzaOracle, nil)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, pizza, "flash")
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if isTapped(g, bear) {
		t.Error("the target was not untapped")
	}

	life := me.Life
	if err := p2Activate(t, g, me, pizza, 0, "CC"); err != nil {
		t.Fatalf("activate the Food ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 || g.Battlefield.Contains(pizza) {
		t.Errorf("Food ability: life %d → %d, pizza still there %v; want +3 and gone", life, me.Life, g.Battlefield.Contains(pizza))
	}
}

// --- Hamato Guardian Stance --------------------------------------------------

func TestTmtB3HamatoGuardianStancePumpsGrantsFlyingAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Hamato Guardian Stance", "Instant", tmtB3HamatoStanceOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 5 {
		t.Errorf("the target is %d/%d, want 3/5", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	resolveAnyScryFor(t, g, me.ID)
}
