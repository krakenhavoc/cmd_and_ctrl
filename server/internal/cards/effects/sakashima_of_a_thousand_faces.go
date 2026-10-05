package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sakashima of a Thousand Faces — Legendary Creature — Human Rogue
// {3}{U}, 3/1:
//
//	"You may have Sakashima enter as a copy of another creature you
//	 control, except it has Sakashima's other abilities.
//	 The "legend rule" doesn't apply to permanents you control.
//	 Partner (You can have two commanders if both have partner.)"
//
// The copy is Sakashima the Impostor's shape (ADR 0043): a copy effect
// with an "except it has" clause that rides in the copiable values as
// an ability grant (CR 707.9a). The bundle carries the legend-rule
// exemption, so a copy of a Grizzly Bears still exempts its controller.
// The name is NOT kept, so a copy of your own legend is a second
// legend of that name, which the exemption is there to allow.
//
// Partner is not modelled, as for every partner card: the deck checker
// allows one commander only. It is deck construction, not battlefield
// behaviour.
const sakashimaFacesGrant = "sakashima-of-a-thousand-faces/legend-rule"

func init() {
	Register(Spec{
		OracleID:             "8ecdaf4b-4442-42da-9714-4257a83faf50",
		Name:                 "Sakashima of a Thousand Faces",
		Completeness:         CompletenessCaveats,
		Caveats:              []string{"Partner does nothing — a deck can still only have one commander."},
		LegendRuleExemptions: LegendRuleDoesntApplyToYours(),
		Grants: []AbilityGrant{{
			Key:              sakashimaFacesGrant,
			Text:             "The \"legend rule\" doesn't apply to permanents you control.",
			LegendRuleExempt: LegendRuleDoesntApplyToYours(),
		}},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Sakashima of a Thousand Faces",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.IsCreature() && c.Controller == controller
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.GrantAbility(sakashimaFacesGrant)
				},
			),
		},
	})
}
