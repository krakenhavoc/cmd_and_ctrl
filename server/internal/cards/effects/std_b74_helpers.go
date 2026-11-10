package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copiesOfAttachedInsteadOfFirstTokens is "the first time you would
// create one or more tokens each turn, you may instead create that many
// tokens that are copies of <the permanent this is attached to>": Moonlit
// Meditation's enchanted permanent and Mirrormind Crown's equipped
// creature.
//
// A "may" replacement (CR 614.1a, game.ReplacementEffect.Optional) on the
// CR 701.7b creation instruction. It keeps each group's count and entry
// clause and swaps the kind (ReplaceTokenKinds) for TokenCopyTemplate of
// the host as the instruction is replaced: its copiable values, copy
// effects included (CR 707.2). "The first time … each turn" is the turn
// tally's TokensCreated plus an instruction still waiting on its answer
// (TokenCreationAwaitingAnswerForEffect), so it is the first creation
// under your control this turn whether or not the host was there for it.
//
// hostOK is the condition on the host itself: Mirrormind Crown's "as
// long as this Equipment is attached to a creature". Nil accepts any
// attached permanent.
func copiesOfAttachedInsteadOfFirstTokens(label, question string, hostOK func(game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:  []game.EventKind{game.EventTokenCreated},
		Optional: true,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventCreateTokens || ev.TokenCount() <= 0 || ev.TokenController != src.Controller {
				return false
			}
			if src.AttachedTo.ID == uuid.Nil || !onBattlefield(g, src.AttachedTo.ID) {
				return false
			}
			if hostOK != nil {
				host, ok := g.LookupCardForEffect(src.AttachedTo.ID)
				if !ok || !hostOK(host) {
					return false
				}
			}
			return g.TurnTallyFor(src.Controller).TokensCreated == 0 &&
				!g.TokenCreationAwaitingAnswerForEffect(src.Controller, ev.ID)
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			tmpl, ok := TokenCopyTemplate(g, src.AttachedTo.ID)
			if !ok {
				return nil
			}
			ev.ReplaceTokenKinds(tmpl)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label:          label,
		PromptQuestion: question,
	}
}
