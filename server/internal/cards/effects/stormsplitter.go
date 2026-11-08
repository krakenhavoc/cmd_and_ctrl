package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormsplitter — Creature — Otter Wizard {3}{R}, 1/4 (Bloomburrow,
// #4467):
//
//	"Haste
//	 Whenever you cast an instant or sorcery spell, create a token that's
//	 a copy of this creature. Exile that token at the beginning of the
//	 next end step."
//
// The cast trigger resolves before the spell. The token is a copy of the
// Stormsplitter's copiable values as the trigger resolves, or as it last
// existed if it has left (CR 707.2, CR 608.2h), as Colorstorm Stallion's
// is, so it has haste and this same trigger. Inalla's delayed exile
// (exileListedCardsBody) takes that token at the beginning of the next
// end step, this turn's if the spell was cast before it (CR 603.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4a9ae0d9-5ee7-4289-9333-ef0bcc3079ca",
		Name:            "Stormsplitter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Stormsplitter — create a token copy of it; exile that token at the beginning of the next end step",
				stormsplitterCopy),
		},
	})
}

// stormsplitterCopy makes the token copy and schedules its exile.
func stormsplitterCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	tmpl, ok := g.TokenCopyTemplateForObjectForEffect(ref)
	if !ok {
		return nil
	}
	made, err := g.CreateTokensForEffect(ctx.Controller(), copiedToken(tmpl), 1, game.TokenEntryOptions{})
	if err != nil || len(made) == 0 {
		return err
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Controller(),
		Label:      "Stormsplitter — exile the token copy",
		Cards:      made,
		Body:       exileListedCardsBody,
	}.Apply(ctx)
}
