package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_only_mana_test.go — the proof cards for #1242 and #1283.
//
//   - Gold, Eldrazi Spawn and Eldrazi Scion: the three catalog tokens
//     whose mana ability is a sacrifice with NO {T}. Until #1242 the
//     auto-tapper refused all three, so a board of them read as
//     unpayable to the preview, the strict gate and every bot.
//   - Deadly Dispute: the cast that names a Spawn to its additional
//     cost is not allowed to crack that same Spawn for its mana.
//   - Cadaverous Bloom: "Exile a card from your hand" — the auto-tapper
//     never plans it, the hand-clicked activation pays it, and the
//     card it spends is EXILED, not discarded.

const cadaverousBloomOracle = "fbb0f73b-5e30-4632-99c1-e49582e41f8d"

// seatToken puts a catalog token under `owner` on the battlefield.
func SeatToken(g *game.Game, owner uuid.UUID, tok game.Card) uuid.UUID {
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = owner, owner
	return PushBattlefieldCardWithTimestamp(g, tok)
}

// --- Cadaverous Bloom ------------------------------------------------

func SeatBloomWithHand(t *testing.T, n int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	bloom := SeatToken(g, me.ID, game.Card{Name: "Cadaverous Bloom", TypeLine: "Enchantment", OracleID: cadaverousBloomOracle})
	hand := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare Spell", TypeLine: "Sorcery", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
		hand = append(hand, id)
	}
	return g, me, bloom, hand
}
