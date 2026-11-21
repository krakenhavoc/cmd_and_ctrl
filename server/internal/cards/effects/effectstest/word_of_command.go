package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0166, #2164: Word of Command and the control of a player it
// gives (CR 723.2).

const WordOfCommandOracle = "e8ad3a77-b293-4d69-b080-27ca9f95d443"

// wocSetup casts Word of Command from seat 0 at seat 1 and answers
// nothing. The victim's hand is `hand`; it controls one untapped basic
// land of each type named in `lands`.
func WocSetup(t *testing.T, lands []string, hand ...game.Card) (g *game.Game, caster, victim *game.Player, ids []uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	caster, victim = g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { ids = RevealHand(victim, hand...) })
	for _, land := range lands {
		PushCatalogPermanent(g, victim.ID, land, "Basic Land — "+land, "", false)
	}
	CastAt(t, g, "Word of Command", "Instant", WordOfCommandOracle, PbPlayer(victim.ID))
	return g, caster, victim, ids
}

func WocBolt() game.Card {
	c := RhBolt()
	c.OracleID = LightningBoltOracle
	return c
}

// openControlledPlay returns the one controlled_play prompt, or nil.
func OpenControlledPlay(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceControlledPlay {
			return c
		}
	}
	return nil
}

// passUntilChoice passes priority until a prompt of that kind is open.
func PassUntilChoice(t *testing.T, g *game.Game, kind game.PendingChoiceKind) {
	t.Helper()
	for range 16 {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == kind {
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("no %s prompt opened", kind)
}
