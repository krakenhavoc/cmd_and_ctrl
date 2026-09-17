package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Vault —
//
// "This artifact doesn't untap during your untap step. At the beginning of
// your upkeep, you may pay {4}. If you do, untap this artifact. At the
// beginning of your draw step, if this artifact is tapped, it deals 1 damage
// to you. {T}: Add {C}{C}{C}."
//
// The draw-step trigger checks its condition both when it triggers and when it
// resolves. If the source has left, its last known tapped state is not yet
// available at resolution, so that case remains a declared caveat.
func init() {
	Register(Spec{
		OracleID:              "736892cb-a34b-4bb9-b56c-e26e3db207a2",
		Name:                  "Mana Vault",
		Completeness:          CompletenessCaveats,
		Caveats:               []string{"If Mana Vault leaves the battlefield before its draw-step ability resolves, it deals no damage."},
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		ManaAbilities:         []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}{C}{C}", Label: "Add {C}{C}{C}"}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Mana Vault — pay {4} to untap", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{4}",
					Question: "Mana Vault — pay {4} to untap it?",
					OnPay: func(ctx *Context) error {
						return UntapTarget{Target: item.SourceCardID}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
			On(game.EventBeginDrawStep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && ev.Kind == game.EventBeginDrawStep && source.Tapped
			}, "Mana Vault — deal 1 damage", func(g *game.Game, item *game.StackItem) error {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.InstanceID == item.SourceCardID && c.Tapped {
						return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					}
				}
				return nil
			}),
		},
	})
}
