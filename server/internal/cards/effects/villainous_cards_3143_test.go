package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// villainous_cards_3143_test.go — #3143: the six villainous-choice cards
// that each needed one piece more than the keyword action. The shape
// itself is villainous_choice_test.go.

const (
	villainMaster  = "e5e2e4ef-1d78-41eb-b634-1a90d4cb7125"
	villainGenesis = "71ff4603-9f7a-459e-b5b3-9b39583f1dc7"
	villainHunted  = "7d178f1c-faf7-4efd-87a6-201a40c55ec8"
	villainShuttle = "22fa2860-d746-4429-b7e3-3403bef27f7b"
	villainMara    = "16090270-0ac5-404c-b0c8-a6c8a4b508a4"
	villainMissy   = "905610cd-1461-4f5a-b9c8-19b193b9d513"
	serraAngelID   = "4b7ac066-e5c7-43e6-9e7e-2739b24a905d"
)

func villainDestroy(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

func villainInExile(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func villainTokensNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			n++
		}
	}
	return n
}

func villainRegistered(t *testing.T, oracle string) Spec {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("%s completeness = %v, want full", spec.Name, spec.Completeness)
	}
	return spec
}

// ---- The Master, Gallifrey's End -------------------------------------------

func masterBoard(t *testing.T) (g *game.Game, me *game.Player, robot uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	b12Push(g, me.ID, "The Master, Gallifrey's End", "Legendary Creature — Time Lord Rogue", villainMaster, 4, 3)
	robot = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	return g, me, robot
}

func TestTheMasterExilesTheDeadArtifactCreatureThenTheRichestOpponentFaces(t *testing.T) {
	g, me, robot := masterBoard(t)
	g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 30, 41, 25

	villainDestroy(g, robot)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if !villainInExile(g, robot) {
		t.Fatal("the Robot was not exiled")
	}
	c := villainPrompt(t, g, g.Seats[2].ID) // 41 life, the most
	if c.PickOptions[0].LifeCost != 4 {
		t.Errorf("option 0 = %+v, want the 4 life priced", c.PickOptions[0])
	}
	villainAnswer(t, g, c, 0)
	if g.Seats[2].Life != 37 {
		t.Errorf("the opponent with the most life has %d, want 41 - 4", g.Seats[2].Life)
	}
	if g.Seats[1].Life != 30 || g.Seats[3].Life != 25 {
		t.Error("another opponent lost life")
	}
	villainRegistered(t, villainMaster)
}

func TestTheMasterSecondOptionMakesATokenCopyOfTheExiledCard(t *testing.T) {
	g, me, robot := masterBoard(t)
	g.Seats[1].Life = 50

	villainDestroy(g, robot)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[1].ID), 1)

	if n := villainTokensNamed(g, me.ID, "Robot"); n != 1 {
		t.Fatalf("%d Robot tokens under my control, want 1", n)
	}
	if !villainInExile(g, robot) {
		t.Error("the exiled card came back; only a copy should")
	}
	if g.Seats[1].Life != 50 {
		t.Errorf("the opponent lost life choosing the token: %d", g.Seats[1].Life)
	}
}

func TestTheMasterAsksTheControllerToBreakATieAmongTheTiedOnly(t *testing.T) {
	g, me, robot := masterBoard(t)
	g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 40, 40, 39

	villainDestroy(g, robot)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the controller was not asked to break the tie: %+v", g.PendingChoices)
	}
	if len(pick.PickOptions) != 2 {
		t.Errorf("%d seats offered, want only the two tied: %+v", len(pick.PickOptions), pick.PickOptions)
	}
	for _, o := range pick.PickOptions {
		if o.Label == g.Seats[3].Name {
			t.Errorf("the opponent below the top was offered: %+v", pick.PickOptions)
		}
	}
	answerChoosePlayer(t, g, me.ID, g.Seats[2])
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[2].ID), 0)
	if g.Seats[2].Life != 36 || g.Seats[1].Life != 40 {
		t.Errorf("lives = %d / %d, want the chosen tied opponent at 36 and the other at 40", g.Seats[2].Life, g.Seats[1].Life)
	}
}

func TestTheMasterIgnoresTokensNonartifactsAndOtherPlayersAndMayBeDeclined(t *testing.T) {
	g, me, robot := masterBoard(t)
	servo := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Servo", TypeLine: "Token Artifact Creature — Servo", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	for _, id := range []uuid.UUID{servo, bear, theirs} {
		villainDestroy(g, id)
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceTriggerPrompt {
				t.Fatal("a death that is not 'a nontoken artifact creature you control' asked The Master's question")
			}
		}
	}
	passPriorityAroundTable(t, g)

	villainDestroy(g, robot)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if villainInExile(g, robot) || !me.Graveyard.Contains(robot) {
		t.Error("declined, the Robot should have stayed in the graveyard")
	}
	villainNoPrompt(t, g)
}

func TestTheMasterDoesNothingIfTheDeadCardLeftTheGraveyardInResponse(t *testing.T) {
	g, me, robot := masterBoard(t)

	villainDestroy(g, robot)
	g.RunStateChecksForTest()
	// In response, the card goes back to its owner's hand.
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(robot, game.ZoneHand, uuid.Nil); err != nil {
			t.Fatal(err)
		}
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if villainInExile(g, robot) {
		t.Error("a new object (CR 400.7) was exiled")
	}
	villainNoPrompt(t, g)
}

// ---- Genesis of the Daleks ---------------------------------------------------

func TestGenesisOfTheDaleksChaptersMakeADalekForEachLoreCounter(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	sagaID := castCatalogSpell(t, g, "Genesis of the Daleks", "Enchantment — Saga", villainGenesis, nil)
	passPriorityAroundTable(t, g)

	if got := villainTokensNamed(g, me.ID, "Dalek"); got != 1 {
		t.Fatalf("after chapter I: %d Daleks, want 1", got)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := villainTokensNamed(g, me.ID, "Dalek"); got != 3 {
		t.Fatalf("after chapter II: %d Daleks, want 1 + 2", got)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := villainTokensNamed(g, me.ID, "Dalek"); got != 6 {
		t.Fatalf("after chapter III: %d Daleks, want 3 + 3", got)
	}
	if got := loreCountersOn(g, sagaID); got != 3 {
		t.Errorf("lore counters = %d, want 3", got)
	}
	if spec := villainRegistered(t, villainGenesis); len(spec.Triggered) != 4 {
		t.Errorf("%d chapter abilities, want 4", len(spec.Triggered))
	}
}

// genesisChapterFour seeds the Saga on its third counter with a board
// to sweep, then runs the advance to chapter IV and targets opp.
func genesisChapterFour(t *testing.T) (g *game.Game, me, opp *game.Player, mineDalek, theirDalek, theirBear, myBear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp = g.Seats[seat], g.Seats[(seat+1)%4]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Genesis of the Daleks", TypeLine: "Enchantment — Saga", OracleID: villainGenesis,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 3},
	})
	mineDalek = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	theirDalek = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek Drone", TypeLine: "Artifact Creature — Dalek", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"+1/+1": 1},
	})
	theirBear = b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	myBear = b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	for _, p := range g.Seats {
		p.Life = 40
	}
	advanceToPrecombatMainOf(t, g, seat)
	answerPickTargetPlayer(t, g, opp.ID)
	return g, me, opp, mineDalek, theirDalek, theirBear, myBear
}

func TestGenesisOfTheDaleksChapterFourFirstOptionDestroysDaleksAndDrainsByTheirPower(t *testing.T) {
	g, _, opp, mineDalek, theirDalek, theirBear, myBear := genesisChapterFour(t)
	// A Dalek that died earlier this turn counts too: 4 power as it died.
	earlier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek Scout", TypeLine: "Artifact Creature — Dalek", Power: 4, Toughness: 4,
		Owner: g.Seats[2].ID, Controller: g.Seats[2].ID,
	})
	villainDestroy(g, earlier)
	passPriorityAroundTable(t, g)

	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 0)

	for _, id := range []uuid.UUID{mineDalek, theirDalek} {
		if findBattlefieldCardForTest(g, id) != nil {
			t.Errorf("Dalek %s survived the sweep", id)
		}
	}
	if findBattlefieldCardForTest(g, theirBear) == nil || findBattlefieldCardForTest(g, myBear) == nil {
		t.Error("a non-Dalek was destroyed by option A")
	}
	// 4 (earlier) + 3 (token) + 3 (2/2 with a +1/+1 counter) = 10, from each
	// opponent of the controller; the controller is untouched.
	for _, p := range g.Seats {
		want := 40 - 10
		if p.ID == g.Seats[g.Turn.ActiveSeat].ID {
			want = 40
		}
		if p.Life != want {
			t.Errorf("%s has %d life, want %d", p.Name, p.Life, want)
		}
	}
}

func TestGenesisOfTheDaleksChapterFourSecondOptionDestroysEverythingElse(t *testing.T) {
	g, _, opp, mineDalek, theirDalek, theirBear, myBear := genesisChapterFour(t)
	passPriorityAroundTable(t, g)

	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)

	for _, id := range []uuid.UUID{theirBear, myBear} {
		if findBattlefieldCardForTest(g, id) != nil {
			t.Errorf("non-Dalek %s survived the sweep", id)
		}
	}
	if findBattlefieldCardForTest(g, mineDalek) == nil || findBattlefieldCardForTest(g, theirDalek) == nil {
		t.Error("a Dalek was destroyed by option B")
	}
	for _, p := range g.Seats {
		if p.Life != 40 {
			t.Errorf("%s lost life to option B: %d", p.Name, p.Life)
		}
	}
}

func TestGenesisOfTheDaleksCountsAChangelingAsADalek(t *testing.T) {
	g, _, opp, _, _, _, _ := genesisChapterFour(t)
	shifter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shifter", TypeLine: "Creature — Shapeshifter", Power: 5, Toughness: 5,
		Keywords: []string{game.KeywordChangeling}, Owner: opp.ID, Controller: opp.ID,
	})
	passPriorityAroundTable(t, g)

	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 0)

	if findBattlefieldCardForTest(g, shifter) != nil {
		t.Error("a changeling is a Dalek and should have been destroyed")
	}
	// 3 + 3 + 5 = 11
	if opp.Life != 29 {
		t.Errorf("the opponent has %d life, want 40 - 11", opp.Life)
	}
}

// ---- Hunted by The Family ------------------------------------------------------

func huntedCast(t *testing.T, g *game.Game, targets ...uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Hunted by The Family", "Sorcery", villainHunted, cardTargets(targets...))
}

func TestHuntedByTheFamilyOptionAMakesA1_1WhiteHumanWithNoAbilitiesForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Serra Angel", TypeLine: "Legendary Artifact Creature — Angel", OracleID: serraAngelID,
		Power: 4, Toughness: 4, Colors: []string{"W"},
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"+1/+1": 1},
	})
	_ = me
	if !effectiveAbilitiesContain(t, g, angel, "flying") {
		t.Fatal("setup: the Angel has no flying")
	}

	huntedCast(t, g, angel)
	passPriorityAroundTable(t, g)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 0)

	c := layeredCard(t, g, angel)
	eff := c.Effective()
	if got := c.CurrentPower(); got != 2 {
		t.Errorf("power = %d, want 1 base + the +1/+1 counter", got)
	}
	if got := c.CurrentToughness(); got != 2 {
		t.Errorf("toughness = %d, want 1 base + the +1/+1 counter", got)
	}
	if len(eff.Colors) != 1 || eff.Colors[0] != "W" {
		t.Errorf("colours = %v, want white", eff.Colors)
	}
	if !c.IsCreature() || c.IsArtifact() || !c.HasSubtype("Human") || c.HasSubtype("Angel") {
		t.Errorf("types: creature %v artifact %v Human %v Angel %v — want a Human creature that is no longer an artifact Angel",
			c.IsCreature(), c.IsArtifact(), c.HasSubtype("Human"), c.HasSubtype("Angel"))
	}
	if len(eff.Abilities) != 0 {
		t.Errorf("abilities = %v, want none", eff.Abilities)
	}
	if c.Controller != opp.ID {
		t.Error("option A changed the creature's controller")
	}
	// No duration: it is still so two turns later.
	advanceToPrecombatMainOf(t, g, 2)
	advanceToPrecombatMainOf(t, g, 3)
	if got := layeredCard(t, g, angel).Effective(); len(got.Abilities) != 0 || got.Power != 1 {
		t.Errorf("the effect lapsed: %+v", got)
	}
	villainRegistered(t, villainHunted)
}

func TestHuntedByTheFamilyOptionBGivesTheCasterATokenCopyAndLeavesTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Serra Angel", TypeLine: "Creature — Angel", OracleID: serraAngelID,
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})

	huntedCast(t, g, angel)
	passPriorityAroundTable(t, g)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)

	if n := villainTokensNamed(g, me.ID, "Serra Angel"); n != 1 {
		t.Fatalf("%d Serra Angel tokens under my control, want 1", n)
	}
	if got := layeredCard(t, g, angel); got.CurrentPower() != 4 || !effectiveAbilitiesContain(t, g, angel, "flying") {
		t.Errorf("the original was changed by option B: %+v", got.Effective())
	}
}

func TestHuntedByTheFamilyAsksEachControllerInTurnOrderAndSkipsATargetThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := make([]uuid.UUID, 4)
	for i := 1; i <= 3; i++ {
		bears[i] = b16Creature(g, g.Seats[i].ID, "Bear", "Creature — Bear", 2, 2, "G")
	}
	// Chosen out of seat order: the questions still go seat 1, 2, 3.
	huntedCast(t, g, bears[3], bears[1], bears[2])
	villainDestroy(g, bears[2]) // in response: that target is gone
	passPriorityAroundTable(t, g)

	villainAnswer(t, g, villainPrompt(t, g, g.Seats[1].ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[3].ID), 0)
	villainNoPrompt(t, g)

	if n := villainTokensNamed(g, me.ID, "Bear"); n != 1 {
		t.Errorf("%d Bear tokens, want 1 (seat 1 chose the copy)", n)
	}
	if got := layeredCard(t, g, bears[3]).Effective(); got.Power != 1 || !hasString(got.Subtypes, "Human") {
		t.Errorf("seat 3's Bear = %+v, want a 1/1 Human", got)
	}
}

func TestHuntedByTheFamilyWithTheValeyardAsksEachControllerTwiceAndChainsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	one := b16Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2, "G")
	two := b16Creature(g, g.Seats[2].ID, "Wolf", "Creature — Wolf", 2, 2, "G")

	huntedCast(t, g, one, two)
	passPriorityAroundTable(t, g)

	// Seat 1 twice, then seat 2 twice, and nothing more: the chain into
	// seat 2's creature started when seat 1's LAST question was answered.
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[1].ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[1].ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[2].ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[2].ID), 1)
	villainNoPrompt(t, g)

	if n := villainTokensNamed(g, me.ID, "Bear"); n != 2 {
		t.Errorf("%d Bear tokens, want 2", n)
	}
	if n := villainTokensNamed(g, me.ID, "Wolf"); n != 2 {
		t.Errorf("%d Wolf tokens, want 2", n)
	}
}

func TestHuntedByTheFamilyTargetsOnlyCreaturesYouDoNotControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	if err := castCatalogSpellErr(t, g, "Hunted by The Family", "Sorcery", villainHunted, cardTargets(mine)); err == nil {
		t.Error("cast with a creature I control as a target")
	}
}

// ---- Midnight Crusader Shuttle ----------------------------------------------------

func shuttleAttacks(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	shuttle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Midnight Crusader Shuttle", TypeLine: "Artifact Creature — Vehicle",
		OracleID: villainShuttle, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	declareAttack(t, g, opp.ID, shuttle)
	passPriorityAroundTable(t, g)
	return shuttle
}

func TestMidnightCrusaderShuttleOptionASacrificesACreatureOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	wolf := b16Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2, "G")
	shuttleAttacks(t, g, me, opp)

	c := villainPrompt(t, g, opp.ID)
	if c.PickOptions[0].Harm == 0 {
		t.Errorf("the sacrifice is free to a player with creatures: %+v", c.PickOptions[0])
	}
	villainAnswer(t, g, c, 0)
	answerSacrifice(t, g, opp.ID, wolf)

	if findBattlefieldCardForTest(g, wolf) != nil || findBattlefieldCardForTest(g, bear) == nil {
		t.Error("the wrong creature was sacrificed")
	}
}

func TestMidnightCrusaderShuttleOptionBTakesTapsAndAttacksWithTheirCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b16Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4, "R")
	b16Creature(g, g.Seats[2].ID, "Elsewhere", "Creature — Bear", 9, 9, "G")
	shuttle := shuttleAttacks(t, g, me, opp)

	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if pick == nil {
		t.Fatalf("the controller was not asked which creature to take: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ogre {
		t.Errorf("offered %v, want only the facing player's creature", pick.ChooseCards)
	}
	if err := g.ResolveTheirPermanents(pick.ID, me.ID, []uuid.UUID{ogre}); err != nil {
		t.Fatalf("ResolveTheirPermanents: %v", err)
	}

	c := findBattlefieldCardForTest(g, ogre)
	if c == nil || c.Controller != me.ID {
		t.Fatalf("I do not control the Ogre: %+v", c)
	}
	if !c.Tapped {
		t.Error("the stolen creature is untapped")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("AttackingTarget = %s, want the facing player", c.AttackingTarget)
	}
	if n := attackEventsForCard(g, ogre); n != 0 {
		t.Errorf("%d EventAttack for the stolen creature; it was put into combat, not declared (CR 506.3c)", n)
	}

	before := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3+4 {
		t.Errorf("the defender lost %d, want the Shuttle's 3 and the Ogre's 4", got)
	}
	_ = shuttle
}

func attackEventsForCard(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventAttack && ev.CardID == id {
			n++
		}
	}
	return n
}

func TestMidnightCrusaderShuttleWithNoCreatureEitherOptionDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shuttleAttacks(t, g, me, opp)

	c := villainPrompt(t, g, opp.ID)
	if c.PickOptions[0].Harm != 0 {
		t.Errorf("sacrificing a creature you do not have is priced as a cost: %+v", c.PickOptions[0])
	}
	villainAnswer(t, g, c, 1)
	if pick := latestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID); pick != nil {
		t.Errorf("the controller was asked to choose among no creatures: %+v", pick)
	}
	villainNoPrompt(t, g)
}

func TestMidnightCrusaderShuttleHasCrew2(t *testing.T) {
	spec := villainRegistered(t, villainShuttle)
	crew := false
	for _, a := range spec.Activated {
		crew = crew || a.Cost.Crew == 2
	}
	if !crew {
		t.Error("Crew 2 is missing")
	}
}

// ---- Ensnared by the Mara -----------------------------------------------------------

func maraLibrary(p *game.Player) (top2, nonland, deep uuid.UUID) {
	mk := func(name, typeLine, cost string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID}
	}
	l1, l2 := mk("Island", "Basic Land — Island", ""), mk("Swamp", "Basic Land — Swamp", "")
	spell := mk("Big Spell", "Creature — Dragon", "{4}{R}{R}")
	spell.Power, spell.Toughness = 5, 5
	d := mk("Deep Card", "Sorcery", "{1}")
	p.Library.Cards = []game.Card{d, spell, l2, l1} // the top is the last element
	return l1.InstanceID, spell.InstanceID, d.InstanceID
}

func castMara(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Ensnared by the Mara", "Sorcery", villainMara, nil)
	passPriorityAroundTable(t, g)
}

func TestEnsnaredByTheMaraOptionAExilesUntilANonlandAndLetsTheCasterCastItFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, hit, deep := maraLibrary(opp)

	castMara(t, g)
	c := villainPrompt(t, g, opp.ID)
	villainAnswer(t, g, c, 0)

	if !villainInExile(g, land) || !villainInExile(g, hit) {
		t.Fatal("the land and the nonland card were not both exiled")
	}
	if villainInExile(g, deep) {
		t.Error("the exile ran past the first nonland card")
	}
	perm := g.CastPermissionOnCardByIDForEffect(hit)
	if !permissionLive(g, perm, me.ID) {
		t.Fatalf("the caster has no live free-cast permission on the nonland card: %+v", perm)
	}
	if permissionLive(g, perm, opp.ID) {
		t.Error("the facing player may cast their own exiled card")
	}
	if p := g.CastPermissionOnCardByIDForEffect(land); permissionLive(g, p, me.ID) {
		t.Error("a land was offered as a free cast")
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the free cast from exile was refused: %v", err)
	}
	if villainInExile(g, hit) {
		t.Error("the card is still in exile after being cast")
	}
	villainRegistered(t, villainMara)
}

func TestEnsnaredByTheMaraOptionAWithNoNonlandCardExilesTheLibraryAndGrantsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	opp.Library.Cards = nil
	for i := 0; i < 3; i++ {
		c := game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID}
		opp.Library.Cards = append(opp.Library.Cards, c)
		ids = append(ids, c.InstanceID)
	}

	castMara(t, g)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 0)

	for _, id := range ids {
		if !villainInExile(g, id) {
			t.Error("a land stayed in a library that has no nonland card")
		}
		if p := g.CastPermissionOnCardByIDForEffect(id); permissionLive(g, p, me.ID) {
			t.Error("a permission was granted on a land")
		}
	}
}

func TestEnsnaredByTheMaraOptionBExilesFourAndDealsTheirTotalManaValue(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land, hit, deep := maraLibrary(opp)
	extra := game.Card{InstanceID: uuid.New(), Name: "Bottom", TypeLine: "Instant", ManaCost: "{2}{U}", Owner: opp.ID, Controller: opp.ID}
	opp.Library.Cards = append([]game.Card{extra}, opp.Library.Cards...) // under the deep card
	before := opp.Life

	castMara(t, g)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)

	for _, id := range []uuid.UUID{land, hit} {
		if !villainInExile(g, id) {
			t.Error("a top card was not exiled")
		}
	}
	if !villainInExile(g, deep) {
		t.Error("the fourth card was not exiled")
	}
	if villainInExile(g, extra.InstanceID) {
		t.Error("a fifth card was exiled")
	}
	// 0 + 0 + 6 + 1 = 7
	if got := before - opp.Life; got != 7 {
		t.Errorf("the opponent took %d, want the total mana value 7", got)
	}
}

func TestEnsnaredByTheMaraOptionBWithAShortLibraryDealsWhatItExiled(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Library.Cards = []game.Card{{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID}}
	before := opp.Life

	castMara(t, g)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)

	if got := before - opp.Life; got != 1 {
		t.Errorf("took %d, want 1", got)
	}
}

func TestEnsnaredByTheMaraAsksEachOpponentInTurnOrder(t *testing.T) {
	g := newCatalogGame(t)
	for _, p := range g.Seats[1:] {
		maraLibrary(p)
	}
	castMara(t, g)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[1].ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[2].ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[3].ID), 1)
	villainNoPrompt(t, g)
}

// ---- Missy -------------------------------------------------------------------------

func missyBoard(t *testing.T) (g *game.Game, me *game.Player, missy uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	missy = b12Push(g, me.ID, "Missy", "Legendary Creature — Time Lord Rogue", villainMissy, 4, 5)
	advanceToMain(t, g)
	return g, me, missy
}

func TestMissyReturnsAnotherNonartifactCreatureFaceDownTappedUnderHerControl(t *testing.T) {
	g, me, _ := missyBoard(t)
	opp := g.Seats[1]
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")

	villainDestroy(g, bear)
	passPriorityAroundTable(t, g)

	isCyberman(t, g, bear)
	c, _ := battlefieldCard(g, bear)
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("controller %s owner %s, want Missy's controller and the card's owner", c.Controller, c.Owner)
	}
	if !c.Tapped {
		t.Error("it came back untapped")
	}
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Error("CR 708.5: only its controller may look at it")
	}
	villainRegistered(t, villainMissy)
}

func TestMissyIgnoresArtifactsTokensAndHerself(t *testing.T) {
	g, me, missy := missyBoard(t)
	robot := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	villainDestroy(g, robot)
	villainDestroy(g, token)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, robot) != nil || findBattlefieldCardForTest(g, token) != nil {
		t.Error("an artifact creature or a token came back")
	}
	if !me.Graveyard.Contains(robot) {
		t.Error("the artifact creature left the graveyard")
	}

	villainDestroy(g, missy)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, missy) != nil {
		t.Error("Missy returned herself; the ability says another")
	}
}

func TestMissyKeepsABodyFaceDownWhateverTheCardIs(t *testing.T) {
	g, me, _ := missyBoard(t)
	opp := g.Seats[1]
	// A noncreature card cannot die as a creature, but an animated land can.
	manland := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mutavault", TypeLine: "Land Creature — Shapeshifter", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID,
	})
	villainDestroy(g, manland)
	passPriorityAroundTable(t, g)
	isCyberman(t, g, manland)
	if c, _ := battlefieldCard(g, manland); c.IsLand() || c.Controller != me.ID {
		t.Errorf("the face-down body kept its land type or its owner's control: %+v", c)
	}
}

func missyEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
}

func TestMissyEndStepOptionAEachArtifactCreatureDealsOneDamage(t *testing.T) {
	g, me, _ := missyBoard(t)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	for _, p := range g.Seats {
		p.Life = 40
	}

	missyEndStep(t, g)
	// Each opponent in turn order: seat 1 takes the damage, seats 2 and 3 let me draw.
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[1].ID), 0)
	handBefore := me.Hand.Size()
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[2].ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[3].ID), 1)

	if g.Seats[1].Life != 38 {
		t.Errorf("seat 1 has %d life, want 40 - 2 (my two artifact creatures; the Bear and their Robot do not count)", g.Seats[1].Life)
	}
	if g.Seats[2].Life != 40 || g.Seats[3].Life != 40 {
		t.Error("an opponent who chose the draw took damage")
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d -> %d, want two draws", handBefore, got)
	}
}

func TestMissyArtifactCreatureOptionCountsHerFaceDownCybermen(t *testing.T) {
	g, me, _ := missyBoard(t)
	opp := g.Seats[1]
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	villainDestroy(g, bear)
	passPriorityAroundTable(t, g)
	isCyberman(t, g, bear)
	for _, p := range g.Seats {
		p.Life = 40
	}

	missyEndStep(t, g)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 0)
	if opp.Life != 39 {
		t.Errorf("the Cyberman dealt %d damage, want 1", 40-opp.Life)
	}
	_ = me
}
