package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disciple of Kangee — Creature — Human Wizard {2}{W}, 2/2 (Planeshift, S82
// OLD #5615):
//
//	"{U}, {T}: Target creature gains flying and becomes blue until end of turn."
//
// The flying grant (layer 6) and the colour set (layer 5, CR 613.1e) on the
// target if it is still legal as the ability resolves. The {T} keeps it from
// being activated the turn it arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "778fb830-2d2b-4f81-a193-e6b652981695",
		Name:         "Disciple of Kangee",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Target creature gains flying and becomes blue until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    game.AbilityCost{Mana: "{U}", Tap: true},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.Do(effects.Instruction(effects.OldB639TargetGainsFlyingAndBecomesBlue)),
		}},
	})
}
