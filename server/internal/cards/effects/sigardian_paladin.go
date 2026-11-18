package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sigardian Paladin — Creature — Human Knight {2}{G}{W}, 4/4 (Crimson Vow):
//
//	"As long as you've put one or more +1/+1 counters on a creature this
//	 turn, this creature has trample and lifelink.
//	 {1}{G}{W}: Target creature you control with a +1/+1 counter on it gains
//	 trample and lifelink until end of turn."
//
// The static is a layer-6 grant judged on the Paladin alone against this
// turn's event log (a +1/+1 placement by its controller on a creature), so it
// ends with the turn. The activated ability is a plain until-end-of-turn
// grant on a creature that carries a +1/+1 counter at announce and resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "472756f7-7fb8-4630-be14-d78bb9c99157",
		Name:         "Sigardian Paladin",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "As long as you've put one or more +1/+1 counters on a creature this turn, this creature has trample and lifelink.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && vowB09YouPutPlusOneThisTurn(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, kw := range []string{"trample", "lifelink"} {
					if !slices.Contains(c.Abilities, kw) {
						c.Abilities = append(c.Abilities, kw)
					}
				}
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}{W}: Target creature you control with a +1/+1 counter on it gains trample and lifelink until end of turn.",
			Cost:    ManaCost("{1}{G}{W}"),
			Targets: TargetCreature("target creature you control with a +1/+1 counter on it", YouControl(), vowB09HasPlusOneCounter()),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "trample", "lifelink")),
			Effect:  vowB09GrantUntilEOTOnTarget("Sigardian Paladin — trample and lifelink until end of turn", "trample", "lifelink"),
		}},
	})
}
