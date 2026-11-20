package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hollowhead Sliver — Creature — Sliver {2}{R}, 2/2 (Modern Horizons, #5519):
//
//	"Sliver creatures you control have "{T}, Discard a card: Draw a card.""
//
// A granted activated ability (ADR 0093), Firewake Sliver's shape, on your
// own Slivers only. The {T} is the symbol, so a Sliver that arrived this
// turn can't use it.
//
// No simplification.
const hollowheadSliverGrant = "hollowhead-sliver/rummage"

func init() {
	Register(Spec{
		OracleID:     "739ad367-a4ff-4779-8888-3c27f0f60253",
		Name:         "Hollowhead Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: hollowheadSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}, Discard a card: Draw a card.",
				Cost:    Plus(TapCost(), DiscardACard()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  Do(DrawCards{N: 1}),
			}},
			Text: "{T}, Discard a card: Draw a card.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, hollowheadSliverGrant),
		},
	})
}
