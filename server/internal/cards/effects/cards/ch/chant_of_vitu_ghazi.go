package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chant of Vitu-Ghazi — Instant {6}{W}{W}:
//
//	"Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	 Prevent all damage that would be dealt by creatures this turn. You gain life equal to the damage prevented this way."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): Ethereal Haze's shield with a CR
// 615.5 follow-up: the life is gained each time the shield prevents
// damage, once per damage instance with what it prevented that time (its
// ruling, CR 615.13).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e245a736-5f65-4159-8e39-e279e1f8794f",
		Name:         "Chant of Vitu-Ghazi",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		OnResolve: effects.SourceShieldSpell(effects.PreventDamageFromSource{Protect: effects.ShieldAnything, Then: effects.PreventedGainLifeBody,
			Queries: []game.PermanentQuery{effects.QueryTypes("creature")}}),
	})
}
