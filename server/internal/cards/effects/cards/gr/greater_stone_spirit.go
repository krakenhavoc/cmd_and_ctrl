package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greater Stone Spirit — Creature — Elemental Spirit {4}{R}{R}, 4/4 (Coldsnap, S82 OLD b560):
//
//	"This creature can't be blocked by creatures with flying.
//	 {2}{R}: Until end of turn, target creature gets +0/+2 and gains "{R}: This
//	 creature gets +1/+0 until end of turn.""
//
// One effect at one timestamp: the +0/+2 and the granted pump row (Banishing
// Knack's grant bundle), both lasting until end of turn on the creature chosen.
//
// No simplification.
func init() {
	const bundle = "greater-stone-spirit/pump"
	effects.Register(effects.Spec{
		OracleID:     "b36c66d3-24ae-4291-8f8b-8716e354975e",
		Name:         "Greater Stone Spirit",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.HasKeyword("flying"), "creatures with flying"),
		},
		Grants: []effects.AbilityGrant{{
			Key:       bundle,
			Activated: []effects.ActivatedAbility{effects.ThisGetsRow("{R}", 1, 0, "Greater Stone Spirit")},
			Text:      "{R}: This creature gets +1/+0 until end of turn.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: Until end of turn, target creature gets +0/+2 and gains \"{R}: This creature gets +1/+0 until end of turn.\"",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 2)),
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantAbilitiesFor{
					Target: target,
					Keys:   []string{bundle},
					Also:   []game.Mod{game.ModifyPTMod(0, 2)},
					Label:  "Greater Stone Spirit — +0/+2 and \"{R}: +1/+0\" until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
