package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos Charm — Instant for {B}{R}:
//
//	"Choose one —
//	 • Exile target player's graveyard.
//	 • Destroy target artifact.
//	 • Each creature deals 1 damage to its controller."
//
// S20 sub-PR 4's showcase modal card: three options, two of them
// targeting different things (a player, an artifact) and one
// untargeted. The chosen option's TargetSpec is what the engine
// validates against and what the client's picker highlights, so
// picking "destroy target artifact" on a board with no artifacts
// shows the picker option greyed out.
func init() {
	Register(Spec{
		OracleID: "5e62b51d-faec-4aa0-9504-cf2c282d08ea",
		Name:     "Rakdos Charm",
		Modes: ChooseOne(
			Mode("Exile target player's graveyard.", TargetPlayer("target player")),
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Each creature deals 1 damage to its controller."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				p := ctx.PlayerByID(item.Targets[0].ID)
				if p == nil || p.Graveyard == nil {
					return nil
				}
				// Snapshot the IDs first — ExileTarget mutates the zone.
				ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
				for _, c := range p.Graveyard.Cards {
					ids = append(ids, c.InstanceID)
				}
				for _, id := range ids {
					if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
			case ctx.HasMode(1):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
			case ctx.HasMode(2):
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if !c.IsCreature() {
						continue
					}
					if err := ctx.Game.DealDamageToPlayerForEffect(c.InstanceID, c.Controller, 1); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
