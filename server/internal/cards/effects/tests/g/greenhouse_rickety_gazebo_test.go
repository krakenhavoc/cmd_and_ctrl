package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// landMakesAnyColor reports whether the land has an any-color tap
// ability it was granted.
func landMakesAnyColor(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(id)
		abs, origins := game.ManaAbilitiesWithOrigins(c)
		for i, ab := range abs {
			if ab.Produced == "{W|U|B|R|G}" && origins.At(i).Granted() {
				found = true
			}
		}
	})
	return found
}

// Greenhouse gives your lands (and only yours) an any-color tap
// ability, and only while its door is unlocked.
func TestGreenhouseGrantsAnyColorToYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B16Creature(g, me.ID, "Swamp", "Basic Land — Swamp", 0, 0)
	theirs := effectstest.B16Creature(g, opp.ID, "Swamp", "Basic Land — Swamp", 0, 0)

	// Casting the Rickety Gazebo half: the Greenhouse door is locked.
	effectstest.RoomsDCast(t, g, me, effectstest.GreenhouseCard(me.ID), 1)
	if landMakesAnyColor(g, mine) {
		t.Fatal("the Greenhouse door is locked but a land has its grant")
	}
	effectstest.RoomsDUnlock(t, g, me, effectstest.FindBattlefieldByName(g, "Rickety Gazebo"), game.DoorLeft)
	if !landMakesAnyColor(g, mine) {
		t.Error("an unlocked Greenhouse did not grant the any-color ability to my land")
	}
	if landMakesAnyColor(g, theirs) {
		t.Error("an opponent's land got the grant")
	}
}

// Rickety Gazebo mills four, then returns up to two permanent cards
// from among them (not the nonpermanent one, not a card already in the
// graveyard).
func TestRicketyGazeboReturnsUpToTwoMilledPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	old := game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(old)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Ring", TypeLine: "Artifact"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	forest, bolt, bear, ring, deep := ids[0], ids[1], ids[2], ids[3], ids[4]

	me.Hand.PushTop(effectstest.GreenhouseCard(me.ID))
	c := me.Hand.Cards[len(me.Hand.Cards)-1]
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatal(err)
	}
	effectstest.RoomsDSettle(t, g, me.ID)

	if !me.Library.Contains(deep) {
		t.Fatal("milled more than four cards")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt for the permanent cards")
	}
	if pick.ChooseMax != 2 || len(pick.ChooseCards) != 3 || effectstest.HasID(pick.ChooseCards, bolt) || effectstest.HasID(pick.ChooseCards, old.InstanceID) {
		t.Fatalf("prompt = max %d cards %v, want up to 2 of the milled permanents (land, creature, artifact)", pick.ChooseMax, pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, forest, ring)
	if !me.Hand.Contains(forest) || !me.Hand.Contains(ring) || me.Hand.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the two chosen permanents should be in hand and the third left in the graveyard")
	}
}
