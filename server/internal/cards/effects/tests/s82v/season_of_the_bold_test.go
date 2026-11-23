package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const seasonOfTheBoldOracle = "5fe29c25-a7ad-4c79-a5a5-da1bbc832141"

func TestSeasonOfTheBoldIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(seasonOfTheBoldOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Season of the Bold: %+v %v", spec, ok)
	}
}

func castSeasonOfTheBold(t *testing.T, g *game.Game, modes []int) error {
	t.Helper()
	_, err := effectstest.CastPawprintSpell(t, g, "Season of the Bold", "Sorcery", "{3}{R}{R}", "{C}{C}{C}{R}{R}", seasonOfTheBoldOracle, modes, nil)
	return err
}

func TestSeasonOfTheBoldTappedTreasureAndImpulseExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Top Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Second", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "Third", TypeLine: "Sorcery", ManaCost: "{R}"},
	)
	if err := castSeasonOfTheBold(t, g, []int{0, 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	treasures, tapped := effectstest.CountTokensNamed(g, me.ID, "Treasure")
	if treasures != 1 || tapped != 1 {
		t.Errorf("Treasures %d (tapped %d), want one tapped", treasures, tapped)
	}
	if !effectstest.InExile(g, ids[0]) || !effectstest.InExile(g, ids[1]) || effectstest.InExile(g, ids[2]) {
		t.Error("the top two cards, and only those, should be exiled")
	}
	// Playable: the land can be played from exile.
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("could not play the exiled land: %v", err)
	}
}

func TestSeasonOfTheBoldDamagesUpToOneTargetCreatureWheneverYouCastUntilEndOfYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castSeasonOfTheBold(t, g, []int{2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Casting any spell fires the delayed trigger, with its own target.
	cast := func() {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cub", TypeLine: "Creature — Bear", ManaCost: "{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		})
		effectstest.FloatForTest(g, me, "R")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
	}
	cast()
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 2 damage")
	}
	// Still active: another spell fires it again, with a fresh target.
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	cast()
	effectstest.AnswerPickTarget(t, g, big)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, big)
	if c.DamageMarked != 2 {
		t.Errorf("the Giant has %d damage marked, want 2 from the second cast", c.DamageMarked)
	}
}

func TestSeasonOfTheBoldRefusesSixPawprints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := castSeasonOfTheBold(t, g, []int{2, 2}); err == nil {
		t.Error("six pawprints were accepted")
	}
}
