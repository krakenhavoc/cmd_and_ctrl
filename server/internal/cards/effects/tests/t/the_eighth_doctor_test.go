package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const eighthDoctorOracle = "0956bc16-ff04-4e96-8059-ef62afa2405f"

func TestTheEighthDoctorIsRegisteredFull(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, eighthDoctorOracle, "The Eighth Doctor")
}

func eighthGy(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

func eighthCast(g *game.Game, p *game.Player, id uuid.UUID) error {
	return g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true})
}

// The enters trigger mills three.
func TestTheEighthDoctorMillsThreeOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 5; i++ {
		effectstest.TopOfLibrary(me, "Card", "Sorcery", "{1}")
	}
	before := len(me.Graveyard.Cards)
	effectstest.CastCatalogSpell(t, g, "The Eighth Doctor", "Legendary Creature — Time Lord Doctor", eighthDoctorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Graveyard.Cards) - before; got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
}

// A historic permanent spell is cast from the graveyard once this turn;
// a nonhistoric one and a second one are refused. The cast permanent
// gains the replacement: destroyed, it is exiled instead.
func TestTheEighthDoctorCastsAHistoricPermanentOnceAndItIsExiledWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "The Eighth Doctor", "Legendary Creature — Time Lord Doctor", eighthDoctorOracle, false)
	relic := eighthGy(g, me, game.Card{Name: "Relic", TypeLine: "Artifact", ManaCost: "{1}"})
	bear := eighthGy(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2})
	hero := eighthGy(g, me, game.Card{Name: "Hero", TypeLine: "Legendary Creature — Human", ManaCost: "{1}", Power: 2, Toughness: 2})
	instant := eighthGy(g, me, game.Card{Name: "Zap", TypeLine: "Legendary Instant", ManaCost: "{1}"})
	effectstest.FloatForTest(g, me, "C")
	if err := eighthCast(g, me, bear); err == nil {
		t.Fatal("cast a nonhistoric creature from the graveyard")
	}
	if err := eighthCast(g, me, instant); err == nil {
		t.Fatal("cast a nonpermanent spell from the graveyard")
	}
	if err := eighthCast(g, me, relic); err != nil {
		t.Fatalf("cast the historic artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.E2Zone(g, relic); z != game.ZoneBattlefield {
		t.Fatalf("the artifact is in %q", z)
	}
	effectstest.FloatForTest(g, me, "C")
	if err := eighthCast(g, me, hero); err == nil {
		t.Fatal("a second graveyard cast this turn")
	}
	effectstest.P1Destroy(t, g, relic)
	if z := effectstest.E2Zone(g, relic); z != game.ZoneExile {
		t.Fatalf("the destroyed artifact is in %q, want exile", z)
	}
}

// A historic (legendary) land is played from the graveyard, gains the
// replacement and spends the turn's use; bounced, it is exiled instead.
func TestTheEighthDoctorPlaysAHistoricLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "The Eighth Doctor", "Legendary Creature — Time Lord Doctor", eighthDoctorOracle, false)
	plain := eighthGy(g, me, game.Card{Name: "Plain Forest", TypeLine: "Basic Land — Forest"})
	land := eighthGy(g, me, game.Card{Name: "Hall", TypeLine: "Legendary Land"})
	hero := eighthGy(g, me, game.Card{Name: "Hero", TypeLine: "Legendary Creature — Human", ManaCost: "{1}", Power: 2, Toughness: 2})
	if err := eighthCast(g, me, plain); err == nil {
		t.Fatal("played a nonhistoric land from the graveyard")
	}
	if err := eighthCast(g, me, land); err != nil {
		t.Fatalf("play the legendary land: %v", err)
	}
	effectstest.FloatForTest(g, me, "C")
	if err := eighthCast(g, me, hero); err == nil {
		t.Fatal("cast after the turn's use was spent on a land")
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(land); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if z := effectstest.E2Zone(g, land); z != game.ZoneExile {
		t.Fatalf("the bounced land is in %q, want exile", z)
	}
}
