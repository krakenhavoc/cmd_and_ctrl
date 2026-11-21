package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shocklands_test.go — "as this land enters, you may pay 2 life. If
// you don't, it enters tapped."
//
// The assertion that earns its keep in every one of these is the tap
// EVENT count, not the Tapped flag. The implementation this replaced
// entered tapped and untapped itself for 2 life, which produces the
// same final Tapped state in both branches and is therefore invisible
// to a state assertion. Events are not fooled:
//
//	replacement      → 0 tap events, 0 untap events, ever
//	ETB-tap patch-up → 1 tap event (or 1 untap on the pay branch)
//
// The prompt itself is the second half of the proof: nothing is on
// the battlefield while it is open, because the entry has not
// happened yet.

const (
	BloodCryptOracle       = "43985bbc-a0f6-4812-984e-392bc8562633"
	HallowedFountainOracle = "f1750962-a87c-49f6-b731-02ae971ac6ea"
	SteamVentsOracle       = "17039058-822d-409f-938c-b727a366ba63"
)

// shocklandOracleIDs is the whole cycle, name → oracle ID, verified
// against Scryfall's default-cards dump.
var ShocklandOracleIDs = map[string]string{
	"Blood Crypt":       BloodCryptOracle,
	"Breeding Pool":     "20283c4a-f1f0-42f0-bc08-6da87474426b",
	"Godless Shrine":    "73864fcc-1bde-4bc0-831e-2b93e546e417",
	"Hallowed Fountain": HallowedFountainOracle,
	"Overgrown Tomb":    "975ec9a3-6f20-4177-8211-82526e092538",
	"Sacred Foundry":    "45181cb8-2090-4471-ba90-e5a8f04d525f",
	"Steam Vents":       SteamVentsOracle,
	"Stomping Ground":   "16052b52-ade1-406f-a06b-ce7ea607fb63",
	"Temple Garden":     "f413a83d-a40d-434c-b20a-4c707c0527fa",
	"Watery Grave":      "fc9ec820-4245-4a96-b009-5308a818ca58",
}

// entryPayLifeChoiceFor returns the open "you may pay N life" prompt
// addressed to chooser, or nil.
func EntryPayLifeChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceEntryPayLife && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// answerEntryPayLife answers the open entry prompt for chooser.
func AnswerEntryPayLife(t *testing.T, g *game.Game, chooser uuid.UUID, pay bool) {
	t.Helper()
	c := EntryPayLifeChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no PendingChoiceEntryPayLife addressed to %s", chooser)
	}
	if err := g.ResolveEntryPayLife(c.ID, chooser, pay); err != nil {
		t.Fatalf("ResolveEntryPayLife: %v", err)
	}
}

// untapEventsFor counts EventUntapCard entries for one card — the
// tell of the enter-tapped-then-untap implementation this replaced.
func UntapEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventUntapCard && ev.CardID == id {
			n++
		}
	}
	return n
}
