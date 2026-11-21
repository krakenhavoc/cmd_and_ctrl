package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tutors_test.go — the "shuffle, then put that card on top" cycle.
//
// The assertions that earn their keep are about WHERE and WHEN, not
// about whether a search happened:
//
//   - the card does not go to hand. These cost one mana because
//     collecting the card costs a draw step; a to-hand search is a
//     strictly better card than the one printed.
//   - the placement happens AFTER the shuffle. A shuffle that ran
//     last would put the tutored card back into the deck at random,
//     which is the bug this ordering exists to avoid — and it is
//     invisible unless the test seeds a library big enough for a
//     shuffle to actually move things.
//   - "reveal it" reaches the whole table. Enlightened Tutor tells
//     every opponent what your next draw is, and that is a real cost
//     of the card. Imperial Seal pays 2 life to skip it.

const (
	EnlightenedTutorOracle = "c5229c17-b7be-4b05-b683-f2277edc4849"
	WorldlyTutorOracle     = "e8863518-0bfa-49c3-8c6e-6c9116a81051"
	MysticalTutorOracle    = "fb81f95c-70f8-4eb7-8d15-15d0ae23ec03"
	ImperialSealOracle     = "16cd0b90-f70c-4efa-b252-8de8784ef9a3"
	BeseechOracle          = "cf94cafc-527e-4b27-8a28-7807435aaccf"
)

// tutorTopOfLibrary casts a put-on-top tutor, answers the search with
// `needle`, and returns the card now sitting on top.
//
// The answer is conditional because the engine skips the prompt when
// the predicate leaves no decision to make — a library with exactly
// one matching card auto-takes it. Both paths have to reach the same
// place, which is itself worth exercising: the put-on-top runs in
// finishSearchLocked, so it must fire whether the pick came from a
// prompt or from the no-decision shortcut.
func TutorTopOfLibrary(t *testing.T, g *game.Game, p *game.Player,
	name, typeLine, oracle string, needle uuid.UUID,
) game.Card {
	t.Helper()
	CastCatalogSpell(t, g, name, typeLine, oracle, nil)
	PassPriorityAroundTable(t, g)
	if SearchChoiceFor(g, p.ID) != nil {
		AnswerSearchByID(t, g, p.ID, needle)
	}
	top, err := p.Library.Top()
	if err != nil {
		t.Fatalf("%s: library is empty after the tutor", name)
	}
	return top
}
