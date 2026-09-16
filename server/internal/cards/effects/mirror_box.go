package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirror Box — Artifact {3} (EDHREC rank 1699):
//
//	"The "legend rule" doesn't apply to permanents you control.
//	 Each legendary creature you control gets +1/+1.
//	 Each nontoken creature you control gets +1/+1 for each other
//	 creature you control with the same name as that creature."
//
// The legends deck's anthem. Two of its three lines are layer 7c
// statics: +1/+1 for every legendary creature the controller
// controls (post-layer supertypes, so a Clone of a legend counts),
// and +1/+1 per OTHER creature the controller controls sharing the
// creature's name — Coat of Arms' shape keyed on the name rather
// than the type, read post-layer so a token copy counts under the
// name it has now, and applied to nontoken creatures only, as
// printed. Two Mirror Boxes double both bonuses, as printed.
//
// Sandbox simplification: the first line is NOT modelled. The legend
// rule is a state-based action the engine runs from the SBA loop
// (legend_rule.go) with no per-controller exemption hook, and the
// only way a card file could switch it off — stripping "Legendary"
// in layer 4 — would also switch off this card's own second line
// and every other legendary-matters effect at the table, which is a
// different card. So a second copy of a legend still prompts the
// legend rule. Weaker than printed, never stronger; the two anthems
// are complete.
func init() {
	Register(Spec{
		OracleID:     "3bed1944-58dc-4679-9aee-7be4d94fb55c",
		Name:         "Mirror Box",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The legend rule still applies to your permanents; only the two +1/+1 bonuses are in effect."},
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && isLegendary(target)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && !IsToken(*target)
				},
				Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
					n := b15OtherCreaturesControlledWithSameName(g, source.Controller, target.InstanceID, c.Name)
					c.Power += n
					c.Toughness += n
				},
			},
		},
	})
}
