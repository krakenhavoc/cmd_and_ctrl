package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Grasp of Fate ---------------------------------------------------------

func TestCmdrGraspOfFateExilesOnePermanentPerOpponentUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	pushVanillaCreature(g, b.ID, "B Spare", 2, 2)
	theirs := pushVanillaCreature(g, b.ID, "B's Bear", 2, 2)
	others := pushVanillaCreature(g, c.ID, "C's Bear", 2, 2)
	land := pushTypedCard(g, g.Seats[3].ID, "D's Forest", "Basic Land — Forest", "")
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	ench := castCatalogSpell(t, g, "Grasp of Fate", "Enchantment", cmdrGraspOfFateOracle, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 16; i++ {
		hobAnswerTargets(t, g, me.ID, theirs, others)
		hobAnswerOrder(t, g, me.ID)
		if len(g.StackMeta) == 0 {
			break
		}
		passPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(others) {
		t.Fatal("the chosen permanents were not exiled")
	}
	if !g.Exile.Contains(theirs) || !g.Exile.Contains(others) {
		t.Fatal("the picks are not in exile")
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(mine) {
		t.Error("a land or my own creature was exiled")
	}
	destroyNow(t, g, ench)
	g.RunStateChecksForTest()
	if battlefieldNamed(g, b.ID, "B's Bear") != 1 || battlefieldNamed(g, c.ID, "C's Bear") != 1 {
		t.Error("the exiled permanents did not return when Grasp of Fate left")
	}
}

func TestCmdrGraspOfFateDoesNotTargetLandsOrYourOwnPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	land := pushTypedCard(g, b.ID, "B's Forest", "Basic Land — Forest", "")
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	castCatalogSpell(t, g, "Grasp of Fate", "Enchantment", cmdrGraspOfFateOracle, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 8; i++ {
		hobAnswerTargets(t, g, me.ID, land, mine)
		if len(g.StackMeta) == 0 {
			break
		}
		passPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(mine) {
		t.Error("Grasp of Fate exiled a land or my own permanent")
	}
}

// --- Emeria Shepherd -----------------------------------------------------------

func cmdrPlayLand(t *testing.T, g *game.Game, name, typeLine string) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
}

// cmdrAnswerOptionAfterTarget finishes a landfall trigger: the target
// pick if one is asked, then the option prompt.
func cmdrAnswerOption(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, index int) {
	t.Helper()
	for i := 0; i < 16; i++ {
		hobAnswerTargets(t, g, me.ID, target)
		if latestOptionPickFor(g, me.ID) != nil {
			answerOptionPick(t, g, me.ID, index)
			passPriorityAroundTable(t, g)
			return
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("no option prompt was asked")
}

func cmdrEmeriaBoard(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdrPush(g, me.ID, "Emeria Shepherd", "Creature — Angel", cmdrEmeriaShepherdOracle, 4, 4)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return g, me, dead
}

func TestCmdrEmeriaShepherdReturnsToHandForAnyLand(t *testing.T) {
	g, me, dead := cmdrEmeriaBoard(t)
	cmdrPlayLand(t, g, "Forest", "Basic Land — Forest")
	cmdrAnswerOption(t, g, me, dead, 1)
	if !me.Hand.Contains(dead) {
		t.Fatal("the card did not return to hand")
	}
	if g.Battlefield.Contains(dead) {
		t.Error("it entered the battlefield off a Forest")
	}
}

func TestCmdrEmeriaShepherdCanPutItOntoTheBattlefieldOffAPlains(t *testing.T) {
	g, me, dead := cmdrEmeriaBoard(t)
	cmdrPlayLand(t, g, "Plains", "Basic Land — Plains")
	cmdrAnswerOption(t, g, me, dead, 2)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the card did not return to the battlefield off a Plains")
	}
}

func TestCmdrEmeriaShepherdMayDeclineAndTakesNoLand(t *testing.T) {
	g, me, dead := cmdrEmeriaBoard(t)
	landCard := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: landCard, Name: "Old Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	cmdrPlayLand(t, g, "Plains", "Basic Land — Plains")
	cmdrAnswerOption(t, g, me, dead, 0)
	if me.Hand.Contains(dead) || g.Battlefield.Contains(dead) || !me.Graveyard.Contains(dead) {
		t.Error("declining moved the card")
	}
	if me.Hand.Contains(landCard) || g.Battlefield.Contains(landCard) {
		t.Error("a land card was returned")
	}
}

func TestCmdrEmeriaShepherdOffersNoBattlefieldReturnOffAForest(t *testing.T) {
	g, me, dead := cmdrEmeriaBoard(t)
	cmdrPlayLand(t, g, "Forest", "Basic Land — Forest")
	for i := 0; i < 16 && latestOptionPickFor(g, me.ID) == nil; i++ {
		hobAnswerTargets(t, g, me.ID, dead)
		passPriorityAroundTable(t, g)
	}
	c := latestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no option prompt")
	}
	if len(c.PickOptions) != 2 {
		t.Errorf("%d options off a Forest, want 2 (leave, hand)", len(c.PickOptions))
	}
}

// --- Norn's Choirmaster ----------------------------------------------------------

func cmdrCastCommander(t *testing.T, g *game.Game, me *game.Player, name string, commander bool) uuid.UUID {
	t.Helper()
	toMainForCost(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, IsCommander: commander})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func TestCmdrNornsChoirmasterProliferatesWhenACommanderEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdrPush(g, me.ID, "Norn's Choirmaster", "Creature — Phyrexian Angel", cmdrNornsChoirmasterOracle, 5, 4)
	counted := cmdrPush(g, me.ID, "Counted", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(counted, game.CounterPlusOne, 1) })

	cmdrCastCommander(t, g, me, "Plain Bear", false)
	passPriorityAroundTable(t, g)
	if proliferatePrompt(g) != nil || counterCount(g, counted, game.CounterPlusOne) != 1 {
		t.Fatal("a non-commander entering proliferated")
	}
	cmdrCastCommander(t, g, me, "Commander Bear", true)
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, counted)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, counted, game.CounterPlusOne); got != 2 {
		t.Errorf("a commander entering proliferated to %d counters, want 2", got)
	}
}

func TestCmdrNornsChoirmasterProliferatesWhenACommanderAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	cmdrPush(g, me.ID, "Norn's Choirmaster", "Creature — Phyrexian Angel", cmdrNornsChoirmasterOracle, 5, 4)
	counted := cmdrPush(g, me.ID, "Counted", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(counted, game.CounterPlusOne, 1) })
	cmdr := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Commander", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	passPriorityAroundTable(t, g)
	for proliferatePrompt(g) != nil {
		answerProliferate(t, g)
		passPriorityAroundTable(t, g)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cmdr, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, counted)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, counted, game.CounterPlusOne); got != 2 {
		t.Errorf("a commander attacking proliferated to %d counters, want 2", got)
	}
}

// --- Reya Dawnbringer ------------------------------------------------------------------

func TestCmdrReyaReturnsACreatureEachUpkeepIfYouMay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdrPush(g, me.ID, "Reya Dawnbringer", "Legendary Creature — Angel", cmdrReyaDawnbringerOracle, 4, 6)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	advanceToUpkeepOfSeat(t, g, 0)
	hobAnswerTargets(t, g, me.ID, dead)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) {
		t.Fatal("declining the may still returned the creature")
	}
}

func TestCmdrReyaReturnsTheCreatureWhenYouAccept(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdrPush(g, me.ID, "Reya Dawnbringer", "Legendary Creature — Angel", cmdrReyaDawnbringerOracle, 4, 6)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	rock := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Dead Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	advanceToUpkeepOfSeat(t, g, 0)
	hobAnswerTargets(t, g, me.ID, dead)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card did not return")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("a noncreature card returned")
	}
}
