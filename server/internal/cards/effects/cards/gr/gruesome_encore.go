package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gruesome Encore — Sorcery {2}{B} (Mirrodin Besieged, S82 OLD b465, #5576):
//
//	"Put target creature card from an opponent's graveyard onto the
//	 battlefield under your control. It gains haste. Exile it at the beginning
//	 of the next end step. If that creature would leave the battlefield, exile
//	 it instead of putting it anywhere else."
//
// The sentences resolve in printed order (CR 608.2c). The creature enters
// under the caster's control (From the Catacombs' route); the entry's keyed
// continuation then, for the creature that entered, grants haste until end
// of turn, schedules the next end step's exile of that object, and registers
// the per-object leave-the-battlefield redirect (Whip of Erebos's shared
// ExileInsteadOfLeavingBattlefield). A target gone by resolution fizzles the
// spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f87bf51e-6218-4418-ae3e-98055e9601e4",
		Name:         "Gruesome Encore",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from an opponent's graveyard", effects.Creature(), func(_ *game.Game, caster uuid.UUID, c game.Card) bool { return c.Owner != caster }),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(t.ID, ctx.Controller(), false,
				nil, ctx.Tail(gruesomeEncoreEntered)); err != nil {
				return effects.NothingIfGone(err)
			}
			return nil
		},
	})
}

// gruesomeEncoreEntered is the entry's continuation, a keyed tail: the
// creature that entered gets haste, the end-step exile and the exile
// redirect. An on-disk key.
var gruesomeEncoreEntered = effects.EnteredTail("gruesome-encore/entered", func(ctx *effects.Context, entered uuid.UUID, o game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	if err := (effects.GrantKeywordUntilEOT{
		Target:   entered,
		Keywords: []string{"haste"},
		Label:    "Gruesome Encore — the returned creature has haste",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.ScheduleDelayedTrigger{
		Label: "Gruesome Encore — exile the returned creature",
		Cards: []uuid.UUID{entered},
		Body:  effects.ExileListedCardsBody,
	}).Apply(ctx); err != nil {
		return err
	}
	effects.ExileInsteadOfLeavingBattlefield(ctx.Game, entered, o.Tail.Player,
		"Gruesome Encore: if that creature would leave the battlefield, exile it instead")
	return nil
})
