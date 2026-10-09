package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Draconic Visitor — Creature — Dragon {3}{R}{R}, 5/5:
//
//	"Flying
//	 If one or more artifact tokens would be created under your
//	 control, that many 5/5 red Dragon creature tokens with flying are
//	 created instead."
//
// Academy Manufactor's shape (CR 614.1): a replacement on the
// token-creation event, scoped to tokens created under the Visitor's
// controller. It rewrites only the groups whose template is an
// artifact, so a Treasure made beside a Soldier becomes a Dragon and
// the Soldier stays a Soldier, and a creature the Dragon replaces keeps
// its count ("that many"). An artifact CREATURE token (a Construct)
// counts as an artifact token too, as printed. The Dragons are
// creature-only, so Treasure's mana ability and a Clue's draw are gone
// with the token they replaced.
//
// No simplification.
func init() {
	isArtifactToken := func(t game.Card) bool { return t.IsArtifact() }
	Register(Spec{
		OracleID:        "6745e850-68f6-480b-a1e0-e06180dde169",
		Name:            "Draconic Visitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens &&
					ev.TokenController == src.Controller &&
					ev.TokenTemplatesMatch(isArtifactToken)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.ReplaceTokenKindsWhere(isArtifactToken, TokenCard("5/5 red Dragon with flying"))
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Draconic Visitor: 5/5 red Dragons with flying instead",
		}},
	})
}
