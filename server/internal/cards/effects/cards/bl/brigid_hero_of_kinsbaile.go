package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brigid, Hero of Kinsbaile — Legendary Creature — Kithkin Archer {2}{W}{W},
// 2/3 (Lorwyn):
//
//	"First strike
//	 {T}: Brigid deals 2 damage to each attacking or blocking creature target
//	 player controls."
//
// The creatures are the ones the chosen player controls as the ability
// resolves (CR 608.2h), all hit in one damage event; the player is re-checked
// at resolution (CR 608.2b). It declares a damage sweep that spares everything
// not in combat (ADR 0126 §6). First strike is the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "929ad5c3-4efe-44c3-93ac-eda432814385",
		Name:            "Brigid, Hero of Kinsbaile",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Brigid deals 2 damage to each attacking or blocking creature target player controls.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					return effects.DamageEachMatching(ctx,
						effects.And(effects.Creature(), effects.AttackingOrBlocking(), effects.ControlledBy(t.ID)), 2)
				}
				return nil
			},
		}},
	})
}
