package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Budoka Pupil // Ichiga, Who Topples Oaks — {1}{G}{G} Creature — Human
// Monk 2/2 // Legendary Creature — Spirit 4/3, a flip card (#3216, ADR
// 0151):
//
//	Top:     "Whenever you cast a Spirit or Arcane spell, you may put a
//	          ki counter on this creature.
//	          At the beginning of the end step, if there are two or more
//	          ki counters on this creature, you may flip it."
//	Flipped: "Trample
//	          Remove a ki counter from Ichiga: Target creature gets
//	          +2/+2 until end of turn."
//
// The ki cycle's shared triggers are kiFlipTriggers. The counters stay
// on the permanent through the flip (it is the same object), so Ichiga
// spends the ones Budoka Pupil collected. Ichiga is green and {1}{G}{G}
// (CR 710.1c).
//
// No simplification.
const budokaPupilOracleID = "6f571092-7244-4223-aa14-f5133cc000d8"

func init() {
	Register(Spec{
		OracleID:     budokaPupilOracleID,
		Name:         "Budoka Pupil",
		Completeness: CompletenessFull,
		Triggered:    kiFlipTriggers("Budoka Pupil"),
	})
	Register(Spec{
		OracleID:        budokaPupilOracleID + "#1",
		Name:            "Ichiga, Who Topples Oaks",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "Remove a ki counter from Ichiga: Target creature gets +2/+2 until end of turn.",
			Cost:    removeAKiCounter(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Ichiga, Who Topples Oaks — +2/+2"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
