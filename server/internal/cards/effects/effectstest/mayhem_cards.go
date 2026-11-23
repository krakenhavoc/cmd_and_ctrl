package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mayhemHandCard seeds a catalog card into a player's hand.
func MayhemHandCard(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Owner: p.ID, Controller: p.ID,
			KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// mayhemDiscard makes `p` discard the card `id` through the engine's
// ordinary discard prompt, so the discard event is the real one.
func MayhemDiscard(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p.ID, N: 1})
	})
	AnswerDiscard(t, g, p.ID, id)
	if !p.Graveyard.Contains(id) {
		t.Fatalf("the discarded card is not in the graveyard")
	}
}

func CastSlashFromGraveyard(g *game.Game, me *game.Player, slash, victim uuid.UUID, alt string) error {
	return g.CastSpell(me.ID, slash, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: alt,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
}
