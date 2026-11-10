package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Esper — Plane — Alara (#4870):
//
//	"Artifact spells cost {1} less to cast.
//	 Whenever chaos ensues, creatures you control that are white, blue,
//	 and/or black become artifacts in addition to their other types until
//	 end of turn. Then each artifact creature you control gains vigilance,
//	 menace, and lifelink until end of turn."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// cost reduction is every player's, and takes generic mana only
// (the rulings). The chaos ability (CR 311.7) is two locked sets, each
// determined as it resolves (CR 611.2c, the ruling): the white, blue and/or
// black creatures become artifacts first, so they are among the artifact
// creatures the second sentence reaches.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4eebfaf3-18da-4661-b6a5-d874585e170a",
		Name:         "Esper",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Artifact spells cost {1} less to cast.", ArtifactSpell()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Esper — creatures you control that are white, blue, and/or black become artifacts until end of turn; then each artifact creature you control gains vigilance, menace, and lifelink until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (ScopedEffectFor{
						Match:    And(Creature(), YouControl(), Or(OfColor("W"), OfColor("U"), OfColor("B"))),
						Mods:     []game.Mod{game.AddTypesMod("Artifact")},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Esper — becomes an artifact until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
					g.RecomputeLayersIfStaleLocked()
					return GrantKeywordUntilEOT{Match: And(Artifact(), Creature(), YouControl()), Keywords: []string{"vigilance", "menace", "lifelink"},
						Label: "Esper — vigilance, menace, and lifelink until end of turn"}.Apply(ctx)
				}),
		},
	})
}
