package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB92ChapterLabel is the stack label of a Saga chapter. saga.go's
// SagaChapterLabel prints no numeral past V, and these Sagas run to VI.
func gapB92ChapterLabel(name string, chapter int, text string) string {
	numerals := []string{"", "I", "II", "III", "IV", "V", "VI"}
	if chapter < 1 || chapter >= len(numerals) {
		return name + " — " + text
	}
	return name + " — " + numerals[chapter] + ": " + text
}

// gapB92ShuffleThenInvestigate is "Tail.Player shuffles their library, then
// investigates" (Blink's "its owner shuffles it into their library, then
// investigates"): the owner is the player of the tail. An on-disk key.
var gapB92ShuffleThenInvestigate = game.RegisterTail("then/gap-b92-shuffle-then-investigate", func(g *game.Game, o game.TailOutcome) error {
	if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	return g.InvestigateForEffect(o.Tail.Player, o.Tail.Source, 1)
})

// gapB92TokenYouControlNotASaga is "target non-Saga token you control".
func gapB92TokenYouControlNotASaga(_ *game.Game, controller uuid.UUID, c game.Card) bool {
	return c.IsToken() && c.Controller == controller && !game.IsSaga(c)
}

// gapB92CopyTheChosenToken is "create a token that's a copy of target <token>"
// as a whole ability body (City of Death's chapters, Octomancer's end step):
// the first target still legal as it resolves is copied (CR 608.2b).
func gapB92CopyTheChosenToken(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return CreateTokenCopy{Controller: item.Controller, Copy: id, N: 1}.Apply(ctx)
}

// gapB92FourDamageToTheChosenCreature is "<source> deals 4 damage to target
// creature" as a whole ability body (The Flux's chapter I, Super-Skrull's
// activated ability).
func gapB92FourDamageToTheChosenCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: id, Amount: 4}.Apply(ctx)
}
