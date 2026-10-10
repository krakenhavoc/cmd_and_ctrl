package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Monumental Henge — Land:
//
//	"This land enters tapped unless you control a Plains.
//	 {T}: Add {W}.
//	 {2}{W}{W}, {T}: Look at the top five cards of your library. You may
//	 reveal a historic card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order.
//	 (Artifacts, legendaries, and Sagas are historic.)"
//
// The entry clause is `EntersTappedUnless` (a replacement, so no tap
// event), reading the lands-and-Plains the controller already has: the
// pipeline runs pre-push, so the entering Henge is not on the board and
// cannot count itself (it has no land types anyway). Plains is read
// post-layer. The dig is `TakeFromLibraryToHand` over the five looked-at
// cards, optional and revealed, with the rest to the bottom in the
// game's seeded random order; historic is the shared `b09IsHistoric`
// test (artifact, Saga or legendary).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c48df45c-3513-4d56-aed6-30c2f3a759cd",
		Name:         "Monumental Henge",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(controlsAPlains)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}{W}, {T}: Look at the top five cards of your library. You may reveal a historic card from among them and put it into your hand. Put the rest on the bottom of your library in a random order",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{W}{W}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Controller
				return TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 5),
					Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return b09IsHistoric(c) },
					Max:      1,
					Optional: true,
					Reveal:   true,
					Label:    "Monumental Henge — you may reveal a historic card and put it into your hand",
					Then:     TakeRestOnBottomInRandomOrder,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// controlsAPlains is Monumental Henge's "unless you control a Plains".
func controlsAPlains(g *game.Game, src *game.Card) bool {
	plains := MatchLandSubtype("Plains")
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != src.InstanceID && c.Controller == src.Controller && plains(c) {
			return true
		}
	}
	return false
}
