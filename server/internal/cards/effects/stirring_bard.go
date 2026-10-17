package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2583836d-7fd3-47b3-9b69-6814efe54a90",
		Name:            "Stirring Bard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Stirring Bard"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Mantle of Inspiration — {T}: Target creature gains menace and haste until end of turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  stirringBardMantle,
		}},
	})
}

// stirringBardMantle grants menace and haste until end of turn to the
// target, if it is still legal (CR 608.2b).
func stirringBardMantle(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"menace", "haste"},
		Label:    "Stirring Bard — menace and haste until end of turn",
	}.Apply(ctx)
}
