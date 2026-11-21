package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Awaken the Sleeper — Sorcery {3}{R} (Phyrexia: All Will Be One):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn. If it's equipped, you may destroy all
//	 Equipment attached to that creature."
//
// #5479, S82. Act of Treason's three clauses, then the rider: "if it's
// equipped" is read after the theft, as the clause is reached, and the
// "you may" is MayChoice's resolution-time yes/no, so it is asked only of a
// creature that really has an Equipment on it. The yes destroys every
// Equipment attached to it as the answer is given (CR 701.8), each one
// still on the battlefield and still attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4df1710-7402-45a7-966f-b51ddc087fbe",
		Name:         "Awaken the Sleeper",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Awaken the Sleeper — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Awaken the Sleeper — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if c, ok := ctx.Game.LookupCardForEffect(target); !ok || !effects.IsEquippedCreature(ctx.Game, c) {
				return nil
			}
			return effects.MayChoice{
				Question: "Awaken the Sleeper — destroy all Equipment attached to that creature?",
				Then:     awakenTheSleeperDestroyEquipment,
				Carry:    []uuid.UUID{target},
			}.Apply(ctx)
		},
	})
}

// awakenTheSleeperDestroyEquipment is the "yes": every Equipment still
// attached to the creature is destroyed. An on-disk key.
var awakenTheSleeperDestroyEquipment = effects.MayYes("confirm/awaken-the-sleeper-destroy-equipment", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	creature := effects.CarryAt(r, 0)
	var equipment []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.HasSubtype("Equipment") && c.IsAttachedTo(creature) {
			equipment = append(equipment, c.InstanceID)
		}
	}
	for _, id := range equipment {
		if err := (effects.DestroyTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
})
