package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eureka Moment draws two, then offers the land drop: only lands are
// candidates, the pick enters untapped, and the turn's land play is
// still available afterwards.
func TestEurekaMomentPutsALandFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	rock := effectstest.HandCardForTest(me, "Sol Ring", "Artifact", "")
	bolt := effectstest.HandCardForTest(me, "Lightning Bolt", "Instant", "")

	effectstest.CastCatalogSpell(t, g, "Eureka Moment", "Instant", effectstest.PutFromHandEurekaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	if !effectstest.HasID(pick.ChooseCards, land) {
		t.Fatal("the land in hand was not offered")
	}
	if pick.ChooseDestination != game.ChooseOntoBattlefield {
		t.Errorf("destination %q, want %q: the bot prices the put by it rather than as a discard (#2680)",
			pick.ChooseDestination, game.ChooseOntoBattlefield)
	}

	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the chosen land did not reach the battlefield")
	}
	if me.Hand.Contains(land) {
		t.Error("the land is on the battlefield and still in hand")
	}
	if effectstest.TappedOnBattlefield(t, g, land) {
		t.Error("nothing said tapped")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d, want 0: this was a put, not a play (CR 305.4)", n)
	}
}

// "You MAY" — declining is a real answer, and it moves nothing.
func TestEurekaMomentDeclinedPutsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Eureka Moment", "Instant", effectstest.PutFromHandEurekaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	before := g.Battlefield.Size()
	effectstest.AnswerChooseCards(t, g, me.ID)
	if g.Battlefield.Size() != before {
		t.Error("a declined \"you may\" put something onto the battlefield")
	}
	if !me.Hand.Contains(land) {
		t.Error("the land left the hand anyway")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the prompt is still open after being answered")
	}
}

// With no land in hand there is nothing to ask about, so no prompt is
// opened — a click the player has no choice about is a bug, not a
// feature.
func TestEurekaMomentWithNoLandInHandAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.HandCardForTest(me, "Sol Ring", "Artifact", "")
	effectstest.CastCatalogSpell(t, g, "Eureka Moment", "Instant", effectstest.PutFromHandEurekaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a hand with no land was still prompted")
	}
}

// The land goes through the real entry, so its own clauses apply:
// Khalni Garden declares SelfEntersTapped() and an enters trigger, so
// it arrives TAPPED — never tapped afterwards, so no tap event — and
// its Plant token trigger uses the stack like any other ETB.
func TestPutFromHandRunsTheLandsOwnEntryClauses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	garden := effectstest.HandCardForTest(me, "Khalni Garden", "Land", effectstest.PutFromHandKhalniOracle)
	effectstest.CastCatalogSpell(t, g, "Eureka Moment", "Instant", effectstest.PutFromHandEurekaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	before := len(g.Events)
	effectstest.AnswerChooseCards(t, g, me.ID, garden)
	if !effectstest.TappedOnBattlefield(t, g, garden) {
		t.Error("Khalni Garden's own \"enters tapped\" did not apply to the put")
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventTapCard && ev.CardID == garden {
			t.Error("the land must ARRIVE tapped, not be tapped after entering")
		}
	}

	plants := effectstest.B30TokensNamed(g, me.ID, "Plant")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B30TokensNamed(g, me.ID, "Plant") != plants+1 {
		t.Error("Khalni Garden's enters trigger did not fire on the put")
	}
}

// Insidious Fungus's third mode is the same clause with the printed
// "tapped" rider.
func TestInsidiousFungusPutsALandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	fungus := effectstest.B12Push(g, me.ID, "Insidious Fungus", "Creature — Fungus", effectstest.PutFromHandFungusOracle, 1, 2)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, fungus, 2, game.ActivateAbilityParams{})

	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick == nil || pick.ChooseDestination != game.ChooseOntoBattlefieldTapped {
		t.Fatalf("prompt = %+v, want destination %q: the bot prices the put by it (#2680)", pick, game.ChooseOntoBattlefieldTapped)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not reach the battlefield")
	}
	if !effectstest.TappedOnBattlefield(t, g, land) {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := effectstest.HandCardForTest(me, "Some Land", tc.typeLine, "")
			life := me.Life
			hand := me.Hand.Size()

			effectstest.CastCatalogSpell(t, g, "Spelunking", "Enchantment", effectstest.PutFromHandSpelunkingOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			// The Spelunking that was seeded and cast is net zero, so
			// the entry trigger's draw is the whole difference.
			if me.Hand.Size() != hand+1 {
				t.Errorf("hand %d → %d: the entry trigger draws one", hand, me.Hand.Size())
			}

			effectstest.AnswerChooseCards(t, g, me.ID, land)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.HandCardForTest(me, "Some Cave", "Land — Cave", "")
	effectstest.CastCatalogSpell(t, g, "Spelunking", "Enchantment", effectstest.PutFromHandSpelunkingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	life := me.Life
	effectstest.AnswerChooseCards(t, g, me.ID)
	if me.Life != life {
		t.Errorf("life %d → %d: a declined offer gains nothing", life, me.Life)
	}
}

// Broken Bond destroys its target and then offers the land drop.
func TestBrokenBondDestroysThenPutsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")

	effectstest.CastCatalogSpell(t, g, "Broken Bond", "Sorcery", effectstest.PutFromHandBrokenBondOracle, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}

	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("the land drop did not happen")
	}
}

// Chulane's cast trigger draws, then offers the land — the whole
// ability, in printed order, off one trigger.
func TestChulaneDrawsThenOffersTheLandDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chulane := effectstest.B12Push(g, me.ID, "Chulane, Teller of Tales", "Legendary Creature — Human Druid", chulaneOracleForPutTest, 2, 4)
	_ = chulane
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")

	hand := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	// The Bear that was seeded and cast is net zero, so the trigger's
	// draw is the whole difference.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: the trigger draws one", hand, me.Hand.Size())
	}

	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("Chulane's land drop did not happen")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d, want 0: Chulane ramps on top of the land drop", n)
	}
}

const chulaneOracleForPutTest = "ebf7ce9b-9e5e-4557-9e28-76556997f0ee"

// A token is not a permanent CARD (CR 108.2) and can't come back onto
// the battlefield (CR 111.8), so the hand clause never offers one —
// the same test the library clause applies.
func TestHandCardsMatchingSkipsTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	token := effectstest.HandCardForTest(me, "Forest Token", "Token Land — Forest", "")
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	var got []uuid.UUID
	g.WithWriteLock(func() { got = effects.HandCardsMatching(g, me.ID, nil) })
	if effectstest.HasID(got, token) {
		t.Error("a token in hand was offered as a permanent card")
	}
	if !effectstest.HasID(got, land) {
		t.Error("the land card in hand was not offered")
	}
}
