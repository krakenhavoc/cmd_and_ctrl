package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Palm Efreet — Creature — Efreet {2}{R}{R}, 2/2 (Mirage, #5633):
//
//	"{1}{R}{R}: This creature deals 2 damage to target creature with flying and
//	 that creature loses flying until end of turn."
//
// The creature loses flying whether or not the damage was prevented, as long as
// the target is still legal (CR 608.2b); a layer-6 effect with its own
// timestamp (CR 613.1f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55cbc0e8-eeec-433a-b70c-3501238b658b",
		Name:         "Burning Palm Efreet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}: This creature deals 2 damage to target creature with flying and that creature loses flying until end of turn.",
			Cost:    effects.ManaCost("{1}{R}{R}"),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Damage: 2, UntilEndOfTurn: game.TargetHarm}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok {
					return nil
				}
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
					return err
				}
				return effects.UntilEndOfTurn(ctx, t.ID, nil, "Burning Palm Efreet — loses flying until end of turn",
					game.RemoveKeywordsMod("flying"))
			},
		}},
	})
}
