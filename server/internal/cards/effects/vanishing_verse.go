package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanishing Verse — Instant {W}{B} (EDHREC rank 4284):
//
//	"Exile target monocolored permanent."
//
// Two mana, instant speed, exile, any permanent type — the only thing
// standing between this and a blank cheque is the word
// "monocolored", and in Commander that word takes out most commanders
// but not the gold ones. It is the Orzhov answer that does not care
// about indestructible, regeneration or a death trigger.
//
// "Monocolored" is EXACTLY one colour, read off the post-layer
// EffectiveColors, which has two consequences the card is played for
// and against:
//
//   - A COLOURLESS permanent is not monocolored (CR 105.2a: colorless
//     is not a colour), so an Eldrazi, a Sol Ring and a land are all
//     illegal targets. This is the most common misread of the card.
//   - A layer-5 colour change counts. A commander a Song of the Dryads
//     has made colourless has stopped being a legal target, and a
//     permanent something painted a second colour has too.
//
// The predicate is re-run at resolution like every target clause
// (CR 608.2b), so a permanent that gains a colour in response is no
// longer legal and the Verse fizzles — which is a real line of play,
// not a gap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b8f0cdf-572d-4025-b930-79291f7c35be",
		Name:         "Vanishing Verse",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target monocolored permanent", b41Monocolored()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
