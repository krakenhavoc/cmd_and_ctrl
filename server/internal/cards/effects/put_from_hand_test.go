package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// put_from_hand_test.go — the catalog half of #654: the shared
// MayPutALandFromHand clause and the five cards that print it.
//
// The engine half (the move itself, the CR 614 pipeline, the land
// drop that isn't one) is pinned in
// server/internal/game/put_from_hand_test.go. What these tests are
// about is the PROMPT: who is offered what, what a decline does, and
// that the card that arrives went through a real entry.

const (
	putFromHandEurekaOracle     = "0e2c11b2-d95f-4402-9a4a-afd3f7ffb8be"
	putFromHandBrokenBondOracle = "858e12e9-3eaa-40cf-9e22-f9ccdfe485b3"
	putFromHandFungusOracle     = "0a8d0217-ff24-4177-b6be-707eb2b6b9e9"
	putFromHandSpelunkingOracle = "2962fe4c-bf48-454b-8a6b-0f8253352ae8"
	putFromHandKhalniOracle     = "b2d5ba45-8674-4428-89db-c2bbbf0bf5c5"
)

// handCardForTest seeds one card into a player's hand.
func handCardForTest(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// answerChooseCards answers the open card-set prompt for a player.
func answerChooseCards(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	pick := chooseCardsChoiceFor(g, chooser)
	if pick == nil {
		t.Fatal("no choose-cards prompt is open")
	}
	if err := g.ResolveChooseCards(pick.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// Eureka Moment draws two, then offers the land drop: only lands are
// candidates, the pick enters untapped, and the turn's land play is
// still available afterwards.
func TestEurekaMomentPutsALandFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	rock := handCardForTest(me, "Sol Ring", "Artifact", "")
	bolt := handCardForTest(me, "Lightning Bolt", "Instant", "")

	castCatalogSpell(t, g, "Eureka Moment", "Instant", putFromHandEurekaOracle, nil)
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Eureka Moment did not offer the land drop")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 — \"you MAY put A land card\"", pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == rock || id == bolt {
			t.Error("only land cards are candidates")
		}
	}
	if !hasID(pick.ChooseCards, land) {
		t.Fatal("the land in hand was not offered")
	}

	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the chosen land did not reach the battlefield")
	}
	if me.Hand.Contains(land) {
		t.Error("the land is on the battlefield and still in hand")
	}
	if tappedOnBattlefield(t, g, land) {
		t.Error("nothing said tapped")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d, want 0: this was a put, not a play (CR 305.4)", n)
	}
}

// "You MAY" — declining is a real answer, and it moves nothing.
func TestEurekaMomentDeclinedPutsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	castCatalogSpell(t, g, "Eureka Moment", "Instant", putFromHandEurekaOracle, nil)
	passPriorityAroundTable(t, g)

	before := g.Battlefield.Size()
	answerChooseCards(t, g, me.ID)
	if g.Battlefield.Size() != before {
		t.Error("a declined \"you may\" put something onto the battlefield")
	}
	if !me.Hand.Contains(land) {
		t.Error("the land left the hand anyway")
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the prompt is still open after being answered")
	}
}

// With no land in hand there is nothing to ask about, so no prompt is
// opened — a click the player has no choice about is a bug, not a
// feature.
func TestEurekaMomentWithNoLandInHandAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	handCardForTest(me, "Sol Ring", "Artifact", "")
	castCatalogSpell(t, g, "Eureka Moment", "Instant", putFromHandEurekaOracle, nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a hand with no land was still prompted")
	}
}

// The land goes through the real entry, so its own clauses apply:
// Khalni Garden declares SelfEntersTapped() and an enters trigger, so
// it arrives TAPPED — never tapped afterwards, so no tap event — and
// its Plant token trigger uses the stack like any other ETB.
func TestPutFromHandRunsTheLandsOwnEntryClauses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	garden := handCardForTest(me, "Khalni Garden", "Land", putFromHandKhalniOracle)
	castCatalogSpell(t, g, "Eureka Moment", "Instant", putFromHandEurekaOracle, nil)
	passPriorityAroundTable(t, g)

	before := len(g.Events)
	answerChooseCards(t, g, me.ID, garden)
	if !tappedOnBattlefield(t, g, garden) {
		t.Error("Khalni Garden's own \"enters tapped\" did not apply to the put")
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventTapCard && ev.CardID == garden {
			t.Error("the land must ARRIVE tapped, not be tapped after entering")
		}
	}

	plants := b30TokensNamed(g, me.ID, "Plant")
	passPriorityAroundTable(t, g)
	if b30TokensNamed(g, me.ID, "Plant") != plants+1 {
		t.Error("Khalni Garden's enters trigger did not fire on the put")
	}
}

// Insidious Fungus's third mode is the same clause with the printed
// "tapped" rider.
func TestInsidiousFungusPutsALandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	fungus := b12Push(g, me.ID, "Insidious Fungus", "Creature — Fungus", putFromHandFungusOracle, 1, 2)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, fungus, 2, game.ActivateAbilityParams{})

	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not reach the battlefield")
	}
	if !tappedOnBattlefield(t, g, land) {
		t.Error("the third mode puts the land onto the battlefield TAPPED")
	}
}

// Spelunking: the ETB draws, then offers the land, and the Cave rider
// rides the continuation — it pays only when a Cave actually entered.
func TestSpelunkingPaysTheCaveLifeOnlyForACave(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		wantLife int
	}{
		{"a Cave", "Land — Cave", 4},
		{"an ordinary land", "Basic Land — Forest", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			land := handCardForTest(me, "Some Land", tc.typeLine, "")
			life := me.Life
			hand := me.Hand.Size()

			castCatalogSpell(t, g, "Spelunking", "Enchantment", putFromHandSpelunkingOracle, nil)
			passPriorityAroundTable(t, g)
			// The Spelunking that was seeded and cast is net zero, so
			// the entry trigger's draw is the whole difference.
			if me.Hand.Size() != hand+1 {
				t.Errorf("hand %d → %d: the entry trigger draws one", hand, me.Hand.Size())
			}

			answerChooseCards(t, g, me.ID, land)
			if !g.Battlefield.Contains(land) {
				t.Fatal("the chosen land did not enter")
			}
			if got := me.Life - life; got != tc.wantLife {
				t.Errorf("life %+d, want %+d", got, tc.wantLife)
			}
		})
	}
}

// Declining Spelunking's offer pays no life and puts nothing: the
// rider reads the permanent that entered, and none did.
func TestSpelunkingDeclinedPaysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	handCardForTest(me, "Some Cave", "Land — Cave", "")
	castCatalogSpell(t, g, "Spelunking", "Enchantment", putFromHandSpelunkingOracle, nil)
	passPriorityAroundTable(t, g)

	life := me.Life
	answerChooseCards(t, g, me.ID)
	if me.Life != life {
		t.Errorf("life %d → %d: a declined offer gains nothing", life, me.Life)
	}
}

// Broken Bond destroys its target and then offers the land drop.
func TestBrokenBondDestroysThenPutsALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")

	castCatalogSpell(t, g, "Broken Bond", "Sorcery", putFromHandBrokenBondOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}

	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("the land drop did not happen")
	}
}

// Chulane's cast trigger draws, then offers the land — the whole
// ability, in printed order, off one trigger.
func TestChulaneDrawsThenOffersTheLandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chulane := b12Push(g, me.ID, "Chulane, Teller of Tales", "Legendary Creature — Human Druid", chulaneOracleForPutTest, 2, 4)
	_ = chulane
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")

	hand := me.Hand.Size()
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	// The Bear that was seeded and cast is net zero, so the trigger's
	// draw is the whole difference.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: the trigger draws one", hand, me.Hand.Size())
	}

	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("Chulane's land drop did not happen")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d, want 0: Chulane ramps on top of the land drop", n)
	}
}

const chulaneOracleForPutTest = "ebf7ce9b-9e5e-4557-9e28-76556997f0ee"

// --- small local helpers -------------------------------------------

func tappedOnBattlefield(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return false
}
