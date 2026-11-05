package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foggy Swamp Visions — Sorcery {1}{B}{B}:
//
//	"As an additional cost to cast this spell, waterbend {X}. (While
//	 paying a waterbend cost, you can tap your artifacts and creatures
//	 to help. Each one pays for {1}.)
//	 Exile X target creature cards from graveyards. For each creature
//	 card exiled this way, create a token that's a copy of it. At the
//	 beginning of your next end step, sacrifice those tokens."
//
// Waterbender's Restoration's cost: Waterbend("{X}") layers an {X} on
// the {1}{B}{B}, each artifact or creature tapped toward it pays {1}
// (CR 701.67a), and the target count is that X. The still-legal targets
// are exiled at once, and only the cards that reached exile are copied
// (CR 400.7), all in one creation so the tokens enter together (CR
// 608.2f), their entry abilities triggering (the ruling). A copy takes
// the card's printed copiable values, an {X} in its cost counting as 0
// (CR 707.2). A delayed trigger (CR 603.7) sacrifices exactly those
// tokens at the beginning of the caster's next end step.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "16af0a22-833e-47c6-ac27-002702194595",
		Name:         "Foggy Swamp Visions",
		Completeness: CompletenessFull,
		TapCost:      Waterbend("{X}"),
		Targets:      targetsCountedByX(TargetCardInGraveyard("X target creature cards from graveyards", Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(foggySwampVisionsCopies))
		},
	})
}

// foggySwampVisionsCopies copies each card that reached exile.
var foggySwampVisionsCopies = CardTail("foggy-swamp-visions/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CreateTokenCopiesTogether{
		Controller: ctx.Controller(),
		Copies:     o.IDs,
		Then:       ctx.ItemTail(foggySwampVisionsSacrificeLater),
	}.Apply(ctx)
})

// foggySwampVisionsSacrificeLater is "At the beginning of your next end
// step, sacrifice those tokens", told the tokens that landed.
var foggySwampVisionsSacrificeLater = CardTail("foggy-swamp-visions/sacrifice-copies", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:              "Foggy Swamp Visions — sacrifice those tokens",
		ControllerTurnOnly: true,
		Cards:              o.IDs,
		Body:               sacrificeListedCardsBody,
	}.Apply(ctx)
})
