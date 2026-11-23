package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// torB1LootIntoExile puts the card in an emptied hand beside a spare,
// casts Faithless Looting and discards both, and checks the madness
// card went to exile with a cast offered.
func TorB1LootIntoExile(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := PushCatalogHandCard(me, name, typeLine, oracle)
	spare := PushCatalogHandCard(me, "Spare", "Sorcery", "")
	CastCatalogSpell(t, g, "Faithless Looting", "Sorcery", FaithlessLootingOracl, nil)
	PassPriorityAroundTable(t, g)
	AnswerChooseCards(t, g, me.ID, card, spare)
	PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(card) || me.Graveyard.Contains(card) {
		t.Fatalf("the discarded %s is not in exile (CR 702.35a)", name)
	}
	if LatestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatalf("the madness trigger offered no cast of %s", name)
	}
	return card
}
