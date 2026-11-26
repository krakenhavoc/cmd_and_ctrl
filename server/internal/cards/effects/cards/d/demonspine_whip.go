package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonspine Whip — Artifact — Equipment {B}{R} (Alara Reborn, #5586):
//
//	"{X}: Equipped creature gets +X/+0 until end of turn.
//	 Equip {1}"
//
// The pump is an ability of the Equipment itself, with X announced as it is
// activated; it lands on the creature the Equipment is attached to as the
// ability resolves, and does nothing for an unattached Whip.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a588f047-02b4-4c6b-bce2-6e4d564296ec",
		Name:         "Demonspine Whip",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{X}: Equipped creature gets +X/+0 until end of turn.",
				Cost:    effects.ManaCost("{X}"),
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					host := g.AttachedHostOf(&src)
					if host == nil || ctx.X() <= 0 {
						return nil
					}
					return effects.BoostUntilEOT{Target: host.InstanceID, Power: ctx.X(), Label: "Demonspine Whip — equipped creature gets +X/+0"}.Apply(ctx)
				},
			},
			effects.EquipAbility("{1}"),
		},
	})
}
