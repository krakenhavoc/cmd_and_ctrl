package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// legend_rule_exemption.go — the card-side sentences over
// game/legend_rule_exemption.go (CR 704.5j, #2177).

// LegendRuleDoesntApplyToYours is "The "legend rule" doesn't apply to
// permanents you control." (Mirror Box, Sakashima of a Thousand Faces).
func LegendRuleDoesntApplyToYours() []game.LegendRuleExemption {
	return []game.LegendRuleExemption{{
		Label: "The \"legend rule\" doesn't apply to permanents you control.",
		Whose: game.LegendRuleExemptYours,
	}}
}

// LegendRuleDoesntApplyToYourCreatures is "The "legend rule" doesn't apply
// to creatures you control." (Council of Reeds, #4707). Narrower than
// LegendRuleDoesntApplyToYours: a legendary noncreature permanent you
// control, and an opponent's legend, still answer to CR 704.5j.
func LegendRuleDoesntApplyToYourCreatures() []game.LegendRuleExemption {
	return []game.LegendRuleExemption{{
		Label: "The \"legend rule\" doesn't apply to creatures you control.",
		Whose: game.LegendRuleExemptYourCreatures,
	}}
}

// LegendRuleDoesntApply is "The "legend rule" doesn't apply." — every
// player's permanents (Mirror Gallery).
func LegendRuleDoesntApply() []game.LegendRuleExemption {
	return []game.LegendRuleExemption{{
		Label: "The \"legend rule\" doesn't apply.",
		Whose: game.LegendRuleExemptEveryone,
	}}
}
