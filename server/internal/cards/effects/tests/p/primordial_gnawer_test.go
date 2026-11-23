package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const primordialGnawerOracle = "80ca6596-9763-4310-b30b-7153ade9f567"

func killPrimordialGnawer(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	gnawer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Primordial Gnawer",
		TypeLine: "Creature — Insect Horror", ManaCost: "{4}{B}",
		OracleID: primordialGnawerOracle, Power: 5, Toughness: 2,
		Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(gnawer); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
}

// "When this creature dies, discover 3": the hit is the first nonland
// card with mana value 3 or less, and declining puts it into hand.
func TestPrimordialGnawerDiscoversThreeOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	three := effectstest.DiscoverLibraryCard(me.ID, "Three Drop", "Creature — Bear", "{2}{G}")
	four := effectstest.DiscoverLibraryCard(me.ID, "Four Drop", "Creature — Bear", "{3}{G}")
	// Bottom first: the four-drop is on top, so the walk passes over it.
	me.Library.Cards = []game.Card{three, four}

	killPrimordialGnawer(t, g, me.ID)
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if effectstest.IsKeywordCastPrompt(c) {
			prompt = c
		}
	}
	if prompt == nil || effectstest.KeywordCastCard(prompt) != three.InstanceID {
		t.Fatalf("discover 3 did not stop on the three-drop: %+v", prompt)
	}
	effectstest.AnswerMayCastPrompt(t, g, me.ID, false)
	if !me.Hand.Contains(three.InstanceID) {
		t.Errorf("the discovered card did not go into hand")
	}
	if !me.Library.Contains(four.InstanceID) {
		t.Errorf("the passed-over four-drop did not go back to the library")
	}
}

// Taking the free cast casts it out of exile for nothing.
func TestPrimordialGnawerDiscoveredCardIsCastableForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = []game.Card{effectstest.DiscoverLibraryCard(me.ID, "Opt", "Instant", "{U}")}
	hit := me.Library.Cards[0].InstanceID

	killPrimordialGnawer(t, g, me.ID)
	// Taking the offer is the cast, made while the trigger resolves.
	effectstest.AnswerMayCastPrompt(t, g, me.ID, true)
	if !g.Stack.Contains(hit) {
		t.Errorf("the discovered spell is not on the stack")
	}
}
