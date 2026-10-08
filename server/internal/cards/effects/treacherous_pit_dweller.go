package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "899387c5-6781-485b-b8bc-aa89d3b97413",
		Name:            "Treacherous Pit-Dweller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, AllOf(Self, EnteredFromAGraveyard),
				"Treacherous Pit-Dweller — target opponent gains control of it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return GainControl{
							Target:     item.SourceCardID,
							Controller: t.ID,
							Duration:   game.IndefiniteDuration(),
							Label:      "Treacherous Pit-Dweller — target opponent gains control of it",
						}.Apply(ctx)
					}
					return nil
				}), TargetPlayer("target opponent", Opponent())),
		},
	})
}
