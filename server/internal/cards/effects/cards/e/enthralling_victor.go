package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enthralling Victor — Creature — Human Warrior {3}{R}, 3/2 (Battlebond, #5529):
//
//	When this creature enters, gain control of target creature an opponent controls with power 2 or less until end of turn. Untap that creature. It gains haste until end of turn.
//
// Zealous Conscripts' steal, narrowed to a creature an opponent controls with power 2 or less; the target is re-checked on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa58591b-240f-4527-8308-7b87263e784f",
		Name:         "Enthralling Victor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Enthralling Victor — gain control of target creature an opponent controls with power 2 or less until end of turn, untap it, it gains haste",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.GainControl{
						Target:   id,
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Enthralling Victor — gain control until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Enthralling Victor — haste until end of turn"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls with power 2 or less", effects.OpponentControls(), effects.PowerLE(2))),
		},
	})
}
