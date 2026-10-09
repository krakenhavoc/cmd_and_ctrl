package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tomik, Izzet Sparkmage — Legendary Creature — Human Wizard {1}{R},
// 1/2:
//
//	"Prowess
//	 If a source you control would deal noncombat damage to an opponent
//	 or a permanent an opponent controls, it deals that much damage plus
//	 1 instead."
//
// Torbran's replacement with the colour clause dropped, the combat
// clause inverted and the bonus cut to one. The source is read the way
// Torbran reads it (damageSourceCharacteristics), so a spell on the
// stack or a creature that has already left is judged as it last was.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "16604fee-cd8d-41e4-8269-3c7ff756613d",
		Name:            "Tomik, Izzet Sparkmage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil || ev.IsCombatDamage {
					return false
				}
				return damageSourceControlledBy(ev, g, src.Controller) && damageHitsAnOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Tomik, Izzet Sparkmage: +1 noncombat damage",
		}},
	})
}
