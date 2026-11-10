package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b2_test.go — Planechase plane cards, batches 2 and 3 (ADR 0216 slice
// 4, #4870). Every test plays the real catalog entry face up on a Planechase
// table (planes_harness_test.go).

var (
	planeNewNewYork       = planeFixture{"4f8cf1ab-1359-44af-a989-dd254fa6c7b5", "New New York", "Plane — New Earth"}
	planeNorthPole        = planeFixture{"a2dfbb3e-3d5f-4b6a-b910-4c7b0c9dc17a", "North Pole Research Base", "Plane — Earth"}
	planeOodSphere        = planeFixture{"8bf638e5-8325-48ca-9039-37a19761ef56", "Ood Sphere", "Plane — Horsehead Nebula"}
	planePompeii          = planeFixture{"5dc1deab-ce04-45c0-a4ff-c526a948e500", "Pompeii", "Plane — Earth"}
	planeCabinetRoom      = planeFixture{"b3acbacd-7e97-4bcf-9281-fffa5ea3e70b", "Prime Minister's Cabinet Room", "Plane — Earth"}
	planeSpectroxMines    = planeFixture{"4f387b4c-3e4a-419d-9dbe-95c5f797747e", "Spectrox Mines", "Plane — Androzani Minor"}
	planeStormcage        = planeFixture{"204ee01c-74ca-4c4c-92db-aebeebb89406", "Stormcage Containment Facility", "Plane — Unknown Planet"}
	planeTardisBay        = planeFixture{"424b2702-e7c6-4056-8466-a43782d7df2b", "TARDIS Bay", "Plane — Gallifrey"}
	planeCaveOfSkulls     = planeFixture{"6ad7027a-5f13-4833-bb88-879ba8830720", "The Cave of Skulls", "Plane — Earth"}
	planeCheetahPlanet    = planeFixture{"0c4679a1-f92f-44a9-9a44-066a137c8cd7", "The Cheetah Planet", "Plane — Outside Mutter's Spiral"}
	planeDiningCar        = planeFixture{"7099883d-0a4b-47ae-bff8-a96efa7a3a9a", "The Dining Car", "Plane — Spacecraft"}
	planeChildhoodBarn    = planeFixture{"2a21b90a-7ea7-4090-82ec-359ba8ba1aad", "The Doctor's Childhood Barn", "Plane — Gallifrey"}
	planeDoctorsTomb      = planeFixture{"fc621e94-e7c7-49ea-9547-4ab041ea39f3", "The Doctor's Tomb", "Plane — Trenzalore"}
	planeTheDrum          = planeFixture{"9f7e1607-34b5-4332-81e3-391fb49cd6ad", "The Drum, Mining Facility", "Plane — Earth"}
	planeLuxLibrary       = planeFixture{"68750b81-1309-4362-84b3-8835b3d71d28", "The Lux Foundation Library", "Plane — The Library"}
	planeMatrixOfTime     = planeFixture{"ed44d47b-2eea-4cb2-a091-07583bdf390f", "The Matrix of Time", "Plane — Gallifrey"}
	planeMoonbase         = planeFixture{"381f510b-8696-4af1-a6c9-d1175e9f36d0", "The Moonbase", "Plane — Moon"}
	planePyramidOfMars    = planeFixture{"5910ae7e-7a74-4296-9ae1-7fa093b4e640", "The Pyramid of Mars", "Plane — Mars"}
	planeTwoStreams       = planeFixture{"d8aa7f2d-2280-42d4-b082-1b5d45df77c5", "Two Streams Facility", "Plane — Apalapucia"}
	planeUnitHQ           = planeFixture{"a7641877-1a61-467c-9365-487cf5a5656d", "UNIT Headquarters", "Plane — Earth"}
	planeBloodhillBastion = planeFixture{"ce8139b9-31f3-464d-86ea-7efbdbfce478", "Bloodhill Bastion", "Plane — Equilor"}
	planeEnigmaRidges     = planeFixture{"93b3acc1-2d41-4f5f-9225-7a30ea2d8158", "Enigma Ridges", "Plane — Echoir"}
	planeEsper            = planeFixture{"4eebfaf3-18da-4661-b6a5-d874585e170a", "Esper", "Plane — Alara"}
	planeGavony           = planeFixture{"2387edac-765b-4502-bd3d-88be8ba7d0a4", "Gavony", "Plane — Innistrad"}
	planeGhirapur         = planeFixture{"a7c6b72c-f9ec-4048-8e4c-99210e4e7b85", "Ghirapur", "Plane — Avishkar"}
	planeGlimmervoidBasin = planeFixture{"43a23de7-0738-4b03-b87d-5a7d1144825c", "Glimmervoid Basin", "Plane — Mirrodin"}
)

// planesB2Count counts the permanents named `name` that `owner` controls.
func planesB2Count(g *game.Game, name string, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			n++
		}
	}
	return n
}

// planesB2Life is a player's life total.
func planesB2Life(p *game.Player) int { return p.Life }

// --- Spectrox Mines ----------------------------------------------------------

func TestSpectroxMinesCostsLifeForATreasureOnPlaneswalkAndUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeSpectroxMines)
	g, _ := newPlaneGame(t, planeSpectroxMines)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	before := planesB2Life(me)
	planeswalkToPlane(t, g, planeSpectroxMines.Oracle)
	passPriorityAroundTable(t, g)
	if got := planesB2Life(me); got != before-3 {
		t.Errorf("life %d after planeswalking in, want %d", got, before-3)
	}
	if n := planesB2Count(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	next := toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if n := planesB2Count(g, "Treasure", next.ID); n != 1 {
		t.Errorf("the new planar controller has %d Treasures after upkeep, want 1", n)
	}
}

func TestSpectroxMinesChaosMakesAFoodAndAHumanRogue(t *testing.T) {
	g, plane := newPlaneGame(t, planeSpectroxMines)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatal("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if planesB2Count(g, "Food", me.ID) != 1 || planesB2Count(g, "Human Rogue", me.ID) != 1 {
		t.Errorf("Food %d, Human Rogue %d, want 1 each", planesB2Count(g, "Food", me.ID), planesB2Count(g, "Human Rogue", me.ID))
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	rollChaos(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Error("a face-down plane's chaos ability triggered")
	}
}

// --- UNIT Headquarters -------------------------------------------------------

func TestUnitHeadquartersMakesATappedSoldierAndChaosGrowsYourCreatures(t *testing.T) {
	requirePlaneRegisteredFull(t, planeUnitHQ)
	g, _ := newPlaneGame(t, planeUnitHQ)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	planeswalkToPlane(t, g, planeUnitHQ.Oracle)
	passPriorityAroundTable(t, g)
	var soldier game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && c.Controller == me.ID {
			soldier = c
		}
	}
	if soldier.InstanceID == uuid.Nil || !soldier.Tapped {
		t.Fatalf("no tapped Soldier token after the planeswalk (%+v)", soldier)
	}
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Wolf", "Creature — Wolf", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p := tmtB1Card(t, g, mine).CurrentPower(); p != 3 {
		t.Errorf("my creature has power %d, want 3", p)
	}
	if p := tmtB1Card(t, g, theirs).CurrentPower(); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	if p := tmtB1Card(t, g, soldier.InstanceID).CurrentPower(); p != 2 {
		t.Errorf("the Soldier has power %d, want 2", p)
	}
}

// --- The Pyramid of Mars -----------------------------------------------------

func TestPyramidOfMarsSurveilsAndChaosReanimates(t *testing.T) {
	requirePlaneRegisteredFull(t, planePyramidOfMars)
	g, _ := newPlaneGame(t, planePyramidOfMars)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	relic := pushGraveyardCardWithTypeLineID(me, "Old Relic", "Artifact")
	planeswalkToPlane(t, g, planePyramidOfMars.Oracle)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("planeswalking in did not start a surveil")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(relic)); err == nil {
		t.Fatal("a noncreature card was accepted as the target")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if planesB2Count(g, "Bear", me.ID) != 1 {
		t.Error("the creature card did not return to the battlefield")
	}
}

// planesB2Keyword reports whether the battlefield card has the keyword now,
// after the layer pass.
func planesB2Keyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := layeredCard(t, g, id)
	return game.HasKeyword(&c, kw)
}

// --- Gavony ------------------------------------------------------------------

func TestGavonyGivesEveryCreatureVigilanceAndChaosGuardsYours(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGavony)
	g, plane := newPlaneGame(t, planeGavony)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, other.ID, "Wolf", "Creature — Wolf", 2, 2)
	if !planesB2Keyword(t, g, mine, "vigilance") || !planesB2Keyword(t, g, theirs, "vigilance") {
		t.Error("not every creature has vigilance")
	}
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatal("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !planesB2Keyword(t, g, mine, "indestructible") {
		t.Error("my creature is not indestructible")
	}
	if planesB2Keyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature is indestructible")
	}
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if planesB2Keyword(t, g, late, "indestructible") {
		t.Error("a creature that arrived after the ability resolved is indestructible")
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if planesB2Keyword(t, g, theirs, "vigilance") {
		t.Error("vigilance outlived the plane")
	}
}

// --- Bloodhill Bastion -------------------------------------------------------

func TestBloodhillBastionGivesEnteringCreaturesDoubleStrikeAndHaste(t *testing.T) {
	requirePlaneRegisteredFull(t, planeBloodhillBastion)
	g, _ := newPlaneGame(t, planeBloodhillBastion)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	var bear game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Test Bear" && c.Controller == me.ID {
			bear = c
		}
	}
	if bear.InstanceID == uuid.Nil {
		t.Fatal("the creature spell did not resolve")
	}
	if !planesB2Keyword(t, g, bear.InstanceID, "double strike") || !planesB2Keyword(t, g, bear.InstanceID, "haste") {
		t.Error("the entering creature lacks double strike or haste")
	}
}

func TestBloodhillBastionChaosBlinksANontokenCreature(t *testing.T) {
	g, _ := newPlaneGame(t, planeBloodhillBastion)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	token := TokenCard("1/1 white Soldier")
	token.InstanceID = uuid.New()
	token.Owner, token.Controller = me.ID, me.ID
	token.Name = "Soldier"
	g.WithWriteLock(func() { g.Battlefield.PushTop(token) })
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(token.InstanceID)); err == nil {
		t.Fatal("a token was accepted as the target")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := planesB2Count(g, "Bear", me.ID); n != 1 {
		t.Fatalf("%d Bears after the blink, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.InstanceID == bear {
			t.Error("the Bear is the same object: it was not exiled and returned")
		}
	}
}

// --- The Lux Foundation Library ----------------------------------------------

func TestLuxFoundationLibraryChaosPutsAShadowCounterOnAnyCreature(t *testing.T) {
	requirePlaneRegisteredFull(t, planeLuxLibrary)
	g, _ := newPlaneGame(t, planeLuxLibrary)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := b12Creature(g, other.ID, "Wolf", "Creature — Wolf", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if n := tmtB1Card(t, g, theirs).Counters[game.CounterShadow]; n != 1 {
		t.Errorf("%d shadow counters, want 1", n)
	}
	if !planesB2Keyword(t, g, theirs, "shadow") {
		t.Error("the creature does not have shadow")
	}
}

func TestLuxFoundationLibraryHasNoMaximumHandSize(t *testing.T) {
	g, _ := newPlaneGame(t, planeLuxLibrary)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var max int
	g.WithWriteLock(func() { max = g.EffectiveMaxHandSizeLocked(me) })
	if max != game.NoMaxHandSize {
		t.Errorf("maximum hand size %d under the plane, want none", max)
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	g.WithWriteLock(func() { max = g.EffectiveMaxHandSizeLocked(me) })
	if max == game.NoMaxHandSize {
		t.Error("no maximum hand size after planeswalking away")
	}
}

// --- Stormcage Containment Facility ------------------------------------------

func TestStormcageGivesCreatureCardsEscapeAndChaosDetains(t *testing.T) {
	requirePlaneRegisteredFull(t, planeStormcage)
	g, _ := newPlaneGame(t, planeStormcage)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := pushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	bolt := pushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
	if p := grantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard); !p.Granted() || p.AltCostKey != "escape" || p.ExileOtherFromGraveyard != 3 {
		t.Errorf("a creature card has no escape: %+v", p)
	}
	if p := grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard); p.Granted() {
		t.Error("an instant card has escape")
	}
	if p := grantedPermissionOn(g, other.ID, bear, game.ZoneGraveyard); p.Granted() {
		t.Error("another player's cast of my card has escape")
	}
	theirs := b12Creature(g, other.ID, "Wolf", "Creature — Wolf", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(mine)); err == nil {
		t.Fatal("my own creature was accepted as the target")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if r := avB13Restrictions(g, theirs); r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Errorf("the target is not detained: %v", r)
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if p := grantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard); p.Granted() {
		t.Error("escape outlived the plane")
	}
}

// --- North Pole Research Base ------------------------------------------------

func TestNorthPoleResearchBaseGivesAnOpponentACardAndATreasureAtYourUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeNorthPole)
	g, _ := newPlaneGame(t, planeNorthPole)
	next := toNextUpkeep(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	before := opp.Hand.Size()
	if asked := advanceAnsweringPlayerTargets(t, g, game.StepDraw, opp.ID); asked != 1 {
		t.Fatalf("the upkeep ability asked for a target %d times, want once", asked)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before+1 {
		t.Errorf("the opponent's hand is %d, want %d", got, before+1)
	}
	if n := planesB2Count(g, "Treasure", opp.ID); n != 1 {
		t.Errorf("the opponent has %d Treasures, want 1", n)
	}
	_ = next
}

func TestNorthPoleResearchBaseChaosMakesAnAlienAndStunsAnOpponentsCreature(t *testing.T) {
	g, _ := newPlaneGame(t, planeNorthPole)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := planesB2Count(g, "Alien", me.ID); n != 1 {
		t.Fatalf("%d Aliens after the chaos ability resolved, want 1", n)
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, theirs)
	if !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("the target is tapped %v with %d stun counters, want tapped with 1", c.Tapped, c.Counters[game.CounterStun])
	}
}

// --- New New York ------------------------------------------------------------

func TestNewNewYorkMakesNoncreatureArtifactsVehiclesAtCombat(t *testing.T) {
	requirePlaneRegisteredFull(t, planeNewNewYork)
	g, _ := newPlaneGame(t, planeNewNewYork)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	golem := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 1, 1)
	theirs := b12Permanent(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Rock", "Artifact")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, rock)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c = layeredCard(t, g, rock)
	if !c.HasSubtypeWord("Vehicle") || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the Rock is %d/%d with subtypes %q, want a 3/3 Vehicle", c.CurrentPower(), c.CurrentToughness(), c.Effective().Subtypes)
	}
	if !planesB2Keyword(t, g, rock, "flying") || !planesB2Keyword(t, g, rock, "haste") {
		t.Error("the Rock lacks flying or haste")
	}
	if layeredCard(t, g, golem).HasSubtypeWord("Vehicle") {
		t.Error("an artifact creature became a Vehicle")
	}
	if layeredCard(t, g, theirs).HasSubtypeWord("Vehicle") {
		t.Error("an opponent's artifact became a Vehicle")
	}
	endTurn(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if layeredCard(t, g, rock).HasSubtypeWord("Vehicle") {
		t.Error("the Rock is still a Vehicle after the turn")
	}
}

func TestNewNewYorkChaosMakesATreasureAndAnAlien(t *testing.T) {
	g, _ := newPlaneGame(t, planeNewNewYork)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if planesB2Count(g, "Treasure", me.ID) != 1 || planesB2Count(g, "Alien", me.ID) != 1 {
		t.Errorf("Treasures %d, Aliens %d, want 1 each", planesB2Count(g, "Treasure", me.ID), planesB2Count(g, "Alien", me.ID))
	}
}

// --- Ghirapur ----------------------------------------------------------------

func TestGhirapurMakesPlainArtifactsFiveThreeVehicles(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGhirapur)
	g, _ := newPlaneGame(t, planeGhirapur)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	cart := b12Permanent(g, me.ID, "Cart", "Artifact — Vehicle")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := layeredCard(t, g, rock)
	if !c.HasSubtypeWord("Vehicle") || c.CurrentPower() != 5 || c.CurrentToughness() != 3 {
		t.Errorf("the Rock is %d/%d, want a 5/3 Vehicle", c.CurrentPower(), c.CurrentToughness())
	}
	if !planesB2Keyword(t, g, rock, "trample") || !planesB2Keyword(t, g, rock, "haste") {
		t.Error("the Rock lacks trample or haste")
	}
	if planesB2Keyword(t, g, cart, "trample") {
		t.Error("a Vehicle that was already one was given trample")
	}
}

func TestGhirapurChaosReturnsANoncreatureArtifactCard(t *testing.T) {
	g, _ := newPlaneGame(t, planeGhirapur)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	golem := pushGraveyardCardWithTypeLineID(me, "Golem", "Artifact Creature — Golem")
	rock := pushGraveyardCardWithTypeLineID(me, "Rock", "Artifact")
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(golem)); err == nil {
		t.Fatal("an artifact creature card was accepted as the target")
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact card did not return to hand")
	}
}

// --- Esper -------------------------------------------------------------------

func TestEsperMakesArtifactSpellsCostOneLess(t *testing.T) {
	requirePlaneRegisteredFull(t, planeEsper)
	g, _ := newPlaneGame(t, planeEsper)
	advanceToMain(t, g)
	me, them := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	if got := priceInHand(t, g, me, "My Chalice", "Artifact", "{2}"); got != 1 {
		t.Errorf("the planar controller's artifact spell costs %d, want 1", got)
	}
	if got := priceInHand(t, g, them, "Their Chalice", "Artifact", "{2}"); got != 1 {
		t.Errorf("an opponent's artifact spell costs %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "My Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("a creature spell costs %d, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if got := priceInHand(t, g, me, "My Chalice", "Artifact", "{2}"); got != 2 {
		t.Errorf("the artifact spell costs %d after the plane left, want 2", got)
	}
}

func TestEsperChaosMakesWhiteBlueBlackCreaturesArtifactsAndGivesThemKeywords(t *testing.T) {
	g, _ := newPlaneGame(t, planeEsper)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	w := b12Creature(g, me.ID, "White Bear", "Creature — Bear", 2, 2)
	r := b12Creature(g, me.ID, "Red Bear", "Creature — Bear", 2, 2)
	golem := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their White Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case w, theirs:
				g.Battlefield.Cards[i].Colors = []string{"W"}
			case r:
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	g.BumpLayerVersionForTest()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !layeredCard(t, g, w).IsArtifact() {
		t.Error("the white creature did not become an artifact")
	}
	if layeredCard(t, g, r).IsArtifact() || layeredCard(t, g, theirs).IsArtifact() {
		t.Error("a red creature or an opponent's creature became an artifact")
	}
	for _, id := range []uuid.UUID{w, golem} {
		for _, kw := range []string{"vigilance", "menace", "lifelink"} {
			if !planesB2Keyword(t, g, id, kw) {
				t.Errorf("an artifact creature lacks %s", kw)
			}
		}
	}
	if planesB2Keyword(t, g, r, "lifelink") {
		t.Error("a non-artifact creature gained lifelink")
	}
}

// --- Enigma Ridges -----------------------------------------------------------

func TestEnigmaRidgesFetchesBasicsForThePlayersBehindOnLands(t *testing.T) {
	requirePlaneRegisteredFull(t, planeEnigmaRidges)
	g, _ := newPlaneGame(t, planeEnigmaRidges)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	for range 3 {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	for range 4 {
		pushLibraryCardForTest(opp, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	}
	handBefore := opp.Hand.Size()
	planeswalkToPlane(t, g, planeEnigmaRidges.Oracle)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the player behind on lands was not asked to search")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("the player with the most lands was asked to search")
	}
	if len(c.SearchCards) < 2 {
		t.Fatalf("the search offers %d cards", len(c.SearchCards))
	}
	beginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, opp.ID, c.SearchCards[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if got := opp.Hand.Size(); got != handBefore+2 {
		t.Errorf("the opponent's hand grew by %d, want 2 (the difference of lands)", got-handBefore)
	}
}

func TestEnigmaRidgesChaosDrawsAndMayPutALandOntoTheBattlefield(t *testing.T) {
	g, _ := newPlaneGame(t, planeEnigmaRidges)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: land, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	})
	before := me.Hand.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after chaos, want %d (one card drawn)", got, before+1)
	}
	p := latestPutFromHandChoice(g, me.ID)
	if p == nil {
		t.Fatal("no prompt for the land from hand")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if planesB2Count(g, "Mountain", me.ID) != 1 {
		t.Error("the land did not enter the battlefield")
	}
}

// --- Pompeii -----------------------------------------------------------------

func TestPompeiiBuildsEruptionCountersAndBlankRollsScry(t *testing.T) {
	requirePlaneRegisteredFull(t, planePompeii)
	g, _ := newPlaneGame(t, planePompeii)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	planeswalkToPlane(t, g, planePompeii.Oracle)
	passPriorityAroundTable(t, g)
	plane, _ := faceUpPlane(g, planePompeii.Oracle)
	if n := plane.Counters["eruption"]; n != 1 {
		t.Fatalf("%d eruption counters after planeswalking in, want 1", n)
	}
	rollBlank(t, g, me.ID)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	sc := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if sc == nil {
		t.Fatal("a blank on the planar die did not scry")
	}
	if err := g.ResolveScry(sc.ID, me.ID, nil, sc.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	plane, _ = faceUpPlane(g, planePompeii.Oracle)
	if n := plane.Counters["eruption"]; n != 2 {
		t.Errorf("%d eruption counters after a blank, want 2", n)
	}
	rollBlank(t, g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) != nil {
		t.Error("another player's blank roll triggered the plane")
	}
}

func TestPompeiiChaosErupts(t *testing.T) {
	g, _ := newPlaneGame(t, planePompeii)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	planeswalkToPlane(t, g, planePompeii.Oracle)
	passPriorityAroundTable(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	island := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	myLife, oppLife := me.Life, opp.Life
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Life != myLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life %d and %d, want %d and %d (one eruption counter)", me.Life, opp.Life, myLife-1, oppLife-1)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 1/1 survived the eruption")
	}
	if !g.Battlefield.Contains(giant) {
		t.Error("the 4/4 died to one damage")
	}
	answerSacrifice(t, g, me.ID, forest)
	answerSacrifice(t, g, opp.ID, island)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || g.Battlefield.Contains(island) {
		t.Error("a player kept their land")
	}
	if _, ok := faceUpPlane(g, planePompeii.Oracle); ok {
		t.Error("the table did not planeswalk after the sacrifices")
	}
}

// --- TARDIS Bay --------------------------------------------------------------

func TestTardisBayGivesCascadeToTheFirstSpellWithManaValueTwoOrMore(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTardisBay)
	g, _ := newPlaneGame(t, planeTardisBay)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 6)
	castWithCost(t, g, "One Drop", "Sorcery", "{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Fatalf("a one-drop got %d cascade offers, want 0", got)
	}
	castWithCost(t, g, "Two Drop", "Sorcery", "{1}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("the first spell with mana value 2 or greater got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	castWithCost(t, g, "Three Drop", "Sorcery", "{2}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("the second such spell got %d cascade offers, want 0", got)
	}
}

func TestTardisBayChaosStealsAnArtifactThenPlaneswalks(t *testing.T) {
	g, _ := newPlaneGame(t, planeTardisBay)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(bear)); err == nil {
		t.Fatal("a nonartifact was accepted as the target")
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, rock); c.Controller != me.ID {
		t.Error("the artifact is not under the planar controller's control")
	}
	endTurn(t, g)
	if c := tmtB1Card(t, g, rock); c.Controller != me.ID {
		t.Error("control of the artifact ended with the turn")
	}
	if _, ok := faceUpPlane(g, planeTardisBay.Oracle); ok {
		t.Error("the table did not planeswalk away")
	}
}

// --- The Moonbase ------------------------------------------------------------

func TestMoonbaseChaosReanimatesFaceDownCybermenFromEachOpponent(t *testing.T) {
	requirePlaneRegisteredFull(t, planeMoonbase)
	g, _ := newPlaneGame(t, planeMoonbase)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	o1 := g.Seats[(g.Turn.ActiveSeat+1)%4]
	o2 := g.Seats[(g.Turn.ActiveSeat+2)%4]
	d1 := pushGraveyardCardWithTypeLineID(o1, "Dead Wolf", "Creature — Wolf")
	d2 := pushGraveyardCardWithTypeLineID(o2, "Dead Bear", "Creature — Bear")
	_ = pushGraveyardCardWithTypeLineID(o2, "Dead Rock", "Artifact")
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	// One "up to one" prompt per opponent, in turn order: take what each offers.
	for range 3 {
		ch := latestChoiceOfKind(g, game.PendingChoicePickTarget)
		if ch == nil {
			break
		}
		var pick []game.TargetRef
		for _, want := range []uuid.UUID{d1, d2} {
			if slices.Contains(ch.PickTargetCards, want) {
				pick = append(pick, game.TargetRef{Kind: game.TargetCard, ID: want})
			}
		}
		if err := g.ResolvePickTargets(ch.ID, ch.Chooser, pick); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.FaceDownIsPermanent() {
			n++
			if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.IsArtifact() || !c.IsCreature() || !c.HasSubtypeWord("Cyberman") {
				t.Errorf("a face-down card is %d/%d %q, want a 2/2 Cyberman artifact creature", c.CurrentPower(), c.CurrentToughness(), c.Effective().Subtypes)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d face-down permanents under my control, want 2", n)
	}
}

func TestMoonbaseGivesEveryCreatureASorceryFlyingAbility(t *testing.T) {
	g, _ := newPlaneGame(t, planeMoonbase)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	b09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating the granted ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !planesB2Keyword(t, g, bear, "flying") {
		t.Error("the creature did not gain flying")
	}
}

// --- The Drum, Mining Facility -----------------------------------------------

func TestTheDrumPumpsOnAnyPlanarRollAndChaosEndsTheTurn(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTheDrum)
	g, plane := newPlaneGame(t, planeTheDrum)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rollBlank(t, g, me.ID)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, bear)
	if c.CurrentPower() != 3 || !planesB2Keyword(t, g, bear, "haste") {
		t.Errorf("after a blank the creature is %d/%d, want 3/3 with haste", c.CurrentPower(), c.CurrentToughness())
	}
	turn := g.Turn.Seq
	rollChaos(t, g, me.ID)
	answerOpenOrderingIfAny(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Errorf("%d items from the plane are still waiting", n)
	}
	if g.Turn.Seq == turn && g.Turn.Step != game.StepCleanup {
		t.Errorf("the turn did not end: still turn %d at %s", g.Turn.Seq, g.Turn.Step)
	}
}

// --- The Cheetah Planet ------------------------------------------------------

func TestCheetahPlanetMakesACatOfANonCatCreatureEachUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeCheetahPlanet)
	g, _ := newPlaneGame(t, planeCheetahPlanet)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	cat := b12Creature(g, me.ID, "Tabby", "Creature — Cat", 1, 1)
	planeswalkToPlane(t, g, planeCheetahPlanet.Oracle)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(cat)); err == nil {
		t.Fatal("a Cat was accepted as the target")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, bear)
	if c.CurrentPower() != 4 || !c.HasSubtypeWord("Cat") || !c.HasSubtypeWord("Bear") {
		t.Errorf("the creature is %d/%d %q, want a 4/4 Bear Cat", c.CurrentPower(), c.CurrentToughness(), c.Effective().Subtypes)
	}
}

func TestCheetahPlanetChaosGivesEveryCatProvokeUntilEndOfTurn(t *testing.T) {
	g, _ := newPlaneGame(t, planeCheetahPlanet)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := b12Creature(g, me.ID, "Tabby", "Creature — Cat", 1, 1)
	theirs := b12Creature(g, opp.ID, "Lion", "Creature — Cat", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	has := func(id uuid.UUID) bool {
		for _, ga := range layeredCard(t, g, id).Effective().GrantedAbilities {
			if ga.Key == game.GrantKey(theCheetahPlanetProvoke) {
				return true
			}
		}
		return false
	}
	if !has(mine) || !has(theirs) || has(bear) {
		t.Errorf("provoke on my Cat %v, their Cat %v, the Bear %v; want true, true, false", has(mine), has(theirs), has(bear))
	}
	endTurn(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if has(mine) {
		t.Error("provoke outlived the turn")
	}
}

// --- Prime Minister's Cabinet Room -------------------------------------------

func TestCabinetRoomCopiesAnOpponentsCreatureAtCombat(t *testing.T) {
	requirePlaneRegisteredFull(t, planeCabinetRoom)
	g, _ := newPlaneGame(t, planeCabinetRoom)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 5, 5)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(giant)); err == nil {
		t.Fatal("an opponent's creature was accepted for the first clause")
	}
	answerPickTarget(t, g, bear)
	answerPickTarget(t, g, giant)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, bear)
	if c.Name != "Giant" || c.CurrentPower() != 5 || !c.HasSubtypeWord("Giant") {
		t.Errorf("the creature is %q %d/%d, want a copy of the Giant", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	endTurn(t, g)
	if layeredCard(t, g, bear).Name != "Giant" {
		t.Error("the copy ended with the turn; it has no duration")
	}
}

func TestCabinetRoomChaosVotesAndExilesTheMostVotedCreatures(t *testing.T) {
	g, _ := newPlaneGame(t, planeCabinetRoom)
	advanceToMain(t, g)
	order := voteOrder(g)
	me := order[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := b12Creature(g, order[1].ID, "Voted Wolf", "Creature — Wolf", 2, 2)
	b := b12Creature(g, order[2].ID, "Other Wolf", "Creature — Wolf", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if ballotFor(g, me.ID) == nil {
		t.Fatal("no ballot for the planar controller")
	}
	if c := ballotFor(g, me.ID); len(c.PickOptions) != 2 {
		t.Errorf("the ballot offers %d creatures, want the 2 the planar controller doesn't control", len(c.PickOptions))
	}
	for _, p := range order {
		voteFor(t, g, p.ID, "Voted Wolf")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the creature with the most votes was not exiled")
	}
	if !g.Battlefield.Contains(b) || !g.Battlefield.Contains(mine) {
		t.Error("a creature with no votes was exiled")
	}
}

// --- The Doctor's Tomb -------------------------------------------------------

func TestDoctorsTombExilesDyingCreaturesAndCostsTheirControllerTwoLife(t *testing.T) {
	requirePlaneRegisteredFull(t, planeDoctorsTomb)
	g, _ := newPlaneGame(t, planeDoctorsTomb)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	wolf := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(wolf) })
	g.RunStateChecksForTest()
	if !g.Exile.Contains(wolf) || opp.Graveyard.Contains(wolf) {
		t.Error("the creature did not go to exile instead of the graveyard")
	}
	if opp.Life != life-2 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, life-2)
	}
	_ = me
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	g.RunStateChecksForTest()
	if !opp.Graveyard.Contains(bear) {
		t.Error("the replacement outlived the plane")
	}
}

func TestDoctorsTombChaosRedistributesLifeTotals(t *testing.T) {
	g, _ := newPlaneGame(t, planeDoctorsTomb)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := g.Seats[(g.Turn.ActiveSeat+1)%4]
	b := g.Seats[(g.Turn.ActiveSeat+2)%4]
	me.Life, a.Life, b.Life = 5, 15, 30
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answer := func(label string) {
		t.Helper()
		c := latestOptionPickFor(g, me.ID)
		if c == nil {
			t.Fatalf("no option prompt waiting for the planar controller (wanted %q)", label)
		}
		for i, o := range c.PickOptions {
			if o.Label == label {
				if err := g.ResolveOptionPick(c.ID, me.ID, i); err != nil {
					t.Fatalf("answering %q: %v", label, err)
				}
				return
			}
		}
		t.Fatalf("the prompt %q offers no %q: %+v", c.Reason, label, c.PickOptions)
	}
	// Include me (5) and the 30-life player, leave the 15-life player out.
	answer("Redistribute " + me.Name + "'s life total (5)")
	answer("Leave " + a.Name + "'s life total as it is")
	answer("Redistribute " + b.Name + "'s life total (30)")
	answer("Leave " + g.Seats[(g.Turn.ActiveSeat+3)%4].Name + "'s life total as it is")
	// Two players, two totals: I get 30, so the other gets 5.
	answer(b.Name + "'s life total (30)")
	if me.Life != 30 || b.Life != 5 || a.Life != 15 {
		t.Errorf("life totals %d, %d, %d, want 30, 5 and the third player's untouched 15", me.Life, b.Life, a.Life)
	}
}

// --- The Doctor's Childhood Barn ---------------------------------------------

func TestChildhoodBarnMakesCreaturesEnterTapped(t *testing.T) {
	requirePlaneRegisteredFull(t, planeChildhoodBarn)
	g, _ := newPlaneGame(t, planeChildhoodBarn)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	var tapped, found bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Test Bear" && c.Controller == me.ID {
			found, tapped = true, c.Tapped
		}
	}
	if !found || !tapped {
		t.Errorf("the creature is on the battlefield %v, tapped %v, want both", found, tapped)
	}
}

func TestChildhoodBarnChaosPhasesOutUntilThePlaneIsLeft(t *testing.T) {
	g, plane := newPlaneGame(t, planeChildhoodBarn)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(land)); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	answerPickTarget(t, g, rock)
	for i := 0; i < 3 && latestChoiceOfKind(g, game.PendingChoicePickTarget) != nil; i++ {
		ch := latestChoiceOfKind(g, game.PendingChoicePickTarget)
		if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
			t.Fatalf("declining an opponent's clause: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	var out bool
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(rock) })
	if !out {
		t.Fatal("the permanent did not phase out")
	}
	// Around to the owner's untap step: it stays out while the plane is face up.
	for range 4 {
		endTurn(t, g)
	}
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(rock) })
	if !out {
		t.Fatal("the permanent phased in at its controller's untap step with the plane face up")
	}
	if _, ok := faceUpPlane(g, planeChildhoodBarn.Oracle); !ok {
		t.Fatal("the plane is not face up any more")
	}
	_ = plane
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(g.Seats[g.Turn.ActiveSeat].ID) })
	g.RunStateChecksForTest()
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(rock) })
	if out {
		t.Error("the permanent did not phase in when the table planeswalked")
	}
}

// --- Two Streams Facility ----------------------------------------------------

// answerTwoStreams answers each player's choice in turn order from `first`.
func answerTwoStreams(t *testing.T, g *game.Game, first int, choices ...int) {
	t.Helper()
	for i, idx := range choices {
		p := g.Seats[(first+i)%4]
		c := latestOptionPickFor(g, p.ID)
		if c == nil {
			t.Fatalf("no choice is waiting for %s", p.Name)
		}
		if err := g.ResolveOptionPick(c.ID, p.ID, idx); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
	}
}

func TestTwoStreamsFacilityRemembersEachPlayersChoice(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTwoStreams)
	g, _ := newPlaneGame(t, planeTwoStreams)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	base := g.LandDropsRemainingFor(me.ID)
	planeswalkToPlane(t, g, planeTwoStreams.Oracle)
	passPriorityAroundTable(t, g)
	// Me: red waterfall. Next: green anchor. The rest: green anchor.
	answerTwoStreams(t, g, g.Turn.ActiveSeat, 1, 0, 0, 0)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := layeredCard(t, g, mine); c.CurrentPower() != 4 || !planesB2Keyword(t, g, mine, "haste") {
		t.Errorf("a creature of a red waterfall player is %d/%d, want power 4 with haste", c.CurrentPower(), c.CurrentToughness())
	}
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 2 || planesB2Keyword(t, g, theirs, "haste") {
		t.Errorf("a creature of a green anchor player is %d/%d, want 2/2 with no haste", c.CurrentPower(), c.CurrentToughness())
	}
	if got := g.LandDropsRemainingFor(opp.ID); got != base+1 {
		t.Errorf("a green anchor player has %d land plays, want %d", got, base+1)
	}
	if got := g.LandDropsRemainingFor(me.ID); got != base {
		t.Errorf("a red waterfall player has %d land plays, want %d", got, base)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := layeredCard(t, g, mine); c.CurrentPower() != 2 {
		t.Errorf("after the flip my creature has power %d, want 2", c.CurrentPower())
	}
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 4 || !planesB2Keyword(t, g, theirs, "haste") {
		t.Errorf("after the flip their creature is %d/%d, want power 4 with haste", c.CurrentPower(), c.CurrentToughness())
	}
	if got := g.LandDropsRemainingFor(me.ID); got != base+1 {
		t.Errorf("after the flip I have %d land plays, want %d", got, base+1)
	}
	if got := g.LandDropsRemainingFor(opp.ID); got != base {
		t.Errorf("after the flip the other player has %d land plays, want %d", got, base)
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 2 {
		t.Error("the plane's bonus outlived it")
	}
	if got := g.LandDropsRemainingFor(me.ID); got != base {
		t.Errorf("land plays %d after the plane left, want %d", got, base)
	}
}

func TestTwoStreamsFacilityAsksAtTheFirstUpkeepOfTheGameOnly(t *testing.T) {
	g, plane := newPlaneGame(t, planeTwoStreams)
	me := g.Seats[g.Turn.ActiveSeat]
	upkeep := game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID}
	g.Turn.Seq, g.Turn.StepOrdinal = 1, 1
	if !twoStreamsAsksForChoices(upkeep, &plane, game.Characteristic{}, g) {
		t.Error("the first upkeep of the game did not ask")
	}
	g.Turn.Seq = 2
	if twoStreamsAsksForChoices(upkeep, &plane, game.Characteristic{}, g) {
		t.Error("a later upkeep asked")
	}
	g.Turn.Seq, g.Turn.StepOrdinal = 1, 2
	if twoStreamsAsksForChoices(upkeep, &plane, game.Characteristic{}, g) {
		t.Error("an additional upkeep of the first turn asked")
	}
}
