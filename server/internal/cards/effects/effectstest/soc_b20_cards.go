package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b20_cards_test.go — Secrets of Strixhaven, twentieth batch (#4456).
// Each test plays the real catalog entry through the engine and asserts
// what the printed text says.

const (
	SocB20OmensOracle      = "863cbd12-da20-456e-8698-06fba4d9bd5c"
	SocB20AgonyOracle      = "55d1a788-fe9a-4e19-a524-6709f728ea28"
	SocB20ArkOracle        = "2c654772-c786-4966-a46e-9556bbf31d7c"
	SocB20ArnynOracle      = "3debd5c7-da3c-4130-8ed9-e156e10034b6"
	SocB20ProcessOracle    = "0c061a1b-206a-4fcc-92e5-30c8ed9d65c1"
	SocB20DustspeakerOrcl  = "295f8b8d-102a-47af-93c2-f8182c5f11ca"
	SocB20AzizaOracle      = "70e10cb9-ae27-4dba-9a0d-34a78f924774"
	SocB20BetrayalOracle   = "6b633e58-ea1b-4ed0-addb-6dfd93645ae4"
	SocB20TomekeeperOracle = "08a0e845-9ef9-4de1-b9fe-4530357346ca"
)

// --- fixtures ----------------------------------------------------------------

// socB20HandCard puts a card with a mana cost into a hand.
func SocB20HandCard(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// socB20CastStrict casts the card from hand with a floated pool and returns
// the cast error.
func SocB20CastStrict(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, pool string, targets []game.TargetRef) error {
	t.Helper()
	AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	FloatForTest(g, me, pool)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

// socB20ChooseCardsPrompt is the open choose-cards prompt for chooser.
func SocB20ChooseCardsPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
