package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// e7_legendaries_test.go — issue #1117, the five hardest cards in the
// Betor life-swap deck. One focused test per card, each asserting the
// clause that could plausibly have been got wrong rather than merely
// that the card registers.

const (
	DragonlordDromokaOracle = "82cdd612-3322-4372-810f-1ff106ea8e6a"
	SelvalaHeartOracle      = "1d725121-e50c-42f0-9128-56802f07c89e"
	TroubleInPairsOracle    = "f349f58b-8cc8-45e4-9565-2b46fdf976c9"
	GixPraetorOracle        = "928d977e-cff0-4e0e-83bb-16d73a754f35"
	ValgavothOracle         = "cae3ec72-436d-4086-9dcb-17b3d92ad5c4"
)

// e7CastInstant announces a plain non-catalog instant from a seat's
// hand and hands back the announce error, so a test can assert either
// outcome. No targets and no catalog entry: the only thing that can
// refuse it is the CR 101.2 cast gate.
func E7CastInstant(g *game.Game, caster *game.Player) error {
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Test Instant", TypeLine: "Instant",
		Owner: caster.ID, Controller: caster.ID,
	})
	return g.CastSpell(caster.ID, id, game.CastSpellParams{})
}

// e7Main walks the cursor to a main phase.
func E7Main(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// confirmForTriggerPrompt returns the newest open "you may" trigger
// prompt owed by a seat, or nil.
func ConfirmForTriggerPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			out = c
		}
	}
	return out
}
