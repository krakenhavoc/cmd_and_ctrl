package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brothers Yamazaki — Legendary Creature — Human Samurai {2}{R}, 2/1
// (Foundations Jumpstart):
//
//	"Bushido 1
//	 If there are exactly two permanents named Brothers Yamazaki on the
//	 battlefield, the "legend rule" doesn't apply to them.
//	 Each other creature named Brothers Yamazaki gets +2/+2 and has haste."
//
// Bushido is the shared trigger. The exemption is the new
// LegendRuleExemptNamedPair (CR 704.5j): read live, it covers both
// permanents while exactly two bear the name, and a third switches it
// off, so the controller chooses which to keep. The pump is one static
// per layer (7c and 6) over "another creature with that name", whoever
// controls it.
//
// No simplification.
func init() {
	const name = "Brothers Yamazaki"
	other := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.InstanceID != source.InstanceID && game.PermanentHasName(*target, name)
	}
	Register(Spec{
		OracleID:     "5c820811-1775-4ea4-b003-721750a4d90f",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Bushido(1, name)},
		LegendRuleExemptions: []game.LegendRuleExemption{{
			Label: "If there are exactly two permanents named Brothers Yamazaki on the battlefield, the \"legend rule\" doesn't apply to them.",
			Whose: game.LegendRuleExemptNamedPair,
		}},
		Static: []game.StaticAbility{
			{
				Label:     "Each other creature named Brothers Yamazaki gets +2/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: other,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Label:     "Each other creature named Brothers Yamazaki has haste.",
				Layer:     game.Layer6Ability,
				AppliesTo: other,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = append(c.Abilities, "haste")
				},
			},
		},
	})
}
