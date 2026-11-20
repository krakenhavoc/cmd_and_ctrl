package effectstest

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s46Tokens lists the tokens `controller` has named `name`.
func S46Tokens(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller && effects.IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}

// s46Settle passes priority around the table and answers trigger-order and
// target prompts until neither is pending.
func S46Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		PassPriorityAroundTable(t, g)
		for TriggerOrderPrompt(g) != nil {
			AnswerTriggerOrderInOfferedOrder(t, g)
		}
		AnswerAnyPendingTargetPrompts(t, g)
	}
}

// s46Encore activates encore from the graveyard and resolves it.
func S46Encore(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tgh int, mana string) (uuid.UUID, *game.Player) {
	t.Helper()
	id, me := ActivateFromGraveyard(t, g, name, typeLine, oracle, p, tgh, mana, game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatalf("%s: encore's cost exiles the card", name)
	}
	PassPriorityAroundTable(t, g)
	return id, me
}
