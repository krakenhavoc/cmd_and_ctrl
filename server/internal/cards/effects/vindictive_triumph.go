package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vindictive Triumph — Instant {W}{B}{B} (Reality Fracture, tracker #2795):
//
//	"Exile target creature or planeswalker. If that permanent's mana value
//	 was 3 or less, return it to the battlefield tapped under your control.
//	 Exile it at the beginning of the next end step."
//
// The mana value is read BEFORE the exile (CR 608.2h last known
// information), the return runs as the exile's continuation (so a commander
// that stops to ask about the command zone is not returned early), and it
// comes back as a new object under the caster's control, tapped. The
// delayed exile is scheduled against that NEW object. A token is exiled
// and ceases to exist, so there is nothing to return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6b478d1-5015-49ba-b1aa-cbc2b08107a6",
		Name:         "Vindictive Triumph",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			target := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			mv, known := ctx.Game.ManaValueForEffect(card)
			small := known && mv <= 3
			caster, source := ctx.Controller(), ctx.Source()
			return ExileTarget{
				Target: target,
				Then: func(ctx *Context, exiled bool) error {
					if !exiled || !small {
						return nil
					}
					return ReturnFromExile{
						Target:     target,
						Controller: caster,
						Tapped:     true,
						Then: func(g *game.Game, entered uuid.UUID) error {
							if entered == uuid.Nil {
								return nil
							}
							g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
								Controller:   caster,
								SourceCardID: source,
								Label:        "Vindictive Triumph — exile the returned permanent",
								At:           game.StepEnd,
								Cards:        []uuid.UUID{entered},
								Body:         exileListedCardsBody,
							})
							return nil
						},
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
