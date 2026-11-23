package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// voting_b_test.go — #2928, the second slice of the other voting cards
// on ADR 0146's seam (amendment 2026-10-13): Selvala's Stampede, Custodi
// Squire, Trial of a Time Lord, Illusion of Choice and The Valeyard.

const (
	SelvalasStampedeOracle = "9d2aea05-f55c-4355-950f-bfdc4e985e80"
	CustodiSquireOracle    = "07c7193e-f591-4aee-a3cc-3a2c52a38f9b"
	TrialOfATimeLordOracle = "a5425d4f-cd6c-4240-89a4-dd13e0e1a95f"
	IllusionOfChoiceOracle = "e3ad9287-ef13-4b66-9df8-b17e15bb4b0f"
	TheValeyardOracle      = "ed202f1d-3511-45f4-bfbc-e6b76d4311d1"
)

// setLibraryTopFirst replaces p's library with the cards, the first
// being the top of it.
func SetLibraryTopFirst(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func LibraryCreature(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Test", Power: 1, Toughness: 1}
}

func LibraryLand(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — Forest"}
}

func ZoneKindOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	if z := g.FindCardZoneForEffect(id); z != nil {
		return z.Kind
	}
	return ""
}

// ---- the bot's move list --------------------------------------------------

// dispatchEveryMove proves the engine accepts each enumerated answer.
func DispatchEveryMove(t *testing.T, g *game.Game, seat uuid.UUID, moves []legal.Move) {
	t.Helper()
	for _, m := range moves {
		cl := g.Clone()
		if err := actions.Dispatch(cl, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: seat, Params: m.Params}); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
		}
	}
}

func ResolveChoiceMoves(moves []legal.Move) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Type == legal.TypeResolveChoice {
			out = append(out, m)
		}
	}
	return out
}
