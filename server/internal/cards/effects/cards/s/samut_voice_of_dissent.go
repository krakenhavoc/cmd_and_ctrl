package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Samut, Voice of Dissent — Legendary Creature — Human Warrior {3}{R}{G}, 3/4
// (Secret Lair Drop):
//
//	"Flash
//	 Double strike, vigilance, haste
//	 Other creatures you control have haste.
//	 {W}, {T}: Untap another target creature."
//
// The keywords are printed data; the grant is a layer-6 haste for the other
// creatures its controller has; the untap is the shared untap-the-first-legal-
// target body over a target other than Samut herself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b1eee275-5106-4c8d-a497-d7945ad81d74",
		Name:            "Samut, Voice of Dissent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "double strike", "vigilance", "haste"},
		Static:          []game.StaticAbility{effects.KeywordGrant(effects.OtherCreaturesYouControl, "haste")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Untap another target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect:  effects.UntapFirstLegalTarget,
		}},
	})
}
