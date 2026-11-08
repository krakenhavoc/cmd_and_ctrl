package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_c_cards_test.go — Foundations, the C cards (#4465) that needed no
// new seam. Each test plays the real catalog entry through the engine
// and asserts what its text says.

const (
	fdnCProwlerOracle    = "898da62e-d29a-4ddc-be69-2ddec54614f5"
	fdnCGuideOracle      = "5b8e0913-47e9-4d56-971d-abd91b0f7587"
	fdnCCollectorOracle  = "52f1b5a4-62ff-44de-80da-503d4e52d312"
	fdnCRecruitOracle    = "07792d72-d5db-41b9-871c-e720e12ce659"
	fdnCInkmageOracle    = "8d979a0b-ba04-4abb-8f78-c707ce15351a"
	fdnCClawsOutOracle   = "499d76c1-97ff-4006-922c-be1684a7754c"
	fdnCSkymageOracle    = "d9a3207a-a330-4517-967e-f37c9a9f1355"
	fdnCConfiscateOracle = "d599a38c-7719-443e-bb59-55bd43a8fee6"
	fdnCGoblinOracle     = "c107e271-5741-4a84-bfc4-3bc322972d0d"
	fdnCCrusaderOracle   = "ea384b0d-3091-4d50-b15f-1f6763647b7c"
	fdnCFeasterOracle    = "4c22257e-8a00-40b7-a29b-55fcec7ddbd4"
	fdnCBarricadeOracle  = "291c7abf-9d64-4ba0-ae21-e3459edc9a95"
)

func TestFdnCEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnCProwlerOracle:    "Cackling Prowler",
		fdnCGuideOracle:      "Campus Guide",
		fdnCCollectorOracle:  "Cat Collector",
		fdnCRecruitOracle:    "Cemetery Recruitment",
		fdnCInkmageOracle:    "Cephalid Inkmage",
		fdnCClawsOutOracle:   "Claws Out",
		fdnCSkymageOracle:    "Clinquant Skymage",
		fdnCConfiscateOracle: "Confiscate",
		fdnCGoblinOracle:     "Courageous Goblin",
		fdnCCrusaderOracle:   "Crusader of Odric",
		fdnCFeasterOracle:    "Crypt Feaster",
		fdnCBarricadeOracle:  "Crystal Barricade",
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

func TestFdnCPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Clinquant Skymage", "Creature — Bird Wizard", fdnCSkymageOracle, 1, 1, []string{"flying"}},
		{"Crypt Feaster", "Creature — Zombie", fdnCFeasterOracle, 3, 4, []string{"menace"}},
		{"Crystal Barricade", "Artifact Creature — Wall", fdnCBarricadeOracle, 0, 4, []string{"defender"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keywords...)
	}
}

// fdnCCast casts a catalog creature from the active seat's hand in a main
// phase and resolves it and its entry trigger.
func fdnCCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID}
	castFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

// --- Cackling Prowler ------------------------------------------------------

func TestFdnCProwlerHasWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prowler := pushCatalogPermanent(g, me.ID, "Cackling Prowler", "Creature — Hyena Rogue", fdnCProwlerOracle, false)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, prowler)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting Cackling Prowler")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(prowler) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

func TestFdnCProwlerGrowsAtYourEndStepOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		prowler := blbB1Push(g, me.ID, "Cackling Prowler", "Creature — Hyena Rogue", fdnCProwlerOracle, 4, 3)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		if died {
			killCreature(t, g, opp.ID, bear)
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		want := 0
		if died {
			want = 1
		}
		if got := plusOnes(g, prowler); got != want {
			t.Errorf("a creature died %v: %d counters, want %d", died, got, want)
		}
	}
}

// --- Campus Guide ----------------------------------------------------------

func TestFdnCGuidePutsABasicLandOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Grizzly Bears", "Creature — Bear"),
		searchTestLand("Mystic Sanctuary", "Land — Island"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	fdnCCast(t, g, "Campus Guide", "Artifact Creature — Golem", fdnCGuideOracle, 2, 1)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	if searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil || searchOptionNamed(g, c, "Mystic Sanctuary") != uuid.Nil {
		t.Error("only basic land cards are legal finds")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	passPriorityAroundTable(t, g)
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.Name != "Forest" {
		t.Fatalf("the top card is %q, want Forest", top.Name)
	}
	if me.Library.Size() != 4 {
		t.Errorf("library = %d, the card never left it", me.Library.Size())
	}
	if names := revealedNames(revealsFor(g, opp.ID)); len(names) != 1 || names[0] != "Forest" {
		t.Errorf("the table was shown %v, want only Forest", names)
	}
}

// --- Cat Collector ---------------------------------------------------------

func TestFdnCCollectorMakesAFoodThenACatOnTheFirstGainOfYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	collector := fdnCCast(t, g, "Cat Collector", "Creature — Human Citizen", fdnCCollectorOracle, 3, 2)
	if got := tmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Fatalf("%d Foods after it entered, want 1", got)
	}
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(collector, me.ID, 2) })
		passPriorityAroundTable(t, g)
	}
	if got := tmtB1Count(g, me.ID, "Cat"); got != 1 {
		t.Errorf("%d Cats after two gains in one turn, want 1", got)
	}
}

func TestFdnCCollectorIgnoresAGainOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	collector := fdnCCast(t, g, "Cat Collector", "Creature — Human Citizen", fdnCCollectorOracle, 3, 2)
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(collector, me.ID, 2) })
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Cat"); got != 0 {
		t.Errorf("%d Cats from a gain on an opponent's turn, want 0", got)
	}
}

// --- Cemetery Recruitment --------------------------------------------------

func TestFdnCRecruitmentReturnsAndDrawsOnlyForAZombie(t *testing.T) {
	for _, zombie := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seedLibrary(me, "Top")
		typeLine := "Creature — Bear"
		if zombie {
			typeLine = "Creature — Zombie"
		}
		dead := pushGraveyardCardTyped(me, "Dead One", typeLine)
		hand := me.Hand.Size()
		castCatalogSpell(t, g, "Cemetery Recruitment", "Sorcery", fdnCRecruitOracle, cardRefs(dead))
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(dead) {
			t.Errorf("zombie %v: the creature card was not returned", zombie)
		}
		want := hand + 1
		if zombie {
			want = hand + 2
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("zombie %v: hand %d, want %d", zombie, got, want)
		}
	}
}

func TestFdnCRecruitmentRefusesANoncreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := pushGraveyardCardTyped(me, "Old Spell", "Sorcery")
	if err := castCatalogSpellErr(t, g, "Cemetery Recruitment", "Sorcery", fdnCRecruitOracle, cardRefs(spell)); err == nil {
		t.Error("a noncreature card was accepted as a target")
	}
}

// --- Cephalid Inkmage ------------------------------------------------------

func TestFdnCInkmageSurveilsThreeOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C", "D")
	fdnCCast(t, g, "Cephalid Inkmage", "Creature — Octopus Wizard", fdnCInkmageOracle, 2, 2)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger queued no surveil")
	}
	if len(c.ScryCards) != 3 {
		t.Errorf("surveil looked at %d cards, want 3", len(c.ScryCards))
	}
}

func TestFdnCInkmageCantBeBlockedWithThreshold(t *testing.T) {
	for _, threshold := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		inkmage := blbB1Push(g, me.ID, "Cephalid Inkmage", "Creature — Octopus Wizard", fdnCInkmageOracle, 2, 2)
		wall := blbB1Push(g, opp.ID, "Their Wall", "Creature — Wall", "", 0, 4)
		n := 6
		if threshold {
			n = 7
		}
		blbMPFillGraveyard(g, me, n)
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(inkmage, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(wall, inkmage)
		if threshold && err == nil {
			t.Error("blocked with threshold")
		}
		if !threshold && err != nil {
			t.Errorf("without threshold the block was refused: %v", err)
		}
	}
}

// --- Claws Out -------------------------------------------------------------

func TestFdnCClawsOutHasAffinityForCats(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if got := selfPricedMV(t, g, me, fdnCClawsOutOracle, "Instant", "{3}{W}{W}"); got != 5 {
		t.Errorf("no Cats: it costs %d, want 5", got)
	}
	blbB1Push(g, me.ID, "Cat", "Creature — Cat", "", 1, 1)
	blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if got := selfPricedMV(t, g, me, fdnCClawsOutOracle, "Instant", "{3}{W}{W}"); got != 4 {
		t.Errorf("one Cat: it costs %d, want 4", got)
	}
	for i := 0; i < 3; i++ {
		blbB1Push(g, me.ID, "Cat", "Creature — Cat", "", 1, 1)
	}
	if got := selfPricedMV(t, g, me, fdnCClawsOutOracle, "Instant", "{3}{W}{W}"); got != 2 {
		t.Errorf("four Cats: it costs %d, want 2 (only the generic part is reduced)", got)
	}
}

func TestFdnCClawsOutPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Claws Out", "Instant", fdnCClawsOutOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("your creature is %d/%d, want 4/4", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

// --- Clinquant Skymage -----------------------------------------------------

func TestFdnCSkymageGrowsOnEachOfYourDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	skymage := b12Push(g, me.ID, "Clinquant Skymage", "Creature — Bird Wizard", fdnCSkymageOracle, 1, 1)
	advanceToMain(t, g)
	for i := 0; i < 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if n := plusOnes(g, skymage); n != 2 {
		t.Fatalf("%d counters after two draws, want 2", n)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("opponent draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, skymage); n != 2 {
		t.Errorf("an opponent's draw grew it to %d counters", n)
	}
}

// --- Confiscate ------------------------------------------------------------

func TestFdnCConfiscateStealsANoncreaturePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castCatalogSpell(t, g, "Confiscate", auraTypeLine, fdnCConfiscateOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, rock); got != me.ID {
		t.Errorf("controller %s, want the Aura's controller %s", got, me.ID)
	}
}

// --- Courageous Goblin -----------------------------------------------------

func TestFdnCGoblinGetsPlusOneAndMenaceOnlyBesideABigCreature(t *testing.T) {
	for _, big := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		goblin := blbB1Push(g, me.ID, "Courageous Goblin", "Creature — Goblin", fdnCGoblinOracle, 2, 2)
		power := 3
		if big {
			power = 4
		}
		blbB1Push(g, me.ID, "Friend", "Creature — Giant", "", power, 4)
		declareAttack(t, g, opp.ID, goblin)
		passPriorityAroundTable(t, g)
		want := 2
		if big {
			want = 3
		}
		if got := effectivePower(t, g, goblin); got != want {
			t.Errorf("a %d-power friend: power %d, want %d", power, got, want)
		}
		if got := hasEffectiveKeyword(t, g, goblin, "menace"); got != big {
			t.Errorf("a %d-power friend: menace %v, want %v", power, got, big)
		}
	}
}

// --- Crusader of Odric -----------------------------------------------------

func TestFdnCCrusaderCountsTheCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	crusader := blbB1Push(g, me.ID, "Crusader of Odric", "Creature — Human Soldier", fdnCCrusaderOracle, 0, 0)
	if p, tough := ptOf(t, g, crusader); p != 1 || tough != 1 {
		t.Errorf("alone it is %d/%d, want 1/1", p, tough)
	}
	blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if p, tough := ptOf(t, g, crusader); p != 3 || tough != 3 {
		t.Errorf("with two more of yours it is %d/%d, want 3/3", p, tough)
	}
}

// --- Crypt Feaster ---------------------------------------------------------

func TestFdnCFeasterGetsPlusTwoWhenItAttacksWithThreshold(t *testing.T) {
	for _, threshold := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		feaster := blbB1Push(g, me.ID, "Crypt Feaster", "Creature — Zombie", fdnCFeasterOracle, 3, 4)
		n := 6
		if threshold {
			n = 7
		}
		blbMPFillGraveyard(g, me, n)
		declareAttack(t, g, opp.ID, feaster)
		passPriorityAroundTable(t, g)
		want := 3
		if threshold {
			want = 5
		}
		if got := effectivePower(t, g, feaster); got != want {
			t.Errorf("threshold %v: power %d, want %d", threshold, got, want)
		}
	}
}

// --- Crystal Barricade -----------------------------------------------------

func TestFdnCBarricadeGivesYouHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if playerHasHexproof(g, me) {
		t.Fatal("hexproof before the Barricade")
	}
	b12Push(g, me.ID, "Crystal Barricade", "Artifact Creature — Wall", fdnCBarricadeOracle, 0, 4)
	if !playerHasHexproof(g, me) {
		t.Error("its controller does not have hexproof")
	}
}

func TestFdnCBarricadePreventsNoncombatDamageToYourOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Push(g, me.ID, "Crystal Barricade", "Artifact Creature — Wall", fdnCBarricadeOracle, 0, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("noncombat damage to another creature you control was dealt")
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat, cardRefs(wall))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, wall); got != 3 {
		t.Errorf("the Barricade has %d damage, want 3: it protects only the others", got)
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature was protected")
	}
}
