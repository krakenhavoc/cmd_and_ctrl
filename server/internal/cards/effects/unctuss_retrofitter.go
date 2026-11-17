package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unctus's Retrofitter — Creature — Phyrexian Artificer {2}{U}, 2/3
// (Phyrexia: All Will Be One):
//
//	"Toxic 1
//	 When this creature enters, up to one target artifact you control
//	 becomes an artifact creature with base power and toughness 4/4 for as
//	 long as this creature remains on the battlefield."
//
// The target gains the Creature type (layer 4) and base 4/4 (layer 7b) for
// as long as the Retrofitter that made the trigger remains on the
// battlefield (CR 611.2b); a Retrofitter that left and returned while the
// trigger waited starts nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "998d8d8a-743d-43f9-8e9b-42460f170243",
		Name:            "Unctus's Retrofitter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Unctus's Retrofitter — up to one target artifact you control becomes a 4/4 artifact creature while this remains",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					d, ok := DurationWhileSourceRemains(ctx, ctx.Source())
					if !ok {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return ScopedEffectFor{
							Target:   t.ID,
							Mods:     append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(4, 4)...),
							Duration: d,
							Label:    "Unctus's Retrofitter — a 4/4 artifact creature while the Retrofitter remains",
						}.Apply(ctx)
					}
					return nil
				}),
				TargetPermanent("up to one target artifact you control", Artifact(), YouControl()).WithCount(0, 1)),
		},
	})
}
