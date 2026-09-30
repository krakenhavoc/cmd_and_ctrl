package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const primordialGnawerOracle = "80ca6596-9763-4310-b30b-7153ade9f567"

func killPrimordialGnawer(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	gnawer := pushBattlefieldCardWithTimestamp(g, game.Card{
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
	passPriorityAroundTable(t, g)
}

// "When this creature dies, discover 3": the hit is the first nonland
// card with mana value 3 or less, and declining puts it into hand.
func TestPrimordialGnawerDiscoversThreeOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	three := discoverLibraryCard(me.ID, "Three Drop", "Creature — Bear", "{2}{G}")
	four := discoverLibraryCard(me.ID, "Four Drop", "Creature — Bear", "{3}{G}")
	// Bottom first: the four-drop is on top, so the walk passes over it.
	me.Library.Cards = []game.Card{three, four}

	killPrimordialGnawer(t, g, me.ID)
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			prompt = c
		}
	}
	if prompt == nil || prompt.MayCastCard != three.InstanceID {
		t.Fatalf("discover 3 did not stop on the three-drop: %+v", prompt)
	}
	answerMayCastPrompt(t, g, me.ID, false)
	if !me.Hand.Contains(three.InstanceID) {
		t.Errorf("the discovered card did not go into hand")
	}
	if !me.Library.Contains(four.InstanceID) {
		t.Errorf("the passed-over four-drop did not go back to the library")
	}
}

// Taking the free cast casts it out of exile for nothing.
func TestPrimordialGnawerDiscoveredCardIsCastableForFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = []game.Card{discoverLibraryCard(me.ID, "Opt", "Instant", "{U}")}
	hit := me.Library.Cards[0].InstanceID

	killPrimordialGnawer(t, g, me.ID)
	answerMayCastPrompt(t, g, me.ID, true)
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("free cast of the discovered card: %v", err)
	}
	if !g.Stack.Contains(hit) {
		t.Errorf("the discovered spell is not on the stack")
	}
}
