package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tspB1ActivateGranted activates the first granted activated ability of
// `card` for `controller` and settles the stack.
func TspB1ActivateGranted(t *testing.T, g *game.Game, controller, card uuid.UUID, targets []game.TargetRef) {
	t.Helper()
	idx, ref := GrantedActivatedIndex(t, g, card)
	if idx < 0 {
		t.Fatal("the Sliver has no granted activated ability")
	}
	B16Activate(t, g, controller, card, idx, game.ActivateAbilityParams{Ref: ref, Targets: targets})
}

// tspB1Loot discards `card` from the active seat's otherwise-empty hand
// through Faithless Looting, so a madness card is exiled and offered.
func TspB1Loot(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	id := PushCatalogHandCard(me, name, typeLine, oracle)
	spare := PushCatalogHandCard(me, "Spare", "Sorcery", "")
	CastCatalogSpell(t, g, "Faithless Looting", "Sorcery", FaithlessLootingOracl, nil)
	PassPriorityAroundTable(t, g)
	AnswerChooseCards(t, g, me.ID, id, spare)
	PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Fatalf("%s was not discarded into exile (CR 702.35a)", name)
	}
	if LatestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatalf("%s: the madness trigger offered no cast", name)
	}
	return id
}
