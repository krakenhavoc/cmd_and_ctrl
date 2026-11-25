package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Macabre Mockery — Instant {2}{B}{R} (Ravnica Allegiance):
//
//	"Put target creature card from an opponent's graveyard onto the battlefield
//	 under your control. It gets +2/+0 and gains haste until end of turn.
//	 Sacrifice it at the beginning of the next end step."
//
// Balduvian Atrocity's return-then-haste-then-sacrifice shape: the card enters
// under your control, and the keyed tail pumps it, gives it haste and
// schedules the sacrifice for the next end step, pinned to that object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d66f7ee7-e9a1-4714-b79c-8c86c6b6d5dc",
		Name:         "Macabre Mockery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from an opponent's graveyard", effects.Creature(), effects.Not(effects.YouOwn())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFromGraveyardTogether{
				Targets:    effects.LegalTargetCardIDs(ctx),
				Controller: ctx.Controller(),
				ThenTail:   macabreMockeryTail,
			}.Apply(ctx)
		},
	})
}

// macabreMockeryTail is the clauses after the return. An on-disk key: never
// renamed, never reused.
var macabreMockeryTail = effects.TogetherTail("macabre-mockery/pump-haste-sacrifice",
	func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
		for _, id := range entered {
			c, ok := effects.BattlefieldCardForEffect(ctx.Game, id)
			if !ok {
				continue
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Label: "Macabre Mockery — +2/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Macabre Mockery — haste until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: ctx.Item.Controller,
				Label:      "Macabre Mockery — sacrifice it",
				Body:       effects.Cmdr2S12SoulstokeSacrifice,
				Params:     game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
