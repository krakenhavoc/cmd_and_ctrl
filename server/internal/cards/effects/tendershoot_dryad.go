package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tendershoot Dryad — Creature — Dryad {4}{G}, 2/2 (EDHREC rank
// 1301):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 At the beginning of each upkeep, create a 1/1 green Saproling
//	 creature token.
//	 Saprolings you control get +2/+2 as long as you have the city's
//	 blessing."
//
// A Saproling every upkeep — every player's, as printed — and a lord
// once the board is wide. The token is an ordinary upkeep trigger
// with no actor test; the anthem is a Layer 7c static over Saprolings
// the controller controls (effective subtypes, so a changeling
// counts).
//
// The city's blessing is the player designation ascend gives and keeps
// (CR 702.131, #2696): the anthem reads it, so a board that shrinks
// below ten permanents keeps the bonus, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a336c10a-b5bd-47ff-ba2d-31e27af1e15a",
		Name:            "Tendershoot Dryad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Tendershoot Dryad — create a Saproling", Do(CreateToken{Template: b11GreenSaprolingToken(), N: 1})),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: WhileCitysBlessing(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller &&
					target.IsCreature() && target.HasSubtype("Saproling")
			}),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}
