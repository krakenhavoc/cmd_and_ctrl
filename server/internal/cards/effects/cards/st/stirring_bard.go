package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stirring Bard — Creature — Dragon Bard {3}{R}, 0/4:
//
//	"Defender
//	 When this creature enters, you take the initiative.
//	 Mantle of Inspiration — {T}: Target creature gains menace and
//	 haste until end of turn."
//
// Defender is engine-side; the ETB is TakeTheInitiative (ADR 0203,
// #4096). Mantle of Inspiration is an ordinary {T} ability: the engine
// holds it back while the Bard is summoning sick (CR 302.6), and the
// grant is GrantKeywordUntilEOT pinned to the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2583836d-7fd3-47b3-9b69-6814efe54a90",
		Name:            "Stirring Bard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Stirring Bard"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Mantle of Inspiration — {T}: Target creature gains menace and haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "menace", "haste")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  stirringBardMantle,
		}},
	})
}

// stirringBardMantle grants menace and haste until end of turn to the
// target, if it is still legal (CR 608.2b).
func stirringBardMantle(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"menace", "haste"},
		Label:    "Stirring Bard — menace and haste until end of turn",
	}.Apply(ctx)
}
