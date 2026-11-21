package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Press the Enemy — Instant {2}{U}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Return target spell or nonland permanent an opponent controls to its
//	 owner's hand. You may cast an instant or sorcery spell with equal or lesser
//	 mana value from your hand without paying its mana cost."
//
// Sink into Stupor's bounce (a spell is returned from the stack without being
// countered), then Cecily's free cast of an instant or sorcery from your hand,
// capped at the returned object's mana value. That value is read BEFORE the
// bounce (CR 608.2h): a spell's counts its X as cast, a permanent's is its
// own. The cast is offered after any prompt the bounce asks (CR 608.2c) and
// only if the target was still legal; with nothing returned the spell is
// countered on resolution (CR 608.2b) and offers nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "347a3622-4425-4ba4-821d-ef3d00beb058",
		Name:         "Press the Enemy",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetSpellOrPermanent("target spell or nonland permanent an opponent controls",
			effects.OpponentControls(), effects.And(effects.Nonland(), effects.OpponentControls())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					mv := effects.GapB131ReturnedManaValue(ctx, t.ID)
					ctx.RecordStep(pressTheEnemyRecord, []uuid.UUID{t.ID}, mv)
				}
				return effects.ReturnTargetSpellOrPermanentToHand(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				r := ctx.StepRecord(pressTheEnemyRecord)
				if len(r.IDs) == 0 || len(r.Ns) == 0 {
					return nil
				}
				effects.OfferFreeHandCast(ctx, ctx.Controller(), pressTheEnemyOffer, "an instant or sorcery spell with equal or lesser mana value",
					effects.FreeInstantOrSorcery().WithManaValueAtMost(r.Ns[0]))
				return nil
			}),
		),
	})
}

const pressTheEnemyRecord = "press-the-enemy/returned"

// pressTheEnemyOffer is the free-cast prompt's key. An on-disk key: never
// renamed, never reused.
var pressTheEnemyOffer = effects.RegisterFreeHandCastOffer("choose-cards/press-the-enemy-cast", "Press the Enemy")
