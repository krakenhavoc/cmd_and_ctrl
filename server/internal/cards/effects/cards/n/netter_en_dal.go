package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Netter en-Dal — Creature — Human Spellshaper {W}, 1/1
// (Nemesis, #5618):
//
//	"{W}, {T}, Discard a card: Target creature can't attack this turn."
//
// The spellshaper cost (nemSpellshaperCost) on Change of Heart's
// restriction, as an activated ability's body
// (nemTargetRestrictedThisTurn with game.CantAttack). It ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cd65dcd-e36d-47b1-b011-961c79004cb0",
		Name:         "Netter en-Dal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}, Discard a card: Target creature can't attack this turn.",
			Cost:    effects.NemSpellshaperCost("{W}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Effect:  effects.NemTargetRestrictedThisTurn(game.CantAttack, "Netter en-Dal — can't attack this turn"),
		}},
	})
}
