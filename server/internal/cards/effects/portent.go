package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Portent — Sorcery {U}:
//
//	"Look at the top three cards of target player's library, then put
//	 them back in any order. You may have that player shuffle.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// #1298, on ADR 0088's put_in_library (2026-09-23 amendment): an
// ordered look at ANOTHER player's library top. LookAtLibraryThenPlace
// marks only the caster a knower and asks a `top` put_in_library whose
// chooser is the caster and whose cards are in the target's library —
// a reorder in place, no zone change. CR 401.4 then leaves the order
// known to the caster alone: three cards to one lane hide their order
// even from a library owner who had scryed them.
//
// "You may have that player shuffle" comes AFTER the order, in the
// put_in_library's Then — the caster decides having seen the cards —
// and the shuffle clears every knower in that library. The draw is a
// CR 603.7 delayed trigger at the next upkeep at the table, whoever's
// it is ("the next turn's upkeep", Arcane Denial's reading), and is
// scheduled whether or not the look found anything.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e744dbb6-2d56-462a-8d49-d79c73944048",
		Name:         "Portent",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (ScheduleDelayedTrigger{
				At:     game.StepUpkeep,
				Label:  "Portent — draw a card",
				Effect: portentDraw,
			}).Apply(ctx); err != nil {
				return err
			}
			var victim uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					victim = t.ID
				}
			}
			if victim == uuid.Nil {
				return nil
			}
			caster, source := item.Controller, ctx.Source()
			return LookAtLibraryThenPlace{
				Owner:     victim,
				N:         3,
				Placement: game.LibraryPlaceTop,
				Label:     "Portent — put them back in any order",
				Then: func(g *game.Game) error {
					g.QueueConfirmForEffect(game.ConfirmPrompt{
						Chooser:      caster,
						Source:       source,
						Question:     "Portent — have that player shuffle their library?",
						AcceptLabel:  "Shuffle",
						DeclineLabel: "Leave it",
						OnAccept: func(g *game.Game) error {
							return g.ShuffleLibraryForEffect(victim)
						},
					})
					return nil
				},
			}.Apply(ctx)
		},
	})
}

// portentDraw is the delayed trigger's effect: its controller draws a
// card. Package-level, so it captures nothing (ScheduleDelayedTrigger).
func portentDraw(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}
