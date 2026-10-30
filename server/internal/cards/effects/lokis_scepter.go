package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loki's Scepter — {2}{R} Legendary Artifact (Marvel Super Heroes
// Commander, #4455):
//
//	"When Loki's Scepter enters, gain control of target creature until end
//	 of turn. Untap that creature. Until end of turn, it becomes a Villain
//	 in addition to its other types and gains haste.
//	 {T}: Add one mana of any color."
//
// Coercive Recruiter's four steps (msc13ThreatenAndPaint is the shared
// body): steal until end of turn, untap, haste, and a layer-4 type add
// pinned to the creature's own entry. "Villain" is a word of the Scepter's
// own text. The mana ability is the any-color tap of Chromatic Lantern.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f8941d8-f8f7-4ec3-86c3-88af8829080b",
		Name:         "Loki's Scepter",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Loki's Scepter — gain control of target creature until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return msc13ThreatenAndPaint(ctx, t.ID, "Loki's Scepter", ctx.TextSubtype("Villain"))
					}
					return nil
				}), TargetCreature("target creature")),
		},
	})
}
