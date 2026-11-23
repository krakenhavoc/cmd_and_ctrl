package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dreadTable is a catalog game at the active seat's precombat main with
// a creature card on top of its library over a land card, returning
// (game, seat, top, second). The cards sit under the seat's draw for the
// turn, which has already happened.
func DreadTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceToMain(t, g)
	top, second := DreadPair(me)
	return g, me, top, second
}

// dreadPair stacks "Top Bear" (a creature card, flippable for {1}{G})
// over "Second Land" and returns their IDs, top first.
func DreadPair(p *game.Player) (top, second uuid.UUID) {
	second = uuid.New()
	top = uuid.New()
	p.Library.PushTop(game.Card{InstanceID: second, Name: "Second Land", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
	p.Library.PushTop(game.Card{InstanceID: top, Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return top, second
}

// requireManifested fails unless id is a face-down manifested permanent
// that only its controller may look at.
func RequireManifested(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) *game.Card {
	t.Helper()
	c := FindBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	if !c.FaceDown || c.FaceDownKind != game.FaceDownManifested {
		t.Fatalf("card %s face down = %v kind %q, want a manifested permanent", id, c.FaceDown, c.FaceDownKind)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && c.IsKnownTo(p.ID) {
			t.Errorf("seat %s knows the manifested card", p.Name)
		}
	}
	return c
}

func MdInGraveyard(p *game.Player, id uuid.UUID) bool { return p.Graveyard.Contains(id) }

// answerDread answers the open manifest-dread prompt with `pick`.
func AnswerDread(t *testing.T, g *game.Game, me *game.Player, pick uuid.UUID) {
	t.Helper()
	PassPriorityAroundTable(t, g)
	AnswerChooseCards(t, g, me.ID, pick)
}
