package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonlit Meditation — Enchantment — Aura {2}{U}:
//
//	"Enchant artifact or creature you control
//	 The first time you would create one or more tokens each turn, you
//	 may instead create that many tokens that are copies of enchanted
//	 permanent."
//
// A "may" replacement (CR 614.1a, game.ReplacementEffect.Optional) on
// the CR 701.7b creation instruction, the event Academy Manufactor
// rewrites. "Instead create that many tokens that are copies" keeps
// each group's count and its own entry clause — tapped, attacking — and
// swaps the kind (ReplaceTokenKinds), so any token, not only an artifact
// or creature token, can become a copy (ruling 2025-07-25). The copy is
// TokenCopyTemplate of the enchanted permanent as the instruction is
// replaced: its copiable values, copy effects included (CR 707.2).
//
// "The first time … each turn" is the first creation instruction under
// your control this turn, whether or not this Aura was there for it
// (ruling: Moonlit Meditation that arrives after you created tokens this
// turn waits for the next turn). The turn tally's TokensCreated says
// whether one has already made tokens, and an instruction still waiting
// on an answer is the first time too (TokenCreationAwaitingAnswerForEffect),
// so a second instruction in the same resolution is never offered the
// swap. Declining uses the turn's first time up. Other replacements on
// the same instruction — a doubler, Academy Manufactor — are ordered by
// you under CR 616.1, and the copies are then created and enter through
// the ordinary entry replacements (the ruling's "applies before anything
// that modifies how those tokens enter").
//
// A legendary host makes legendary copies, and the legend rule applies
// (CR 704.5j).
//
// No simplification.
const moonlitMeditationLabel = "Moonlit Meditation: create copies of the enchanted permanent instead"

func init() {
	Register(Spec{
		OracleID:     "df702b0c-e011-497f-ae29-9876efac4a4c",
		Name:         "Moonlit Meditation",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
		Replacements: []game.ReplacementEffect{{
			Watches:  []game.EventKind{game.EventTokenCreated},
			Optional: true,
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCreateTokens || ev.TokenCount() <= 0 || ev.TokenController != src.Controller {
					return false
				}
				if src.AttachedTo.ID == uuid.Nil || !onBattlefield(g, src.AttachedTo.ID) {
					return false
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
			Label:          moonlitMeditationLabel,
			PromptQuestion: "Moonlit Meditation — create copies of the enchanted permanent instead?",
		}},
	})
}
