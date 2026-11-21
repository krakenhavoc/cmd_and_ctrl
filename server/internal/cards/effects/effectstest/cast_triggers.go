package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_triggers_test.go — S19 sub-PR 6: cast / opponent-draws
// triggers and the pay-unless prompt. Seat 0 is the active player
// in newCatalogGame; the "opponent" cards live under seat 1 so a
// seat-0 cast / draw is an opponent's from their point of view.

const (
	RhysticStudyOracle      = "53236dd7-845a-444c-96d5-f41ed7325d8f"
	SmotheringTitheOracle   = "153376c9-dffd-458c-8ce3-a4c8269bc4e9"
	EsperSentinelOracle     = "5def9f38-0a0b-4e8d-9f9d-29dcb46520b4"
	BeastWhispererOracle    = "5da7eea8-bb9e-47ce-a554-8a1ee058bd7a"
	ConsecratedSphinxOracle = "311a449d-dc74-46e6-9a47-6a597931f736"
	LightningBoltOracle     = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// answerPayUnless finds the pending pay_unless prompt addressed to
// chooserID and answers it (true = pay).
func AnswerPayUnless(t *testing.T, g *game.Game, chooserID uuid.UUID, pay bool) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || c.Kind != game.PendingChoicePayUnless || c.Chooser != chooserID {
			continue
		}
		snapshot := *c
		if err := g.ResolvePayUnless(c.ID, chooserID, pay); err != nil {
			t.Fatalf("ResolvePayUnless: %v", err)
		}
		return &snapshot
	}
	t.Fatalf("no pay_unless prompt addressed to %s", chooserID)
	return nil
}

func HasPayUnlessFor(g *game.Game, chooserID uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooserID {
			return true
		}
	}
	return false
}

func CountOnBattlefield(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}
