package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b190_test.go — S82 GAP batch 190 (#5948), Warhammer 40,000 Commander
// creatures. Each test plays the real catalog entry.

const (
	gapB190StalkerOracle      = "1b3ae9c3-1940-4fb8-a47c-2d9e1790f4a3"
	gapB190VenomOracle        = "f3a07ae4-b0f8-43e6-adba-217f220cc2d0"
	gapB190DefilerOracle      = "eba7ef08-f444-4d23-92b8-29d2ac08b156"
	gapB190HelbruteOracle     = "4b216b4d-f9cb-4ce2-af70-c91eef3db083"
	gapB190LuciusOracle       = "fa726aac-f4ae-4e34-ac08-36c1f50215f3"
	gapB190AbaddonOracle      = "8dedb9a7-0afa-4dff-b124-2344c170f11a"
	gapB190BelakorOracle      = "3690ca72-c925-42ab-acc2-7cc1fda05789"
	gapB190CallidusOracle     = "4240dab9-ace7-490d-97c3-489a9fd52038"
	gapB190DatasmithOracle    = "23a265ef-d651-4d8d-94dd-0072f83541f9"
	gapB190EisenhornOracle    = "db07a3f9-c6d1-4cae-a899-12b3e8d94018"
	gapB190GreyfaxOracle      = "9439d6eb-4c10-451d-85c4-3752611c671c"
	gapB190MarneusOracle      = "f6479f7e-01f4-49f1-a444-04bf38934f6b"
	gapB190ArtifactDemonTL    = "Artifact Creature — Demon"
	gapB190ArtifactAstartesTL = "Artifact Creature — Astartes Dreadnought"
	gapB190CreatureBearTL     = "Creature — Bear"
	gapB190LegendaryAstarTL   = "Legendary Creature — Astartes Warrior"
)

func TestGapB190CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB190StalkerOracle:   "Triarch Stalker",
		gapB190VenomOracle:     "Venomcrawler",
		gapB190DefilerOracle:   "Chaos Defiler",
		gapB190HelbruteOracle:  "Helbrute",
		gapB190LuciusOracle:    "Lucius the Eternal",
		gapB190AbaddonOracle:   "Abaddon the Despoiler",
		gapB190BelakorOracle:   "Be'lakor, the Dark Master",
		gapB190CallidusOracle:  "Callidus Assassin",
		gapB190DatasmithOracle: "Cybernetica Datasmith",
		gapB190EisenhornOracle: "Inquisitor Eisenhorn",
		gapB190GreyfaxOracle:   "Inquisitor Greyfax",
		gapB190MarneusOracle:   "Marneus Calgar",
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

// --- Venomcrawler ----------------------------------------------------------

func TestGapB190VenomcrawlerGrowsWhenAnyOtherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crawler := effectstest.B12Push(g, me.ID, "Venomcrawler", gapB190ArtifactDemonTL, gapB190VenomOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", gapB190CreatureBearTL, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, crawler); p != 3 {
		t.Fatalf("power %d after an opposing creature died, want 3", p)
	}
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, crawler); p != 4 {
		t.Errorf("power %d after my own creature died, want 4", p)
	}
}

func TestGapB190VenomcrawlerHasLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crawler := effectstest.B12Push(g, me.ID, "Venomcrawler", gapB190ArtifactDemonTL, gapB190VenomOracle, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, crawler, "lifelink") {
		t.Error("Venomcrawler has no lifelink")
	}
}

// --- Inquisitor Greyfax ----------------------------------------------------

func TestGapB190GreyfaxGivesOtherCreaturesYouControlPlusOneAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	greyfax := effectstest.B12Push(g, me.ID, "Inquisitor Greyfax", "Legendary Creature — Human Inquisitor", gapB190GreyfaxOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", gapB190CreatureBearTL, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 || !effectstest.HasKeywordOnBattlefield(t, g, mine, "vigilance") {
		t.Errorf("my creature: power %d, want 3 with vigilance", p)
	}
	if p := effectstest.EffectivePower(t, g, greyfax); p != 3 {
		t.Errorf("Greyfax pumped itself: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 || effectstest.HasKeywordOnBattlefield(t, g, theirs, "vigilance") {
		t.Errorf("an opposing creature was affected: power %d", p)
	}
}

func TestGapB190GreyfaxTapsAnOpposingCreatureAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	greyfax := effectstest.B12Push(g, me.ID, "Inquisitor Greyfax", "Legendary Creature — Human Inquisitor", gapB190GreyfaxOracle, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := effectstest.ActivateOn(t, g, me.ID, greyfax, 0, theirs); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.FindBattlefieldByID(g, theirs)
	if !c.Tapped {
		t.Error("the opposing creature was not tapped")
	}
	clues := 0
	for _, bc := range g.Battlefield.Cards {
		if bc.Controller == me.ID && bc.Name == "Clue" {
			clues++
		}
	}
	if clues != 1 {
		t.Errorf("%d Clues, want 1", clues)
	}
}

func TestGapB190GreyfaxCannotTapYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	greyfax := effectstest.B12Push(g, me.ID, "Inquisitor Greyfax", "Legendary Creature — Human Inquisitor", gapB190GreyfaxOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := effectstest.ActivateOn(t, g, me.ID, greyfax, 0, mine); err == nil {
		t.Error("the ability accepted a creature its controller controls")
	}
}

// --- Marneus Calgar --------------------------------------------------------

func TestGapB190MarneusMakesTwoWarriorsAndDrawsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marneus := effectstest.B12Push(g, me.ID, "Marneus Calgar", "Legendary Creature — Astartes Warrior", gapB190MarneusOracle, 3, 5)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, marneus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	warriors := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Astartes Warrior" {
			warriors++
			if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
				t.Errorf("Astartes Warrior is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
			}
		}
	}
	if warriors != 2 {
		t.Fatalf("%d Astartes Warriors, want 2", warriors)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want exactly one card for two tokens entering together", hand, me.Hand.Size())
	}
}

func TestGapB190MarneusIgnoresOpposingTokensAndNontokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Marneus Calgar", "Legendary Creature — Astartes Warrior", gapB190MarneusOracle, 3, 5)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.AvB16CastCreature(t, g, "Bear", gapB190CreatureBearTL, "", 2, 2)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 white Astartes Warrior with vigilance"), 1); err != nil {
			t.Error(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw for a nontoken or an opposing token", hand, me.Hand.Size())
	}
}

// --- Cybernetica Datasmith -------------------------------------------------

func gapB190DatasmithTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smith := effectstest.B12Push(g, me.ID, "Cybernetica Datasmith", "Artifact Creature — Human Artificer", gapB190DatasmithOracle, 0, 1)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.SeedLibrary(opp, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	return g, me, opp, smith
}

func gapB190PlayerTargets(first, second uuid.UUID) []game.TargetRef {
	return []game.TargetRef{
		{Kind: game.TargetPlayer, ID: first, Slot: 0},
		{Kind: game.TargetPlayer, ID: second, Slot: 1},
	}
}

func TestGapB190DatasmithDrawsForOnePlayerAndGivesAnotherARobot(t *testing.T) {
	g, me, opp, smith := gapB190DatasmithTable(t)
	effectstest.FloatMana(t, g, me, "{U}")
	meHand, oppHand := me.Hand.Size(), opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, smith, 0, game.ActivateAbilityParams{Strict: true, Targets: gapB190PlayerTargets(me.ID, opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != meHand+1 || opp.Hand.Size() != oppHand {
		t.Errorf("hands %d -> %d and %d -> %d, want only the first target to draw", meHand, me.Hand.Size(), oppHand, opp.Hand.Size())
	}
	robots := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Robot" && c.IsToken() {
			robots++
			if c.Controller != opp.ID || c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !c.HasCardType("artifact") {
				t.Errorf("Robot: controller %v, %d/%d, types %q", c.Controller, c.CurrentPower(), c.CurrentToughness(), c.TypeLine)
			}
		}
	}
	if robots != 1 {
		t.Fatalf("%d Robot tokens, want 1 under the second target's control", robots)
	}
}

func TestGapB190DatasmithRobotCannotBlock(t *testing.T) {
	g, me, opp, smith := gapB190DatasmithTable(t)
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, smith, 0, game.ActivateAbilityParams{Strict: true, Targets: gapB190PlayerTargets(me.ID, opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var robot uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Robot" && c.IsToken() {
			robot = c.InstanceID
		}
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", gapB190CreatureBearTL, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.AttackIntoBlocks(t, g, opp.ID, bear)
	if effectstest.BlockRefusal(t, g, robot, bear) == "" {
		t.Error("the Robot token was allowed to block")
	}
	if r := effectstest.BlockRefusal(t, g, other, bear); r != "" {
		t.Errorf("an ordinary creature's block was refused: %v", r)
	}
}

func TestGapB190DatasmithNeedsTwoDifferentPlayers(t *testing.T) {
	g, me, _, smith := gapB190DatasmithTable(t)
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, smith, 0, game.ActivateAbilityParams{Strict: true, Targets: gapB190PlayerTargets(me.ID, me.ID)}); err == nil {
		t.Error("the same player was accepted for both clauses")
	}
}

func TestGapB190DatasmithHasProtectionFromRobots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smith := effectstest.B12Push(g, me.ID, "Cybernetica Datasmith", "Artifact Creature — Human Artificer", gapB190DatasmithOracle, 0, 1)
	robot := effectstest.B12Creature(g, opp.ID, "Robot Soldier", "Artifact Creature — Robot", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AttackIntoBlocks(t, g, opp.ID, smith)
	if effectstest.BlockRefusal(t, g, robot, smith) == "" {
		t.Error("a Robot blocked a creature with protection from Robots")
	}
	if r := effectstest.BlockRefusal(t, g, bear, smith); r != "" {
		t.Errorf("a non-Robot's block was refused: %v", r)
	}
}

// --- Be'lakor, the Dark Master --------------------------------------------

func TestGapB190BelakorDrawsAndLosesLifeForEachDemonYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Imp", "Creature — Demon", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand, life := me.Hand.Size(), me.Life
	effectstest.AvB16CastCreature(t, g, "Be'lakor, the Dark Master", "Legendary Creature — Demon Noble", gapB190BelakorOracle, 6, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("hand %d -> %d, life %d -> %d, want 2 cards and 2 life for two Demons", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestGapB190BelakorCountsOnlyDemonsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, opp.ID, "Their Imp", "Creature — Demon", 1, 1)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand, life := me.Hand.Size(), me.Life
	effectstest.AvB16CastCreature(t, g, "Be'lakor, the Dark Master", "Legendary Creature — Demon Noble", gapB190BelakorOracle, 6, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d, want 1 and 1 (only Be'lakor)", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestGapB190BelakorAnotherDemonDealsItsPowerToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Be'lakor, the Dark Master", "Legendary Creature — Demon Noble", gapB190BelakorOracle, 6, 5)
	life := opp.Life
	effectstest.AvB16CastCreature(t, g, "Big Imp", "Creature — Demon", "", 3, 3)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 damage (the Demon's power)", life, opp.Life)
	}
}

func TestGapB190BelakorIgnoresNonDemonsAndOpposingDemons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Be'lakor, the Dark Master", "Legendary Creature — Demon Noble", gapB190BelakorOracle, 6, 5)
	life := opp.Life
	effectstest.AvB16CastCreature(t, g, "Big Bear", gapB190CreatureBearTL, "", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Imp", "Creature — Demon", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); c != nil {
		t.Fatalf("a target prompt opened for a non-Demon: %+v", c)
	}
	if opp.Life != life {
		t.Errorf("opponent life %d -> %d, want no damage", life, opp.Life)
	}
}

// --- Inquisitor Eisenhorn --------------------------------------------------

// gapB190OpponentsTurn walks the table to the opponent's next precombat main
// phase, so that nothing has been drawn by the first player yet this turn (a
// turn's draw step already counts as the first card drawn).
func gapB190OpponentsTurn(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepPrecombatMain {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached the opponent's main phase")
}

func TestGapB190EisenhornRevealingAnInstantFirstDrawMakesCherubael(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Inquisitor Eisenhorn", "Legendary Creature — Human Inquisitor", gapB190EisenhornOracle, 2, 3)
	gapB190OpponentsTurn(t, g)
	effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) }) // the first card I draw this turn
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cherubael" && c.Controller == me.ID {
			found = true
			if c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "flying") {
				t.Errorf("Cherubael is %d/%d", c.CurrentPower(), c.CurrentToughness())
			}
		}
	}
	if !found {
		t.Error("revealing an instant did not make Cherubael")
	}
}

func TestGapB190EisenhornDecliningTheRevealMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Inquisitor Eisenhorn", "Legendary Creature — Human Inquisitor", gapB190EisenhornOracle, 2, 3)
	gapB190OpponentsTurn(t, g)
	effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cherubael" {
			t.Error("Cherubael was made without a reveal")
		}
	}
}

func TestGapB190EisenhornOnlyTheFirstCardDrawnEachTurnCounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Inquisitor Eisenhorn", "Legendary Creature — Human Inquisitor", gapB190EisenhornOracle, 2, 3)
	effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	effectstest.SeedLibraryTop(me, "Rock", "Artifact", "{1}")
	effectstest.AdvanceToMain(t, g)                            // the draw step draws the artifact: the first card
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) }) // the instant is the second
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestConfirmFor(g, me.ID); c != nil {
		t.Errorf("a reveal prompt opened for a card that was not the first drawn: %+v", c)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cherubael" {
			t.Error("Cherubael was made for the second card drawn")
		}
	}
}

func TestGapB190EisenhornCombatDamageInvestigatesThatManyTimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eisenhorn := effectstest.B12Push(g, me.ID, "Inquisitor Eisenhorn", "Legendary Creature — Human Inquisitor", gapB190EisenhornOracle, 3, 3)
	effectstest.AttackWith(t, g, opp.ID, eisenhorn)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	clues := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Clue" {
			clues++
		}
	}
	if clues != 3 {
		t.Errorf("%d Clues, want 3 for 3 combat damage", clues)
	}
}

// --- Callidus Assassin -----------------------------------------------------

func gapB190CastCallidusAsCopyOf(t *testing.T, g *game.Game, copied uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Callidus Assassin", "Creature — Human Shapeshifter Assassin", gapB190CallidusOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, copied)
	return id
}

func TestGapB190CallidusEntersTappedAsACopyAndDestroysASameNamedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", gapB190CreatureBearTL, 2, 2)
	theirs := effectstest.SeedCopyableCreature(g, opp.ID, "Grizzly Bears", gapB190CreatureBearTL, 2, 2)
	different := effectstest.SeedCopyableCreature(g, opp.ID, "Runeclaw Bear", gapB190CreatureBearTL, 2, 2)
	id := gapB190CastCallidusAsCopyOf(t, g, mine)
	// The pick: a differently named creature is not a legal target.
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: different}); err == nil {
		t.Fatal("a creature with a different name was accepted")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Grizzly Bears" || !got.Tapped {
		t.Errorf("Callidus is %q tapped=%v, want a tapped Grizzly Bears", got.Name, got.Tapped)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the same-named opposing creature was not destroyed")
	}
	if !g.Battlefield.Contains(different) || !g.Battlefield.Contains(id) {
		t.Error("the differently named creature or the copy itself was destroyed")
	}
}

func TestGapB190CallidusTriggerMayTargetNothingAndNeverItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", gapB190CreatureBearTL, 2, 2)
	id := gapB190CastCallidusAsCopyOf(t, g, mine)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
		t.Fatal("the copy could target itself (\"other\")")
	}
}

func TestGapB190CallidusDeclinedCopyIsUntappedWithNoTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", gapB190CreatureBearTL, 2, 2)
	id := gapB190CastCallidusAsCopyOf(t, g, uuid.Nil)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Callidus Assassin" || got.Tapped {
		t.Errorf("declined copy: %q tapped=%v, want an untapped Callidus Assassin", got.Name, got.Tapped)
	}
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); c != nil {
		t.Error("a declined copy has the granted trigger")
	}
}

func TestGapB190CallidusHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Callidus Assassin", "Creature — Human Shapeshifter Assassin", gapB190CallidusOracle, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flash") {
		t.Error("Callidus Assassin has no flash")
	}
}

// --- Triarch Stalker -------------------------------------------------------

func gapB190StalkerChooses(t *testing.T, g *game.Game) (stalker uuid.UUID, chosen, other *game.Player) {
	t.Helper()
	me := g.Seats[0]
	stalker = effectstest.B12Push(g, me.ID, "Triarch Stalker", gapB190ArtifactDemonTL, gapB190StalkerOracle, 4, 5)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	id := g.ChosenPlayerOf(stalker)
	for _, s := range g.Seats[1:] {
		if s.ID == id {
			chosen = s
		} else if other == nil {
			other = s
		}
	}
	if chosen == nil {
		t.Fatal("Triarch Stalker has no chosen opponent after its trigger")
	}
	return stalker, chosen, other
}

func TestGapB190StalkerGivesCreaturesAttackingTheChosenPlayerMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, chosen, other := gapB190StalkerChooses(t, g)
	toChosen := effectstest.B12Creature(g, me.ID, "Bear A", gapB190CreatureBearTL, 2, 2)
	toOther := effectstest.B12Creature(g, me.ID, "Bear B", gapB190CreatureBearTL, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(toChosen, chosen.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(toOther, other.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, toChosen, "menace") {
		t.Error("a creature attacking the chosen player lacks menace")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, toOther, "menace") {
		t.Error("a creature attacking a different player has menace")
	}
}

func TestGapB190StalkerGivesNothingOnceItHasLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stalker, chosen, _ := gapB190StalkerChooses(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.KillCreature(t, g, me.ID, stalker)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, chosen.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "menace") {
		t.Error("menace outlived Triarch Stalker")
	}
}

// --- Chaos Defiler ---------------------------------------------------------

func gapB190DefilerBoards(g *game.Game) (bears map[uuid.UUID]uuid.UUID, picks map[uuid.UUID][]uuid.UUID) {
	bears = map[uuid.UUID]uuid.UUID{}
	picks = map[uuid.UUID][]uuid.UUID{}
	for _, s := range g.Seats[1:] {
		effectstest.B12Permanent(g, s.ID, "Forest", "Basic Land — Forest")
		b := effectstest.B12Creature(g, s.ID, "Bear", gapB190CreatureBearTL, 2, 2)
		bears[s.ID] = b
		picks[s.ID] = []uuid.UUID{b}
	}
	return bears, picks
}

func gapB190DestroyedCount(g *game.Game, bears map[uuid.UUID]uuid.UUID) int {
	n := 0
	for _, b := range bears {
		if !g.Battlefield.Contains(b) {
			n++
		}
	}
	return n
}

func TestGapB190DefilerEntersChoosesOnePerOpponentAndDestroysExactlyOneAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bears, picks := gapB190DefilerBoards(g)
	effectstest.AvB16CastCreature(t, g, "Chaos Defiler", gapB190ArtifactDemonTL, gapB190DefilerOracle, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if legs := effectstest.AnswerPermanentPicks(t, g, me.ID, picks); legs != 3 {
		t.Fatalf("%d picks asked, want one per opponent (3)", legs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB190DestroyedCount(g, bears); n != 1 {
		t.Errorf("%d chosen permanents were destroyed, want exactly 1", n)
	}
	lands := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" {
			lands++
		}
	}
	if lands != 3 {
		t.Errorf("%d lands left, want all 3 (lands are never chosen)", lands)
	}
}

func TestGapB190DefilerWontLetYouChooseALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Creature(g, opp.ID, "Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.AvB16CastCreature(t, g, "Chaos Defiler", gapB190ArtifactDemonTL, gapB190DefilerOracle, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	var leg *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTheirPermanents && c.FromPlayer == opp.ID {
			leg = c
		}
	}
	if leg == nil {
		t.Fatal("no pick over the opponent's board")
	}
	if err := g.ResolveTheirPermanents(leg.ID, me.ID, []uuid.UUID{forest}); err == nil {
		t.Error("a land was accepted as the choice")
	}
}

func TestGapB190DefilerDiesDoesItAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	defiler := effectstest.B12Push(g, me.ID, "Chaos Defiler", gapB190ArtifactDemonTL, gapB190DefilerOracle, 5, 4)
	bears, picks := gapB190DefilerBoards(g)
	effectstest.KillCreature(t, g, me.ID, defiler)
	effectstest.PassPriorityAroundTable(t, g)
	if legs := effectstest.AnswerPermanentPicks(t, g, me.ID, picks); legs != 3 {
		t.Fatalf("%d picks asked on death, want 3", legs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB190DestroyedCount(g, bears); n != 1 {
		t.Errorf("%d chosen permanents were destroyed on death, want exactly 1", n)
	}
}

func TestGapB190DefilerSkipsAnOpponentWithNoNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.AvB16CastCreature(t, g, "Chaos Defiler", gapB190ArtifactDemonTL, gapB190DefilerOracle, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTheirPermanents && c.FromPlayer == opp.ID {
			t.Fatal("an opponent with only a land was asked about")
		}
	}
}

// --- Helbrute ----------------------------------------------------------------

func TestGapB190HelbruteCastsFromTheGraveyardByExilingAnotherCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	brute := effectstest.GraveyardCardOf(g, me, "Helbrute", gapB190ArtifactAstartesTL, "{3}{B}{R}", gapB190HelbruteOracle)
	spent := effectstest.GraveyardCardOf(g, me, "Dead Bear", gapB190CreatureBearTL, "{1}{G}", "")
	notCreature := effectstest.GraveyardCardOf(g, me, "Spent Bolt", "Instant", "{R}", "")
	effectstest.PayMana(t, g, me, "{C}{C}{C}{B}{R}")
	cast := func(ids []uuid.UUID) error {
		return g.CastSpell(me.ID, brute, game.CastSpellParams{
			FromZone: "graveyard", AlternativeCost: "exile_from_graveyard_in_addition", AltCostIDs: ids, Strict: true,
		})
	}
	if err := cast([]uuid.UUID{notCreature}); err == nil {
		t.Fatal("exiled a noncreature card to pay")
	}
	if err := cast([]uuid.UUID{brute}); err == nil {
		t.Fatal("exiled Helbrute itself to pay (it must be another card)")
	}
	if err := cast(nil); err == nil {
		t.Fatal("cast with nothing exiled")
	}
	if err := cast([]uuid.UUID{spent}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	if !g.Exile.Contains(spent) {
		t.Error("the exiled creature card was not exiled")
	}
	if !g.Stack.Contains(brute) {
		t.Error("Helbrute did not reach the stack")
	}
}

func TestGapB190HelbruteHasHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Helbrute", gapB190ArtifactAstartesTL, gapB190HelbruteOracle, 5, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
		t.Error("Helbrute has no haste")
	}
}

// --- Lucius the Eternal -------------------------------------------------------

// gapB190BattlefieldByName finds a card by name: a card that returns from exile
// is a new object, so its old instance ID names nothing (CR 400.7).
func gapB190BattlefieldByName(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

func gapB190LuciusDies(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	lucius := effectstest.B12Push(g, me.ID, "Lucius the Eternal", gapB190LegendaryAstarTL, gapB190LuciusOracle, 5, 3)
	effectstest.KillCreature(t, g, me.ID, lucius)
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)
	return lucius
}

func TestGapB190LuciusExilesItselfThenReturnsWhenTheChosenCreatureLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	lucius := gapB190LuciusDies(t, g, victim)
	if effectstest.B12ZoneOf(g, lucius) != game.ZoneExile {
		t.Fatalf("Lucius is in %v, want exile", effectstest.B12ZoneOf(g, lucius))
	}
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := gapB190BattlefieldByName(g, "Lucius the Eternal")
	if !ok {
		t.Fatalf("Lucius is in %v after the chosen creature left, want the battlefield", effectstest.B12ZoneOf(g, lucius))
	}
	if c.Controller != me.ID {
		t.Errorf("Lucius returned under %v, want its owner", c.Controller)
	}
}

func TestGapB190LuciusStaysInExileWhileTheChosenCreatureStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other Bear", gapB190CreatureBearTL, 2, 2)
	lucius := gapB190LuciusDies(t, g, victim)
	effectstest.KillCreature(t, g, opp.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, lucius) != game.ZoneExile {
		t.Errorf("Lucius is in %v, want exile while the chosen creature is still there", effectstest.B12ZoneOf(g, lucius))
	}
}

func TestGapB190LuciusReturnsWhenTheChosenCreatureIsBouncedToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	lucius := gapB190LuciusDies(t, g, victim)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, victim); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := gapB190BattlefieldByName(g, "Lucius the Eternal"); !ok {
		t.Errorf("Lucius is in %v, want the battlefield (leaves the battlefield, not only dies)", effectstest.B12ZoneOf(g, lucius))
	}
}

func TestGapB190LuciusCanOnlyChooseAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", gapB190CreatureBearTL, 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", gapB190CreatureBearTL, 2, 2)
	lucius := effectstest.B12Push(g, me.ID, "Lucius the Eternal", gapB190LegendaryAstarTL, gapB190LuciusOracle, 5, 3)
	effectstest.KillCreature(t, g, me.ID, lucius)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("a creature its controller controls was accepted")
	}
}

// --- Abaddon the Despoiler -------------------------------------------------------

func gapB190AbaddonTable(t *testing.T, lost int) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedCheapLibrary(g, me, 6)
	effectstest.B12Push(g, me.ID, "Abaddon the Despoiler", "Legendary Creature — Astartes Warrior", gapB190AbaddonOracle, 5, 5)
	if lost > 0 {
		effectstest.LoseLife(t, g, opp, lost)
	}
	return g, me
}

func TestGapB190AbaddonGivesCascadeToASpellWithinTheLifeLostByOpponents(t *testing.T) {
	g, me := gapB190AbaddonTable(t, 4)
	effectstest.CastWithCost(t, g, "Four Drop", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("a mana value 4 spell after 4 life lost got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
}

func TestGapB190AbaddonGivesNothingToASpellCostingMoreThanTheLifeLost(t *testing.T) {
	g, _ := gapB190AbaddonTable(t, 3)
	effectstest.CastWithCost(t, g, "Four Drop", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("a mana value 4 spell after 3 life lost got %d cascade offers, want 0", got)
	}
}

func TestGapB190AbaddonCountsLifeLostByEveryOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCheapLibrary(g, me, 6)
	effectstest.B12Push(g, me.ID, "Abaddon the Despoiler", "Legendary Creature — Astartes Warrior", gapB190AbaddonOracle, 5, 5)
	effectstest.LoseLife(t, g, g.Seats[1], 2)
	effectstest.LoseLife(t, g, g.Seats[2], 2)
	effectstest.CastWithCost(t, g, "Four Drop", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Errorf("%d cascade offers after 2+2 life lost by two opponents, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
}

func TestGapB190AbaddonIgnoresLifeYouLostAndAFreeSpellWithNoLifeLost(t *testing.T) {
	g, me := gapB190AbaddonTable(t, 0)
	effectstest.LoseLife(t, g, me, 5)
	effectstest.CastWithCost(t, g, "Four Drop", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("my own life loss gave a mana value 4 spell %d cascade offers, want 0", got)
	}
}

func TestGapB190AbaddonOnlyHelpsSpellsCastFromYourHand(t *testing.T) {
	g, me := gapB190AbaddonTable(t, 6)
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: "Exiled Spell", TypeLine: "Instant", ManaCost: "{3}{G}", Owner: me.ID, Controller: me.ID})
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player: me.ID, CastOnly: true, Duration: game.WhileInZoneDuration(), Cost: "{0}",
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("a spell cast from exile got %d cascade offers, want 0", got)
	}
}

func TestGapB190AbaddonOnlyHelpsOnYourOwnTurn(t *testing.T) {
	g, me := gapB190AbaddonTable(t, 0)
	gapB190OpponentsTurn(t, g)
	effectstest.LoseLife(t, g, g.Seats[2], 6) // lost this turn, which is the opponent's
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Flash Spell", TypeLine: "Instant", ManaCost: "{3}{G}", Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast during the opponent's turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("a spell cast on an opponent's turn got %d cascade offers, want 0", got)
	}
}
