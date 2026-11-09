package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const brigidOracle = "60fc89a1-cc6f-4a32-84c3-09abf7a4a1df"

func brigidRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Brigid, Clachan's Heart // Brigid, Doun's Mind",
		Layout:   "transform",
		TypeLine: "Legendary Creature — Kithkin Warrior // Legendary Creature — Kithkin Soldier",
		OracleID: uuid.MustParse(brigidOracle),
		Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Brigid, Clachan's Heart", TypeLine: "Legendary Creature — Kithkin Warrior", ManaCost: "{2}{W}", Colors: []string{"W"}, Power: "3", Toughness: "2",
				OracleText: "Whenever this creature enters or transforms into Brigid, Clachan's Heart, create a 1/1 green and white Kithkin creature token.\nAt the beginning of your first main phase, you may pay {G}. If you do, transform Brigid."},
			{Name: "Brigid, Doun's Mind", TypeLine: "Legendary Creature — Kithkin Soldier", Colors: []string{"G", "W"}, Power: "3", Toughness: "2",
				OracleText: "{T}: Add X {G} or X {W}, where X is the number of other creatures you control.\nAt the beginning of your first main phase, you may pay {W}. If you do, transform Brigid."},
		},
	}
}

// brigidPut puts the imported card onto p's battlefield, front face up,
// without entering triggers and not summoning sick.
func brigidPut(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(brigidRow(), p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatal("Brigid did not reach the hand")
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

func TestBrigidIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{brigidOracle, brigidOracle + "#1"} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Fatalf("%s: spec = %+v, %v", oracle, spec, ok)
		}
	}
}

// Entering makes a 1/1 green and white Kithkin.
func TestBrigidClachansHeartMakesAKithkinAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	importAndCast(t, g, brigidRow(), me)
	monSettle(t, g)
	if n := rfdCountNamed(g, me.ID, "Kithkin"); n != 1 {
		t.Fatalf("%d Kithkin tokens, want 1", n)
	}
	c := layeredCard(t, g, findBattlefieldByName(g, "Kithkin"))
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.HasColor("G") || !c.HasColor("W") {
		t.Errorf("the token is not a 1/1 green and white Kithkin: %+v", c)
	}
}

// Declining the {G} keeps Brigid on her front face.
func TestBrigidClachansHeartDeclinedStaysFrontFace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := brigidPut(t, g, me)
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	answerPayUnless(t, g, me.ID, false)
	monSettle(t, g)
	if f := iwFace(g, id); f != 0 {
		t.Fatalf("Brigid turned over without the {G} paid (face %d)", f)
	}
}

// Paying {G} turns Brigid over; Doun's Mind taps for X of one of green
// or white, X the other creatures; paying {W} next first main phase
// turns her back, which makes a Kithkin.
func TestBrigidTransformsTapsForXAndTransformsBack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	id := brigidPut(t, g, me)
	b12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	if f := iwFace(g, id); f != 1 {
		t.Fatalf("Brigid is on face %d after paying {G}, want the back face", f)
	}
	if n := rfdCountNamed(g, me.ID, "Kithkin"); n != 0 {
		t.Fatalf("turning into Doun's Mind made %d Kithkin", n)
	}

	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Doun's Mind: %v", err)
	}
	ashlingPickColor(t, g, me.ID, "W")
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool holds %d mana, want 2 (two other creatures you control)", got)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "W" {
			t.Errorf("pool token %+v, want white", tok)
		}
	}

	advanceToStepOf(t, g, seat, game.StepUpkeep)
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	if f := iwFace(g, id); f != 0 {
		t.Fatalf("Brigid is on face %d after paying {W}, want the front face", f)
	}
	if n := rfdCountNamed(g, me.ID, "Kithkin"); n != 1 {
		t.Errorf("transforming into Clachan's Heart made %d Kithkin, want 1", n)
	}
}
