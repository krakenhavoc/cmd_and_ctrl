package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Esix, Fractal Bloom — Legendary Creature — Fractal {4}{G}{U}, 4/4
// (Murders at Karlov Manor Commander, #5457):
//
//	"Flying
//	 The first time you would create one or more tokens during each of
//	 your turns, you may instead choose a creature other than Esix and
//	 create that many tokens that are copies of that creature."
//
// Moonlit Meditation's replacement with a choice of object inside it: an
// "instead" on the creation instruction (CR 701.7b, CR 614.1a) that keeps
// each group's count and entry clause and swaps the kind
// (ReplaceTokenKinds) for a copy of the creature chosen. The choice is
// the copy_target prompt asked of a token creation
// (game.CopySelector on RepEventCreateTokens, game.CopyUseTokens): any
// creature on the battlefield but Esix, whoever controls it, and it is
// not targeted, so hexproof, shroud and protection do not narrow it
// (CR 115.10a). Choosing none is the "may" declined, and the tokens are
// created as they were. The copies take the creature's copiable values
// as the instruction is replaced (TokenCopyTemplate, CR 707.2): not its
// counters, tapped state or attachments.
//
// "The first time … during each of your turns" is the first creation
// instruction under your control in a turn of yours, whether or not Esix
// was there for it (the turn tally's TokensCreated, plus an instruction
// still waiting on an answer). Declining uses it up, and another
// player's turn never offers it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9d22960b-babc-4cf3-b228-d32e13bc6014",
		Name:            "Esix, Fractal Bloom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			RowLabel: "The first time you would create tokens during each of your turns, you may create copies of a creature instead.",
			Watches:  []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCreateTokens || ev.TokenCount() <= 0 || ev.TokenController != src.Controller {
					return false
				}
				return isActivePlayer(g, src.Controller) &&
					g.TurnTallyFor(src.Controller).TokensCreated == 0 &&
					!g.TokenCreationAwaitingAnswerForEffect(src.Controller, ev.ID)
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
				tmpl, ok := TokenCopyTemplate(g, ev.CopySourceForEffect())
				if !ok {
					return nil
				}
				ev.ReplaceTokenKinds(tmpl)
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:          "Esix, Fractal Bloom: create copies of a creature instead",
			PromptQuestion: "Esix, Fractal Bloom — create that many copies of a creature instead?",
			CopySelector: &game.CopySelector{
				Candidates: func(_ *game.ReplacementEvent, g *game.Game, src *game.Card) []uuid.UUID {
					return copyCandidates(g, src.InstanceID, func(c game.Card) bool { return c.IsCreature() })
				},
			},
		}},
	})
}
