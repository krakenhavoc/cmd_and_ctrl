package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a9_helpers.go — shared bodies for The Hobbit, half A, batch 9 (#4454).

// hobA9InsteadCreateTokens is the "if you would create a <kind> token,
// instead create ..." replacement of Academy Manufactor and Bilbo, Fellow
// Conspirator (CR 701.7b, 614.1a): every group whose template satisfies
// `pred` is rewritten into one group per template, keeping its count, and
// the groups of other kinds in the same instruction are left alone. Only
// the source's controller's creations are watched. The templates are
// built when the replacement applies, not when the card registers.
func hobA9InsteadCreateTokens(label string, pred func(game.Card) bool, templates func() []game.Card) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventTokenCreated},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventCreateTokens &&
				ev.TokenController == src.Controller &&
				ev.TokenTemplatesMatch(pred)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.ReplaceTokenKindsWhere(pred, templates()...)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// hobA9OtherSubtypeYouControl counts the permanents `source`'s controller
// controls with `subtype`, not counting the source itself: the "another
// Dwarf" / "two or more other Wolves" of a self-referential static. The
// permanents are read after the layers, so a changeling counts.
func hobA9OtherSubtypeYouControl(g *game.Game, source *game.Card, subtype string) int {
	n := 0
	for _, id := range permanentsControlledByMatching(g, source.Controller, Subtype(subtype)) {
		if id != source.InstanceID {
			n++
		}
	}
	return n
}

// hobA9WhileYouControlAnother is the AppliesTo of "<this creature> has
// <keyword> as long as you control another <subtype>": the source alone,
// while its controller controls some other permanent with the subtype
// (Bolg's Company's haste, Dain's Company's lifelink).
func hobA9WhileYouControlAnother(subtype string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && hobA9OtherSubtypeYouControl(g, source, subtype) > 0
	}
}
