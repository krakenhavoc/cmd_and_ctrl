package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rebound_spells_test.go — the rebound cards whose own text is more
// than one primitive (#1854, ADR 0107 §3). Rebound itself is pinned in
// game/rebound_test.go and rebound_cards_test.go.

const (
	BlessedReincarnationOracle = "ca29588c-f117-418e-be9e-fa2ee89862ca"
	BlossomingCalmOracle       = "489a60f1-83f8-465b-918f-7d63d4f76d14"
	ConsumingVaporsOracle      = "a10ce333-48d7-499b-9355-f381f2395497"
	FeveredSuspicionOracle     = "51545f6e-d0e3-4b4c-bc73-c731f85e26b0"
	FlameOnOracle              = "e715d33e-3d60-4623-b843-db0b906ae98b"
	NomadsAssemblyOracle       = "a9a0088a-3c86-4baf-9754-79006f733ce1"
	QuantumMisalignmentOracle  = "197582b9-4c86-41a3-ad0c-789a0db8e087"
	RecurringInsightOracle     = "c6815bbb-24a5-4c9f-bf30-6190bc766d05"
	SurrealMemoirOracle        = "fa482ded-b24c-4f9e-948e-d0620e907b8b"
	SurvivalCacheOracle        = "5fb8be5a-3666-4680-84e2-341cb269df07"
	TransposeOracle            = "77d5f298-06b4-49d1-9b33-f1f176665ba2"
)

// castReboundSpellAtMain seeds a catalog card in the active seat's
// hand at its main phase and casts it.
func CastReboundSpellAtMain(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceTo(t, g, game.StepPrecombatMain)
	id := HandCardForTest(me, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return me, id
}

// libraryTop pushes cards on top of a library, last argument on top.
func LibraryTop(p *game.Player, cards ...game.Card) {
	for _, c := range cards {
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
	}
}

func CardNamed(name, typeLine string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = typeLine
	return c
}
