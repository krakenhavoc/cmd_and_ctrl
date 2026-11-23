package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// proliferate_test.go — S21's Proliferate primitive (CR 701.34) and
// the five catalog cards that were gated on it.

const (
	SteadyProgressOracle = "d2145ad3-fe7e-459b-b155-1b3ed089e936"
	KarnsBastionOracle   = "9fb8cd81-403a-4988-8f1c-b8eccf8abd9c"
	ContagionClaspOracle = "43f2d81e-aa01-4fa9-9046-6a27a05dbd2d"
	FluxChannelerOracle  = "83874e60-291b-47a7-ba9f-69437fa7e3c7"
	EvolutionSageOracle  = "b45fdeab-00cc-4422-af9f-66f30a880a7c"
	InexorableTideOracle = "40ec0a47-badf-4074-b0a8-749bb7c17b95"
)

// pushCounterCreature seeds a 4/4 so a -1/-1 counter (or two) does
// not immediately kill it — pushCatalogPermanent's 1/1 dies to the
// CR 704.5f SBA the moment the first counter lands, which would make
// every assertion below read "card not found".
func PushCounterCreature(g *game.Game, owner uuid.UUID, name string, counter string, n int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: owner, Controller: owner,
	})
	if n > 0 {
		_ = g.AddCounter(id, counter, n)
	}
	return id
}

func CounterCount(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	return -1
}

// --- the player's choice (#2525) -------------------------------------

// proliferatePrompt returns the open proliferate prompt, or nil.
func ProliferatePrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceProliferate {
			return c
		}
	}
	return nil
}

// answerProliferate answers the open proliferate prompt with `picks`
// (permanents and seats share the one list).
func AnswerProliferate(t *testing.T, g *game.Game, picks ...uuid.UUID) {
	t.Helper()
	c := ProliferatePrompt(g)
	if c == nil {
		t.Fatal("no proliferate prompt is open")
	}
	if err := g.ResolveProliferate(c.ID, c.Chooser, picks); err != nil {
		t.Fatalf("ResolveProliferate: %v", err)
	}
}

// settleAnsweringProliferate is passPriorityAroundTable for a test that
// does not care WHAT was proliferated: whenever a proliferate prompt
// opens it takes the engine's suggested answer, which is what the old
// auto-pick did, and the stack is settled behind it.
func SettleAnsweringProliferate(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		PassPriorityAroundTable(t, g)
		c := ProliferatePrompt(g)
		if c == nil {
			return
		}
		AnswerProliferate(t, g, c.ChooseSuggested...)
	}
	t.Fatal("proliferate prompts never stopped")
}
