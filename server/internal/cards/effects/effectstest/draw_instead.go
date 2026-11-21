package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tests for #2127 (dredge) and #2168 (a draw replaced by an effect that
// asks which card you take). Seat 1 draws throughout: seat 0 is the
// active player sitting in its draw step, where DrawCard is a no-op.

const (
	StinkweedImpOracle   = "e005bc76-4985-4ec6-b9f6-cf6d0d9f5df4" // dredge 5
	GolgariThugOracle    = "a426a258-fd8b-489c-8642-9868ee47de85" // dredge 4
	ShamblingShellOracle = "9c50f6b1-8850-4b08-a994-3ccc17f25b83" // dredge 3
	GraveTrollOracle     = "686f4a37-b5e4-46d4-9e0e-11794b2d12cd" // dredge 6
	UnderrealmLichOracle = "e1bce9c3-300c-4a9d-abe0-a1f02d3a1105"
	ForbiddenCryptOracle = "8b34b211-985b-4934-bb3b-8e6672685ac2"
	NecrobloomOracle     = "b981af39-4ee6-4fbc-9a89-618dcad9dfbf"
)

func GyCardWithOracle(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func BfCardWithOracle(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID,
	})
}

func ZoneHas(z *game.Zone, id uuid.UUID) bool { return z.Contains(id) }

func DrawN(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, n) })
}

func OptionalOffer(g *game.Game, p *game.Player) *game.PendingChoice {
	return LatestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, p.ID)
}

func AnswerOptional(t *testing.T, g *game.Game, p *game.Player, apply bool) {
	t.Helper()
	offer := OptionalOffer(g, p)
	if offer == nil {
		t.Fatal("no optional_replacement prompt is open")
	}
	if err := g.ResolveOptionalReplacement(offer.ID, p.ID, apply); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}

// --- a doubler of draws -------------------------------------------------

// orderFirst answers a CR 616 prompt with the effect whose label starts
// with `first` ahead of the rest.
func OrderFirst(t *testing.T, g *game.Game, p *game.Player, first string) {
	t.Helper()
	order := LatestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, p.ID)
	if order == nil {
		t.Fatal("expected a CR 616 ordering prompt")
	}
	var ids []game.ReplacementEffectID
	for _, id := range order.ReplacementEffectIDs {
		label, _ := g.ReplacementOptionMetaForEffect(id)
		if len(label) >= len(first) && label[:len(first)] == first {
			ids = append([]game.ReplacementEffectID{id}, ids...)
		} else {
			ids = append(ids, id)
		}
	}
	if err := g.ResolveReplacementOrder(order.ID, p.ID, ids); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}
