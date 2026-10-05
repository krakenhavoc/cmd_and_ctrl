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
// The first line is the CR 704.5j exemption: a player-scoped static
// (LegendRuleExemptions) the legend-rule state-based action reads off
// the battlefield each time it runs (game/legend_rule_exemption.go,
// #2177). Only the controller's permanents are exempt, so an opponent's
// duplicate legends still go. When the Box leaves, the next check
// applies the rule normally and the controller chooses which to keep.
func init() {
	Register(Spec{
		OracleID:             "3bed1944-58dc-4679-9aee-7be4d94fb55c",
		Name:                 "Mirror Box",
		Completeness:         CompletenessFull,
		LegendRuleExemptions: LegendRuleDoesntApplyToYours(),
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
