package effects

import (
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
