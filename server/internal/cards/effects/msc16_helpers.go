package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc16_helpers.go — shared bodies for Marvel Super Heroes Commander, batch 16
// (#4455).

// msc16EndStepForTokens registers the tail of a "tapped and attacking token
// copy" trigger under `key` (an on-disk identity, ledgered "tail <key>"): at
// the beginning of the next end step the delayed trigger runs `body` over the
// tokens that landed (CR 603.7). Sandstorm Crasher's tail, with the body
// supplied.
func msc16EndStepForTokens(key, label string, body game.BodyRef) game.TailRef {
	return AttackingTokensThen(key, func(ctx *Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{Label: label, Cards: created, Body: body}.Apply(ctx)
	})
}

// msc16AttackingNonlegendaryCopies creates `n` tapped and attacking tokens
// that are copies of `copyOf`, except they aren't legendary and, when
// `addedSubtype` is not empty, are that subtype in addition to their other
// types. They were never declared as attackers (CR 508.4), so nothing that
// watches attacks sees them; their controller chooses what each attacks,
// the trigger's own defender first.
func msc16AttackingNonlegendaryCopies(ctx *Context, item *game.StackItem, copyOf uuid.UUID, n int, addedSubtype, label string, tail game.TailRef) error {
	tmpl, ok := TokenCopyTemplate(ctx.Game, copyOf)
	if !ok {
		return nil
	}
	tmpl.TypeLine = removeLegendaryFromTypeLine(tmpl.TypeLine)
	if addedSubtype != "" {
		tmpl.TypeLine = withSubtypeInAddition(tmpl.TypeLine, addedSubtype)
	}
	return CreateTokensAttackingYourChoice{
		Template: tmpl,
		Copy:     true,
		N:        n,
		Tapped:   true,
		Prefer:   triggerAttackTarget(item),
		Label:    label,
		Tail:     game.KeyedTail{Ref: tail},
	}.Apply(ctx)
}
