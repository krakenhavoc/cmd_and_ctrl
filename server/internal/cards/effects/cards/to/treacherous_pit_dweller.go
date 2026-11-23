package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treacherous Pit-Dweller — Creature — Demon {B}{B}, 4/3:
//
//	"When this creature enters from a graveyard, target opponent gains
//	 control of it.
//	 Undying"
//
// The trigger reads where the creature entered from (EventETB.EnteredFrom,
// ADR 0113 amendment 2026-10-08), so it fires when undying brings it
// back and never when it is cast. The control change has no stated
// duration (CR 611.2a), and is pinned to the object that entered: if it
// left in response, nothing is taken (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "899387c5-6781-485b-b8bc-aa89d3b97413",
		Name:            "Treacherous Pit-Dweller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.EnteredFromAGraveyard),
				"Treacherous Pit-Dweller — target opponent gains control of it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.GainControl{
							Target:     item.SourceCardID,
							Controller: t.ID,
							Duration:   game.IndefiniteDuration(),
							Label:      "Treacherous Pit-Dweller — target opponent gains control of it",
						}.Apply(ctx)
					}
					return nil
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
