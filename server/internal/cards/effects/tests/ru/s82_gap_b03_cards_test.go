package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b03_cards_test.go — S82 GAP batch 3 (#5948): Mystery Booster
// Commander creatures. Every test plays the real catalog card.

const (
	gapB03Glava        = "94d9a596-96b3-4fd6-b3f9-62920929c668"
	gapB03Zagorka      = "e444c824-6b9c-4b72-8aa9-a86489c47cbe"
	gapB03Elda         = "7c39639e-c611-46ee-beda-aef466170ba2"
	gapB03Jandor       = "8c05de43-6f9e-46c1-bc01-551a7fc570d0"
	gapB03AutumnWillow = "01f9cfd9-47da-4dbe-a21e-53eccad3864a"
	gapB03Grizzlegom   = "3f5db55a-8bae-4e97-86bd-f2a2dacf07a4"
	gapB03Rikala       = "45284633-8aab-4078-ae5c-7b0d75efe2f5"
	gapB03Maular       = "214f4e0a-eba7-4a08-86b2-d490d6ae1048"
	gapB03Meatsqueak   = "57994ace-3dd6-407f-b465-c70b552bd466"
	gapB03Selenia      = "7126be80-cebc-4868-b39d-8b215410f575"
	gapB03Blor         = "f5f2fa96-8e87-44bc-8c82-e217706d3cab"
)

func TestGapB03BlorHasItsThreeKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Blor the Impervious", "Legendary Creature — Frog Warrior", gapB03Blor, false)
	for _, kw := range []string{"trample", "hexproof", "indestructible"} {
		if !effectstest.HasEffectiveAbility(t, g, id, kw) {
			t.Errorf("Blor lacks %s", kw)
		}
	}
}

func gapB03Life(g *game.Game, id uuid.UUID) (n int) {
	g.ReadSnapshot(func() { n = effectstest.LifeOf(g, id) })
	return n
}

func gapB03Change(g *game.Game, id uuid.UUID, d int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, id, d) })
}

func TestGapB03SeleniaDoublesYourLifeGainOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Selenia, the Cursed Heart", "Legendary Creature — Angel Warrior", gapB03Selenia, false)
	before, oppBefore := gapB03Life(g, me.ID), gapB03Life(g, opp.ID)
	gapB03Change(g, me.ID, 3)
	gapB03Change(g, opp.ID, 3)
	if got := gapB03Life(g, me.ID); got != before+6 {
		t.Errorf("my life = %d, want %d", got, before+6)
	}
	if got := gapB03Life(g, opp.ID); got != oppBefore+3 {
		t.Errorf("opponent life = %d, want %d", got, oppBefore+3)
	}
	gapB03Change(g, me.ID, -2)
	if got := gapB03Life(g, me.ID); got != before+4 {
		t.Errorf("a loss was doubled: life %d, want %d", got, before+4)
	}
}

func TestGapB03SeleniaDyingCursesAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Selenia, the Cursed Heart", "Legendary Creature — Angel Warrior", gapB03Selenia, false)
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	var curse game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Selenia's Curse" {
				curse = c
			}
		}
	})
	if curse.InstanceID == uuid.Nil {
		t.Fatal("no Selenia's Curse on the battlefield")
	}
	if curse.Controller != me.ID || curse.AttachedTo.Kind != game.TargetPlayer {
		t.Fatalf("curse controller %v attached %+v", curse.Controller, curse.AttachedTo)
	}
	victim := curse.AttachedTo.ID
	if victim != opp.ID {
		t.Fatal("the curse does not enchant the chosen opponent")
	}
	before := gapB03Life(g, victim)
	gapB03Change(g, victim, -3)
	if got := gapB03Life(g, victim); got != before-6 {
		t.Errorf("enchanted player's life %d, want %d", got, before-6)
	}
	other := me.ID
	beforeMine := gapB03Life(g, other)
	gapB03Change(g, other, -3)
	if got := gapB03Life(g, other); got != beforeMine-3 {
		t.Errorf("an unenchanted player's loss was doubled: %d, want %d", got, beforeMine-3)
	}
}

func gapB03Count(g *game.Game, owner uuid.UUID, name string) (n int) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == owner {
				n++
			}
		}
	})
	return n
}

func TestGapB03MeatsqueakMakesOneFoodEachTurnAndScalesPerSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	meat := effectstest.PushCatalogPermanent(g, me.ID, "Meatsqueak, Hoard Lord", "Legendary Creature — Squirrel", gapB03Meatsqueak, false)
	bear1 := effectstest.B12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	bear2 := effectstest.B12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, me.ID, bear2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB03Count(g, me.ID, "Food"); n != 1 {
		t.Fatalf("Foods after two deaths in one turn = %d, want 1", n)
	}
	if p, tt := effectstest.PtOf(t, g, meat); p != 1 || tt != 1 {
		t.Errorf("P/T with one Food = %d/%d, want the fixture's 1/1", p, tt)
	}
	for i := 0; i < 5; i++ {
		effectstest.B12Permanent(g, me.ID, "Food", "Artifact — Food")
	}
	if p, tt := effectstest.PtOf(t, g, meat); p != 1 || tt != 1 {
		t.Errorf("P/T with six Foods = %d/%d, want the fixture's 1/1", p, tt)
	}
	effectstest.B12Permanent(g, me.ID, "Food", "Artifact — Food")
	if p, tt := effectstest.PtOf(t, g, meat); p != 4 || tt != 4 {
		t.Errorf("P/T with seven Foods = %d/%d, want 4/4 (1/1 and +3/+3)", p, tt)
	}
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tt := effectstest.PtOf(t, g, other); p != 2 || tt != 2 {
		t.Errorf("a non-Squirrel got the bonus: %d/%d", p, tt)
	}
}

func gapB03Creature(g *game.Game, owner uuid.UUID, name, oracle, cost string, p, tt int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Dinosaur Mutant", OracleID: oracle,
		ManaCost: cost, Power: p, Toughness: tt, Owner: owner, Controller: owner,
	})
}

func TestGapB03MaularDoublesAnAttackingBigCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maular := gapB03Creature(g, me.ID, "Maular, the Next Evolution", gapB03Maular, "{5}{G}{G}", 10, 10)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(maular, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(small, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, maular); p != 20 || tt != 20 {
		t.Errorf("Maular P/T = %d/%d, want 20/20", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, small); p != 2 || tt != 2 {
		t.Errorf("a mana value 0 attacker was doubled: %d/%d", p, tt)
	}
}

func gapB03Counters(g *game.Game, id uuid.UUID, kind string) (n int) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[kind]
			}
		}
	})
	return n
}

func TestGapB03RikalaCyclesTideCountersAndPumpsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rik := gapB03Creature(g, me.ID, "Rikala, Homarid King", gapB03Rikala, "{1}{G}{U}", 0, 4)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	const want = 1
	if n := gapB03Counters(g, rik, "tide"); n != want {
		t.Fatalf("after upkeep: %d tide counters, want %d", n, want)
	}
	if p, _ := effectstest.PtOf(t, g, ally); p != 2+want {
		t.Errorf("ally power = %d, want %d", p, 2+want)
	}
	if p, _ := effectstest.PtOf(t, g, rik); p != want {
		t.Errorf("Rikala power = %d, want %d", p, want)
	}
	if p, _ := effectstest.PtOf(t, g, foe); p != 2 {
		t.Errorf("an opponent's creature got the bonus: %d", p)
	}
}

func TestGapB03RikalaAddsAnExtraBlueFromAnIslandAtThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rik := gapB03Creature(g, me.ID, "Rikala, Homarid King", gapB03Rikala, "{1}{G}{U}", 0, 4)
	island := effectstest.B31Push(g, me.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.StAddCounters(t, g, rik, "tide", 2)
	effectstest.TapForMana(t, g, me.ID, island)
	if got := effectstest.SortedPool(me); len(got) != 1 {
		t.Fatalf("pool with two counters = %v, want just {U}", got)
	}
	effectstest.StAddCounters(t, g, rik, "tide", 1)
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 2 {
		t.Fatalf("a Forest added extra mana: %v", got)
	}
	island2 := effectstest.B31Push(g, me.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	effectstest.TapForMana(t, g, me.ID, island2)
	if got := effectstest.SortedPool(me); len(got) != 4 {
		t.Errorf("pool after the second Island at three counters = %v, want four mana", got)
	}
}

func TestGapB03RikalaResetsAtFourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rik := gapB03Creature(g, me.ID, "Rikala, Homarid King", gapB03Rikala, "{1}{G}{U}", 0, 4)
	effectstest.StAddCounters(t, g, rik, "tide", 3)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB03Counters(g, rik, "tide"); n != 0 {
		t.Errorf("tide counters after the fourth = %d, want 0", n)
	}
}

func TestGapB03GrizzlegomCountsEachLandTypeWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	griz := gapB03Creature(g, me.ID, "Grizzlegom, Hurloon Hero", gapB03Grizzlegom, "{5}{R}{G}", 4, 4)
	for _, l := range []struct {
		name string
		n    int
	}{{"Plains", 2}, {"Island", 1}, {"Swamp", 3}, {"Mountain", 2}, {"Forest", 4}} {
		for i := 0; i < l.n; i++ {
			effectstest.B31Push(g, me.ID, l.name, "Basic Land — "+l.name, "", "", 0, 0)
		}
	}
	effectstest.B31Push(g, opp.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	myLife, oppLife := gapB03Life(g, me.ID), gapB03Life(g, opp.ID)
	var handBefore int
	g.ReadSnapshot(func() { handBefore = len(me.Hand.Cards) })
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(griz, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB03Count(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("Soldiers = %d, want 2 (one per Plains)", n)
	}
	g.ReadSnapshot(func() {
		if got := len(me.Hand.Cards) - handBefore; got != 1 {
			t.Errorf("cards drawn = %d, want 1 (one per Island)", got)
		}
	})
	if got := gapB03Life(g, opp.ID); got != oppLife-3 {
		t.Errorf("opponent life %d, want %d (three of my Swamps)", got, oppLife-3)
	}
	if n := gapB03Counters(g, griz, game.CounterPlusOne); n != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 (one per Mountain)", n)
	}
	if got := gapB03Life(g, me.ID); got != myLife+4 {
		t.Errorf("my life %d, want %d (four Forests)", got, myLife+4)
	}
}

func TestGapB03AutumnWillowMakesADryadAndDoublesLandCreatureMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Autumn Willow, Harmony", "Legendary Creature — Avatar", gapB03AutumnWillow, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var dryad uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Forest Dryad" && c.Controller == me.ID {
				dryad = c.InstanceID
			}
		}
	})
	if dryad == uuid.Nil {
		t.Fatal("no Forest Dryad token on the battlefield")
	}
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 1 {
		t.Errorf("a plain Forest produced %v, want one {G}", got)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dryad {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.TapForMana(t, g, me.ID, dryad)
	if got := effectstest.SortedPool(me); len(got) != 3 {
		t.Errorf("pool after tapping the Dryad = %v, want three {G}", got)
	}
}

func TestGapB03JandorMakesHorsesMountsThatSaddleAndPayOffOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Jandor, Fortuned Traveler", "Legendary Creature — Human Citizen", gapB03Jandor, false)
	horse := effectstest.B12Creature(g, me.ID, "Horse", "Creature — Horse", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rider := effectstest.B12Creature(g, me.ID, "Rider", "Creature — Human", 2, 2)
	var hasMount bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == horse {
				hasMount = c.HasSubtype("Mount")
			}
			if c.InstanceID == bear && c.HasSubtype("Mount") {
				t.Error("a Bear became a Mount")
			}
		}
	})
	if !hasMount {
		t.Fatal("the Horse is not a Mount")
	}
	if err := effectstest.SaddleAbility(t, g, me.ID, bear, rider); err == nil {
		t.Error("a Bear could be saddled")
	}
	if err := effectstest.SaddleAbility(t, g, me.ID, horse, rider); err != nil {
		t.Fatalf("saddling the Horse: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	handBefore := len(me.Hand.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, horse)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, horse); p != 5 || tt != 5 {
		t.Errorf("saddled Horse P/T = %d/%d, want 5/5", p, tt)
	}
	if got := len(me.Hand.Cards) - handBefore; got != 1 {
		t.Errorf("cards drawn = %d, want 1", got)
	}
	if n := gapB03Count(g, me.ID, "Food"); n != 1 {
		t.Errorf("Food tokens = %d, want 1", n)
	}
}

func TestGapB03JandorDoesNothingForAnUnsaddledAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Jandor, Fortuned Traveler", "Legendary Creature — Human Citizen", gapB03Jandor, false)
	horse := effectstest.B12Creature(g, me.ID, "Horse", "Creature — Horse", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, horse)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, horse); p != 3 {
		t.Errorf("an unsaddled Horse got the bonus: %d", p)
	}
	if n := gapB03Count(g, me.ID, "Food"); n != 0 {
		t.Errorf("Food tokens = %d, want 0", n)
	}
}

func gapB03StackCopies(g *game.Game) map[uuid.UUID]int {
	copies := map[uuid.UUID]int{}
	for _, it := range g.StackMeta {
		if it != nil && it.IsCopy && it.Kind == game.StackItemSpell {
			copies[it.Controller]++
		}
	}
	return copies
}

func TestGapB03EldaGivesDemonstrateToTheFirstNonlegendaryCreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Elda, Conjurer of Spectacle", "Legendary Creature — Human Wizard", gapB03Elda, false)
	// A legendary creature spell is not the first nonlegendary one.
	effectstest.CastFromHandForTest(t, g, me, "Legend", "Legendary Creature — Human", "", "", game.CastSpellParams{})
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a legendary creature spell was offered demonstrate")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if c := gapB03StackCopies(g); c[me.ID] != 1 || c[opp.ID] != 1 || len(c) != 2 {
		t.Fatalf("copies by controller %v, want one for you and one for the chosen opponent", c)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB03Count(g, me.ID, "Bear"); n != 2 {
		t.Errorf("Bears under your control = %d, want 2 (the spell and your copy)", n)
	}
	if n := gapB03Count(g, opp.ID, "Bear"); n != 1 {
		t.Errorf("Bears under the opponent's control = %d, want 1", n)
	}
	// A second nonlegendary creature spell this turn has no demonstrate.
	effectstest.CastFromHandForTest(t, g, me, "Second Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("the second nonlegendary creature spell was offered demonstrate")
	}
}

func TestGapB03EldaDemonstrateMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Elda, Conjurer of Spectacle", "Legendary Creature — Human Wizard", gapB03Elda, false)
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c := gapB03StackCopies(g); len(c) != 0 {
		t.Errorf("a declined demonstrate made copies: %v", c)
	}
	if n := gapB03Count(g, me.ID, "Bear"); n != 1 {
		t.Errorf("Bears = %d, want 1", n)
	}
}

func gapB03Answer(t *testing.T, g *game.Game, p *game.Player, yes bool) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, p.ID)
	if c == nil {
		t.Fatalf("%s was not asked: %+v", p.Name, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, p.ID, yes); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
}

func TestGapB03ZagorkaOffersEachPlayerASanctumAndBarsTheOpponentsWhoTakeIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if len(g.Seats) < 4 {
		t.Skip("needs four seats")
	}
	mySeat := g.Turn.ActiveSeat
	me := g.Seats[mySeat]
	takes, declines, third := g.Seats[(mySeat+1)%4], g.Seats[(mySeat+2)%4], g.Seats[(mySeat+3)%4]
	effectstest.PushCatalogPermanent(g, me.ID, "Zagorka, Mother of Sanctum", "Legendary Creature — Human Peasant", gapB03Zagorka, false)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	gapB03Answer(t, g, me, true)
	gapB03Answer(t, g, takes, true)
	gapB03Answer(t, g, declines, false)
	gapB03Answer(t, g, third, true)
	for _, p := range []*game.Player{me, takes, third} {
		if n := gapB03Count(g, p.ID, "Sanctum"); n != 1 {
			t.Errorf("%s has %d Sanctum, want 1", p.Name, n)
		}
	}
	if n := gapB03Count(g, declines.ID, "Sanctum"); n != 0 {
		t.Errorf("a player who declined got a Sanctum")
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Sanctum" && !c.Tapped {
				t.Errorf("a Sanctum entered untapped")
			}
		}
		if effectstest.TrekB48Restricted(me, me.ID) || effectstest.TrekB48Restricted(declines, me.ID) {
			t.Error("the controller or the player who declined is restricted")
		}
		if !effectstest.TrekB48Restricted(takes, me.ID) || !effectstest.TrekB48Restricted(third, me.ID) {
			t.Error("an opponent who took a Sanctum is not restricted")
		}
	})
	raider := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: takes.ID, Controller: takes.ID,
	})
	effectstest.TrekB48ToAttackersOf(t, g, (mySeat+1)%4)
	var pe *game.PlayerCantAttackError
	if err := g.Clone().DeclareAttacker(raider, me.ID); !errors.As(err, &pe) {
		t.Errorf("attack at Zagorka's controller: err = %v, want *PlayerCantAttackError", err)
	}
	if err := g.Clone().DeclareAttacker(raider, declines.ID); err != nil {
		t.Errorf("attack at another player: %v", err)
	}
}

func TestGapB03SanctumTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tok := effects.SanctumToken()
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = me.ID, me.ID
	g.Battlefield.PushTop(tok)
	if err := g.ActivateManaAbility(me.ID, tok.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Sanctum: %v", err)
	}
	var pending *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			pending = c
		}
	}
	if pending == nil {
		t.Fatal("no colour choice queued for the any-colour Sanctum")
	}
	if err := g.ResolveManaChoice(pending.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "U" {
		t.Errorf("mana pool = %+v, want one {U}", me.ManaPool)
	}
	if !g.Battlefield.Contains(tok.InstanceID) {
		t.Error("tapping the Sanctum sacrificed it")
	}
}

// gapB03StackX is the X of the topmost stack item sourced from id.
func gapB03StackX(g *game.Game, id uuid.UUID) (x int, found bool) {
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == id && it.Kind != game.StackItemTriggered {
			return it.XValue, true
		}
	}
	return 0, false
}

func TestGapB03GlavaSetsXToFiveForAPermanentSpellAndAnAbilityOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Glava, Five-Advents Mage", "Legendary Creature — Human Wizard", gapB03Glava, false)
	spell := effectstest.CastFromHandForTest(t, g, me, "Xbot", "Artifact Creature — Construct", "{X}{G}", "", game.CastSpellParams{XValue: 2})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if x, ok := gapB03StackX(g, spell); !ok || x != 5 {
		t.Fatalf("the spell's X = %d (found %v), want 5", x, ok)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Only once each turn: the second X thing is not offered.
	archon := effectstest.PushCatalogPermanent(g, me.ID, "Vengeful Archon", "Creature — Archon", effectstest.Pr8VengefulArchon, false)
	if err := g.ActivateCatalogAbility(me.ID, archon, 0, game.ActivateAbilityParams{XValue: 1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("a second use was offered in the same turn")
	}
	if x, ok := gapB03StackX(g, archon); !ok || x != 1 {
		t.Errorf("the ability's X = %d (found %v), want the paid 1", x, ok)
	}
}

func TestGapB03GlavaSetsXOfAnActivatedAbilityAndADeclineKeepsTheTurnsUse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Glava, Five-Advents Mage", "Legendary Creature — Human Wizard", gapB03Glava, false)
	archon := effectstest.PushCatalogPermanent(g, me.ID, "Vengeful Archon", "Creature — Archon", effectstest.Pr8VengefulArchon, false)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}
	if err := g.ActivateCatalogAbility(me.ID, archon, 0, game.ActivateAbilityParams{XValue: 1, Targets: target}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	if x, _ := gapB03StackX(g, archon); x != 1 {
		t.Errorf("a declined offer changed X to %d", x)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, archon, 0, game.ActivateAbilityParams{XValue: 1, Targets: target}); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if x, ok := gapB03StackX(g, archon); !ok || x != 5 {
		t.Errorf("the ability's X = %d (found %v), want 5", x, ok)
	}
}

func TestGapB03GlavaIgnoresAnInstantWithX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Glava, Five-Advents Mage", "Legendary Creature — Human Wizard", gapB03Glava, false)
	effectstest.CastFromHandForTest(t, g, me, "Xfire", "Instant", "{X}{R}", "", game.CastSpellParams{XValue: 2})
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("an instant spell with X was offered")
	}
}
