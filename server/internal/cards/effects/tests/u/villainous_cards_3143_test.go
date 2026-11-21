package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

func villainRegistered(t *testing.T, oracle string) effects.Spec {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("%s completeness = %v, want full", spec.Name, spec.Completeness)
	}
	return spec
}

func TestTheMasterExilesTheDeadArtifactCreatureThenTheRichestOpponentFaces(t *testing.T) {
	g, me, robot := effectstest.MasterBoard(t)
	g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 30, 41, 25

	effectstest.VillainDestroy(g, robot)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if !villainInExile(g, robot) {
		t.Fatal("the Robot was not exiled")
	}
	c := effectstest.VillainPrompt(t, g, g.Seats[2].ID) // 41 life, the most
	if c.PickOptions[0].LifeCost != 4 {
		t.Errorf("option 0 = %+v, want the 4 life priced", c.PickOptions[0])
	}
	effectstest.VillainAnswer(t, g, c, 0)
	if g.Seats[2].Life != 37 {
		t.Errorf("the opponent with the most life has %d, want 41 - 4", g.Seats[2].Life)
	}
	if g.Seats[1].Life != 30 || g.Seats[3].Life != 25 {
		t.Error("another opponent lost life")
	}
	villainRegistered(t, effectstest.VillainMaster)
}

func TestTheMasterSecondOptionMakesATokenCopyOfTheExiledCard(t *testing.T) {
	g, me, robot := effectstest.MasterBoard(t)
	g.Seats[1].Life = 50

	effectstest.VillainDestroy(g, robot)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[1].ID), 1)

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
	g, me, robot := effectstest.MasterBoard(t)
	g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 40, 40, 39

	effectstest.VillainDestroy(g, robot)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestOptionPickFor(g, me.ID)
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
	effectstest.AnswerChoosePlayer(t, g, me.ID, g.Seats[2])
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[2].ID), 0)
	if g.Seats[2].Life != 36 || g.Seats[1].Life != 40 {
		t.Errorf("lives = %d / %d, want the chosen tied opponent at 36 and the other at 40", g.Seats[2].Life, g.Seats[1].Life)
	}
}

func TestTheMasterIgnoresTokensNonartifactsAndOtherPlayersAndMayBeDeclined(t *testing.T) {
	g, me, robot := effectstest.MasterBoard(t)
	servo := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Servo", TypeLine: "Token Artifact Creature — Servo", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	for _, id := range []uuid.UUID{servo, bear, theirs} {
		effectstest.VillainDestroy(g, id)
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceTriggerPrompt {
				t.Fatal("a death that is not 'a nontoken artifact creature you control' asked The Master's question")
			}
		}
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.VillainDestroy(g, robot)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if villainInExile(g, robot) || !me.Graveyard.Contains(robot) {
		t.Error("declined, the Robot should have stayed in the graveyard")
	}
	villainNoPrompt(t, g)
}

func TestTheMasterDoesNothingIfTheDeadCardLeftTheGraveyardInResponse(t *testing.T) {
	g, me, robot := effectstest.MasterBoard(t)

	effectstest.VillainDestroy(g, robot)
	g.RunStateChecksForTest()
	// In response, the card goes back to its owner's hand.
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(robot, game.ZoneHand, uuid.Nil); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if villainInExile(g, robot) {
		t.Error("a new object (CR 400.7) was exiled")
	}
	villainNoPrompt(t, g)
}

// ---- Genesis of the Daleks ---------------------------------------------------

func TestGenesisOfTheDaleksChaptersMakeADalekForEachLoreCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	sagaID := effectstest.CastCatalogSpell(t, g, "Genesis of the Daleks", "Enchantment — Saga", effectstest.VillainGenesis, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := villainTokensNamed(g, me.ID, "Dalek"); got != 1 {
		t.Fatalf("after chapter I: %d Daleks, want 1", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := villainTokensNamed(g, me.ID, "Dalek"); got != 3 {
		t.Fatalf("after chapter II: %d Daleks, want 1 + 2", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := villainTokensNamed(g, me.ID, "Dalek"); got != 6 {
		t.Fatalf("after chapter III: %d Daleks, want 3 + 3", got)
	}
	if got := effectstest.LoreCountersOn(g, sagaID); got != 3 {
		t.Errorf("lore counters = %d, want 3", got)
	}
	if spec := villainRegistered(t, effectstest.VillainGenesis); len(spec.Triggered) != 4 {
		t.Errorf("%d chapter abilities, want 4", len(spec.Triggered))
	}
}

// genesisChapterFour seeds the Saga on its third counter with a board
// to sweep, then runs the advance to chapter IV and targets opp.
func genesisChapterFour(t *testing.T) (g *game.Game, me, opp *game.Player, mineDalek, theirDalek, theirBear, myBear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp = g.Seats[seat], g.Seats[(seat+1)%4]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Genesis of the Daleks", TypeLine: "Enchantment — Saga", OracleID: effectstest.VillainGenesis,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 3},
	})
	mineDalek = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	theirDalek = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek Drone", TypeLine: "Artifact Creature — Dalek", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"+1/+1": 1},
	})
	theirBear = effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	myBear = effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	for _, p := range g.Seats {
		p.Life = 40
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	return g, me, opp, mineDalek, theirDalek, theirBear, myBear
}

func TestGenesisOfTheDaleksChapterFourFirstOptionDestroysDaleksAndDrainsByTheirPower(t *testing.T) {
	g, _, opp, mineDalek, theirDalek, theirBear, myBear := genesisChapterFour(t)
	// A Dalek that died earlier this turn counts too: 4 power as it died.
	earlier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek Scout", TypeLine: "Artifact Creature — Dalek", Power: 4, Toughness: 4,
		Owner: g.Seats[2].ID, Controller: g.Seats[2].ID,
	})
	effectstest.VillainDestroy(g, earlier)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 0)

	for _, id := range []uuid.UUID{mineDalek, theirDalek} {
		if effectstest.FindBattlefieldCardForTest(g, id) != nil {
			t.Errorf("Dalek %s survived the sweep", id)
		}
	}
	if effectstest.FindBattlefieldCardForTest(g, theirBear) == nil || effectstest.FindBattlefieldCardForTest(g, myBear) == nil {
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
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)

	for _, id := range []uuid.UUID{theirBear, myBear} {
		if effectstest.FindBattlefieldCardForTest(g, id) != nil {
			t.Errorf("non-Dalek %s survived the sweep", id)
		}
	}
	if effectstest.FindBattlefieldCardForTest(g, mineDalek) == nil || effectstest.FindBattlefieldCardForTest(g, theirDalek) == nil {
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
	shifter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shifter", TypeLine: "Creature — Shapeshifter", Power: 5, Toughness: 5,
		Keywords: []string{game.KeywordChangeling}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 0)

	if effectstest.FindBattlefieldCardForTest(g, shifter) != nil {
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
	effectstest.CastCatalogSpell(t, g, "Hunted by The Family", "Sorcery", effectstest.VillainHunted, effectstest.CardTargets(targets...))
}

func TestHuntedByTheFamilyOptionAMakesA1_1WhiteHumanWithNoAbilitiesForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Serra Angel", TypeLine: "Legendary Artifact Creature — Angel", OracleID: effectstest.SerraAngelID,
		Power: 4, Toughness: 4, Colors: []string{"W"},
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"+1/+1": 1},
	})
	_ = me
	if !effectstest.EffectiveAbilitiesContain(t, g, angel, "flying") {
		t.Fatal("setup: the Angel has no flying")
	}

	huntedCast(t, g, angel)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 0)

	c := effectstest.LayeredCard(t, g, angel)
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
	effectstest.AdvanceToPrecombatMainOf(t, g, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, 3)
	if got := effectstest.LayeredCard(t, g, angel).Effective(); len(got.Abilities) != 0 || got.Power != 1 {
		t.Errorf("the effect lapsed: %+v", got)
	}
	villainRegistered(t, effectstest.VillainHunted)
}

func TestHuntedByTheFamilyOptionBGivesTheCasterATokenCopyAndLeavesTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Serra Angel", TypeLine: "Creature — Angel", OracleID: effectstest.SerraAngelID,
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})

	huntedCast(t, g, angel)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)

	if n := villainTokensNamed(g, me.ID, "Serra Angel"); n != 1 {
		t.Fatalf("%d Serra Angel tokens under my control, want 1", n)
	}
	if got := effectstest.LayeredCard(t, g, angel); got.CurrentPower() != 4 || !effectstest.EffectiveAbilitiesContain(t, g, angel, "flying") {
		t.Errorf("the original was changed by option B: %+v", got.Effective())
	}
}

func TestHuntedByTheFamilyAsksEachControllerInTurnOrderAndSkipsATargetThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bears := make([]uuid.UUID, 4)
	for i := 1; i <= 3; i++ {
		bears[i] = effectstest.B16Creature(g, g.Seats[i].ID, "Bear", "Creature — Bear", 2, 2, "G")
	}
	// Chosen out of seat order: the questions still go seat 1, 2, 3.
	huntedCast(t, g, bears[3], bears[1], bears[2])
	effectstest.VillainDestroy(g, bears[2]) // in response: that target is gone
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[1].ID), 1)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[3].ID), 0)
	villainNoPrompt(t, g)

	if n := villainTokensNamed(g, me.ID, "Bear"); n != 1 {
		t.Errorf("%d Bear tokens, want 1 (seat 1 chose the copy)", n)
	}
	if got := effectstest.LayeredCard(t, g, bears[3]).Effective(); got.Power != 1 || !effectstest.HasString(got.Subtypes, "Human") {
		t.Errorf("seat 3's Bear = %+v, want a 1/1 Human", got)
	}
}

func TestHuntedByTheFamilyWithTheValeyardAsksEachControllerTwiceAndChainsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	one := effectstest.B16Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2, "G")
	two := effectstest.B16Creature(g, g.Seats[2].ID, "Wolf", "Creature — Wolf", 2, 2, "G")

	huntedCast(t, g, one, two)
	effectstest.PassPriorityAroundTable(t, g)

	// Seat 1 twice, then seat 2 twice, and nothing more: the chain into
	// seat 2's creature started when seat 1's LAST question was answered.
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[1].ID), 1)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[1].ID), 1)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[2].ID), 1)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[2].ID), 1)
	villainNoPrompt(t, g)

	if n := villainTokensNamed(g, me.ID, "Bear"); n != 2 {
		t.Errorf("%d Bear tokens, want 2", n)
	}
	if n := villainTokensNamed(g, me.ID, "Wolf"); n != 2 {
		t.Errorf("%d Wolf tokens, want 2", n)
	}
}

func TestHuntedByTheFamilyTargetsOnlyCreaturesYouDoNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Hunted by The Family", "Sorcery", effectstest.VillainHunted, effectstest.CardTargets(mine)); err == nil {
		t.Error("cast with a creature I control as a target")
	}
}

// ---- Midnight Crusader Shuttle ----------------------------------------------------

func shuttleAttacks(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	shuttle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Midnight Crusader Shuttle", TypeLine: "Artifact Creature — Vehicle",
		OracleID: effectstest.VillainShuttle, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.DeclareAttack(t, g, opp.ID, shuttle)
	effectstest.PassPriorityAroundTable(t, g)
	return shuttle
}

func TestMidnightCrusaderShuttleOptionASacrificesACreatureOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	wolf := effectstest.B16Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2, "G")
	shuttleAttacks(t, g, me, opp)

	c := effectstest.VillainPrompt(t, g, opp.ID)
	if c.PickOptions[0].Harm == 0 {
		t.Errorf("the sacrifice is free to a player with creatures: %+v", c.PickOptions[0])
	}
	effectstest.VillainAnswer(t, g, c, 0)
	effectstest.AnswerSacrifice(t, g, opp.ID, wolf)

	if effectstest.FindBattlefieldCardForTest(g, wolf) != nil || effectstest.FindBattlefieldCardForTest(g, bear) == nil {
		t.Error("the wrong creature was sacrificed")
	}
}

func TestMidnightCrusaderShuttleOptionBTakesTapsAndAttacksWithTheirCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B16Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4, "R")
	effectstest.B16Creature(g, g.Seats[2].ID, "Elsewhere", "Creature — Bear", 9, 9, "G")
	shuttle := shuttleAttacks(t, g, me, opp)

	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if pick == nil {
		t.Fatalf("the controller was not asked which creature to take: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ogre {
		t.Errorf("offered %v, want only the facing player's creature", pick.ChooseCards)
	}
	if err := g.ResolveTheirPermanents(pick.ID, me.ID, []uuid.UUID{ogre}); err != nil {
		t.Fatalf("ResolveTheirPermanents: %v", err)
	}

	c := effectstest.FindBattlefieldCardForTest(g, ogre)
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
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shuttleAttacks(t, g, me, opp)

	c := effectstest.VillainPrompt(t, g, opp.ID)
	if c.PickOptions[0].Harm != 0 {
		t.Errorf("sacrificing a creature you do not have is priced as a cost: %+v", c.PickOptions[0])
	}
	effectstest.VillainAnswer(t, g, c, 1)
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID); pick != nil {
		t.Errorf("the controller was asked to choose among no creatures: %+v", pick)
	}
	villainNoPrompt(t, g)
}

func TestMidnightCrusaderShuttleHasCrew2(t *testing.T) {
	spec := villainRegistered(t, effectstest.VillainShuttle)
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
	effectstest.CastCatalogSpell(t, g, "Ensnared by the Mara", "Sorcery", effectstest.VillainMara, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestEnsnaredByTheMaraOptionAExilesUntilANonlandAndLetsTheCasterCastItFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, hit, deep := maraLibrary(opp)

	castMara(t, g)
	c := effectstest.VillainPrompt(t, g, opp.ID)
	effectstest.VillainAnswer(t, g, c, 0)

	if !villainInExile(g, land) || !villainInExile(g, hit) {
		t.Fatal("the land and the nonland card were not both exiled")
	}
	if villainInExile(g, deep) {
		t.Error("the exile ran past the first nonland card")
	}
	perm := g.CastPermissionOnCardByIDForEffect(hit)
	if !effectstest.PermissionLive(g, perm, me.ID) {
		t.Fatalf("the caster has no live free-cast permission on the nonland card: %+v", perm)
	}
	if effectstest.PermissionLive(g, perm, opp.ID) {
		t.Error("the facing player may cast their own exiled card")
	}
	if p := g.CastPermissionOnCardByIDForEffect(land); effectstest.PermissionLive(g, p, me.ID) {
		t.Error("a land was offered as a free cast")
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the free cast from exile was refused: %v", err)
	}
	if villainInExile(g, hit) {
		t.Error("the card is still in exile after being cast")
	}
	villainRegistered(t, effectstest.VillainMara)
}

func TestEnsnaredByTheMaraOptionAWithNoNonlandCardExilesTheLibraryAndGrantsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	opp.Library.Cards = nil
	for i := 0; i < 3; i++ {
		c := game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID}
		opp.Library.Cards = append(opp.Library.Cards, c)
		ids = append(ids, c.InstanceID)
	}

	castMara(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 0)

	for _, id := range ids {
		if !villainInExile(g, id) {
			t.Error("a land stayed in a library that has no nonland card")
		}
		if p := g.CastPermissionOnCardByIDForEffect(id); effectstest.PermissionLive(g, p, me.ID) {
			t.Error("a permission was granted on a land")
		}
	}
}

func TestEnsnaredByTheMaraOptionBExilesFourAndDealsTheirTotalManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land, hit, deep := maraLibrary(opp)
	extra := game.Card{InstanceID: uuid.New(), Name: "Bottom", TypeLine: "Instant", ManaCost: "{2}{U}", Owner: opp.ID, Controller: opp.ID}
	opp.Library.Cards = append([]game.Card{extra}, opp.Library.Cards...) // under the deep card
	before := opp.Life

	castMara(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)

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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	opp.Library.Cards = []game.Card{{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID}}
	before := opp.Life

	castMara(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)

	if got := before - opp.Life; got != 1 {
		t.Errorf("took %d, want 1", got)
	}
}

func TestEnsnaredByTheMaraAsksEachOpponentInTurnOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for _, p := range g.Seats[1:] {
		maraLibrary(p)
	}
	castMara(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[1].ID), 1)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[2].ID), 1)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[3].ID), 1)
	villainNoPrompt(t, g)
}

// ---- Missy -------------------------------------------------------------------------

func missyBoard(t *testing.T) (g *game.Game, me *game.Player, missy uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	missy = effectstest.B12Push(g, me.ID, "Missy", "Legendary Creature — Time Lord Rogue", effectstest.VillainMissy, 4, 5)
	effectstest.AdvanceToMain(t, g)
	return g, me, missy
}

func TestMissyReturnsAnotherNonartifactCreatureFaceDownTappedUnderHerControl(t *testing.T) {
	g, me, _ := missyBoard(t)
	opp := g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")

	effectstest.VillainDestroy(g, bear)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.IsCyberman(t, g, bear)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("controller %s owner %s, want Missy's controller and the card's owner", c.Controller, c.Owner)
	}
	if !c.Tapped {
		t.Error("it came back untapped")
	}
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Error("CR 708.5: only its controller may look at it")
	}
	villainRegistered(t, effectstest.VillainMissy)
}

func TestMissyIgnoresArtifactsTokensAndHerself(t *testing.T) {
	g, me, missy := missyBoard(t)
	robot := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.VillainDestroy(g, robot)
	effectstest.VillainDestroy(g, token)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, robot) != nil || effectstest.FindBattlefieldCardForTest(g, token) != nil {
		t.Error("an artifact creature or a token came back")
	}
	if !me.Graveyard.Contains(robot) {
		t.Error("the artifact creature left the graveyard")
	}

	effectstest.VillainDestroy(g, missy)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, missy) != nil {
		t.Error("Missy returned herself; the ability says another")
	}
}

func TestMissyKeepsABodyFaceDownWhateverTheCardIs(t *testing.T) {
	g, me, _ := missyBoard(t)
	opp := g.Seats[1]
	// A noncreature card cannot die as a creature, but an animated land can.
	manland := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mutavault", TypeLine: "Land Creature — Shapeshifter", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.VillainDestroy(g, manland)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.IsCyberman(t, g, manland)
	if c, _ := effectstest.BattlefieldCard(g, manland); c.IsLand() || c.Controller != me.ID {
		t.Errorf("the face-down body kept its land type or its owner's control: %+v", c)
	}
}

func missyEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMissyEndStepOptionAEachArtifactCreatureDealsOneDamage(t *testing.T) {
	g, me, _ := missyBoard(t)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Robot", TypeLine: "Artifact Creature — Robot", Power: 2, Toughness: 2,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	for _, p := range g.Seats {
		p.Life = 40
	}

	missyEndStep(t, g)
	// Each opponent in turn order: seat 1 takes the damage, seats 2 and 3 let me draw.
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[1].ID), 0)
	handBefore := me.Hand.Size()
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[2].ID), 1)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[3].ID), 1)

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
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.VillainDestroy(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.IsCyberman(t, g, bear)
	for _, p := range g.Seats {
		p.Life = 40
	}

	missyEndStep(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 0)
	if opp.Life != 39 {
		t.Errorf("the Cyberman dealt %d damage, want 1", 40-opp.Life)
	}
	_ = me
}
