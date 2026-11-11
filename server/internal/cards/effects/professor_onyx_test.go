package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const professorOnyxOracle = "9df2e909-ed13-456a-9636-7398732009a9"

func TestProfessorOnyxMinusThreeSacrificesTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWalkerForTest(g, me.ID, "Professor Onyx", professorOnyxOracle, 5)
	small := pushCreatureToBattlefieldForTest(g, opp.ID, "Small")
	big := pushCreatureToBattlefieldForTest(g, opp.ID, "Big")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case small:
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 1
			case big:
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 5, 5
			}
		}
	})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)
	var haveSmall, haveBig bool
	for _, c := range g.Battlefield.Cards {
		haveSmall = haveSmall || c.InstanceID == small
		haveBig = haveBig || c.InstanceID == big
	}
	if haveBig || !haveSmall {
		t.Errorf("small on battlefield=%v big=%v; want the 5-power creature sacrificed only", haveSmall, haveBig)
	}
}

func TestProfessorOnyxPlusOneLosesLifeAndTakesOne(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := pushWalkerForTest(g, me.ID, "Professor Onyx", professorOnyxOracle, 5)
	advanceToMainOf(t, g, seat)
	hand, life, gy := handSize(me), lifeOf(g, me.ID), me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 {
		t.Fatalf("want a prompt over the top three cards, got %+v", pick)
	}
	answerChooseCards(t, g, me.ID, pick.ChooseCards[0])
	if lifeOf(g, me.ID) != life-1 {
		t.Errorf("life = %d, want %d", lifeOf(g, me.ID), life-1)
	}
	if handSize(me) != hand+1 || me.Graveyard.Size() != gy+2 {
		t.Errorf("hand %d (want %d), graveyard %d (want %d)", handSize(me), hand+1, me.Graveyard.Size(), gy+2)
	}
}

func TestProfessorOnyxMinusEightEmptyHandsLoseThreeSevenTimes(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := pushWalkerForTest(g, me.ID, "Professor Onyx", professorOnyxOracle, 8)
	var opps []*game.Player
	for i := 1; i < len(g.Seats); i++ {
		o := g.Seats[(seat+i)%len(g.Seats)]
		emptyHandToLibrary(g, o)
		opps = append(opps, o)
	}
	advanceToMainOf(t, g, seat)
	before := lifeOf(g, me.ID)
	var lives []int
	for _, o := range opps {
		lives = append(lives, lifeOf(g, o.ID))
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, o := range opps {
		if got := lifeOf(g, o.ID); got != lives[i]-21 {
			t.Errorf("opponent %d life = %d, want %d", i, got, lives[i]-21)
		}
	}
	if lifeOf(g, me.ID) != before {
		t.Error("the controller lost life")
	}
}

func TestProfessorOnyxMinusEightAsksEachRoundAndDiscardSavesTheLife(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := pushWalkerForTest(g, me.ID, "Professor Onyx", professorOnyxOracle, 8)
	asked := g.Seats[(seat+1)%len(g.Seats)]
	for i := 2; i < len(g.Seats); i++ {
		emptyHandToLibrary(g, g.Seats[(seat+i)%len(g.Seats)])
	}
	advanceToMainOf(t, g, seat)
	handBefore, lifeBefore := handSize(asked), lifeOf(g, asked.ID)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The two other opponents have empty hands, so their rounds are
	// settled without a question; the asked opponent is queried seven times.
	for round := 0; round < 7; round++ {
		pick := chooseCardsChoiceFor(g, asked.ID)
		if pick == nil {
			t.Fatalf("round %d: no discard prompt", round)
		}
		if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
			t.Fatalf("round %d: bounds %d..%d, want 0..1", round, pick.ChooseMin, pick.ChooseMax)
		}
		if round < 2 {
			answerChooseCards(t, g, asked.ID, pick.ChooseCards[0])
		} else {
			answerChooseCards(t, g, asked.ID)
		}
	}
	if chooseCardsChoiceFor(g, asked.ID) != nil {
		t.Fatal("an eighth prompt opened")
	}
	if got := handSize(asked); got != handBefore-2 {
		t.Errorf("hand = %d, want %d", got, handBefore-2)
	}
	if got := lifeOf(g, asked.ID); got != lifeBefore-15 {
		t.Errorf("life = %d, want %d (five declines of 3)", got, lifeBefore-15)
	}
}

func TestProfessorOnyxMagecraftDrains(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushWalkerForTest(g, me.ID, "Professor Onyx", professorOnyxOracle, 5)
	advanceToMainOf(t, g, seat)
	before := lifeOf(g, me.ID)
	var lives []int
	for _, p := range g.Seats {
		lives = append(lives, lifeOf(g, p.ID))
	}
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 2
		if p.ID == me.ID {
			want = before + 2
		}
		if got := lifeOf(g, p.ID); got != want {
			t.Errorf("seat %d life = %d, want %d", i, got, want)
		}
	}
}
