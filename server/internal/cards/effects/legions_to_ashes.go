package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legions to Ashes — Sorcery {1}{W}{B} (deck request #3655):
//
//	"Exile target nonland permanent an opponent controls and all tokens
//	 that player controls with the same name as that permanent."
//
// The name and the controller are read before anything is exiled, so a
// token target still finds its twins. The tokens are those the target's
// controller controls as the spell resolves (CR 608.2), whatever the
// target itself was. If the target is gone or illegal the spell does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc29cb25-57f5-4cac-9764-efa644b41296",
		Name:         "Legions to Ashes",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var victim uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					victim = t.ID
					break
				}
			}
			if victim == uuid.Nil {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(victim)
			if !ok {
				return nil
			}
			ids := []uuid.UUID{victim}
			for i := range ctx.Game.Battlefield.Cards {
				o := &ctx.Game.Battlefield.Cards[i]
				if o.InstanceID != victim && o.IsToken() && o.Controller == c.Controller && o.Name == c.Name {
					ids = append(ids, o.InstanceID)
				}
			}
			for _, id := range ids {
				if err := (ExileTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
