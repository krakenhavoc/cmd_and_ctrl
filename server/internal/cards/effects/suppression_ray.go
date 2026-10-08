package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Suppression Ray // Orderly Plaza — modal double-faced card. This file
// is the FRONT face, Sorcery {3}{W/U}{W/U}:
//
//	"Tap all creatures target player controls. You may pay any amount of
//	 {E}. If you do, choose up to that many creatures tapped this way.
//	 Put a stun counter on each of them."
//
// The back face, Orderly Plaza, is registered with the MDFC land cycle
// in mdfc_lands.go.
//
// ADR 0129 §3 (#1995, owner decision 3): "tapped this way" is the
// creatures this spell tapped — one already tapped was not. The payment
// is the pay_amount prompt; then the caster chooses up to that many of
// those creatures, still on the battlefield, and each gets a stun
// counter. The bot pays one energy for each creature it tapped, as far as its energy goes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b592568b-11b0-4081-90a7-30cfb9c1ba80",
		Name:         "Suppression Ray",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			victim := item.Targets[0].ID
			var tapped []uuid.UUID
			for _, c := range ctx.Game.Battlefield.Cards {
				if c.Controller == victim && c.IsCreature() && !c.Tapped {
					tapped = append(tapped, c.InstanceID)
				}
			}
			for _, id := range tapped {
				if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			n := len(tapped)
			return PayEnergyAmount{
				Question: "Suppression Ray — pay any amount of {E}; that many creatures tapped this way get a stun counter",
				Unit:     game.PayAmountCounters,
				Goal: func(ctx *Context) int {
					return min(n, game.PlayerEnergy(ctx.Game.PlayerByIDForEffect(ctx.Controller())))
				},
				Then: func(ctx *Context, paid int) error {
					if paid <= 0 || n == 0 {
						return nil
					}
					most := paid
					if most > n {
						most = n
					}
					return ChoosePermanents{
						Of:       []uuid.UUID{victim},
						Question: "Suppression Ray — choose creatures tapped this way to get a stun counter",
						Candidates: func(g *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
							var still []uuid.UUID
							for _, id := range tapped {
								if onBattlefield(g, id) {
									still = append(still, id)
								}
							}
							hi := most
							if hi > len(still) {
								hi = len(still)
							}
							return still, 0, hi
						},
						Then: func(ctx *Context, picked game.PromptedPicks) error {
							for _, id := range picked.Cards() {
								if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), id, game.CounterStun, 1); err != nil {
									return err
								}
							}
							return nil
						},
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
