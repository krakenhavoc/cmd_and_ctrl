package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Neriv, Heart of the Storm — Legendary Creature — Spirit Dragon
// {1}{R}{W}{B}, 4/5 (Tarkir: Dragonstorm, #4463):
//
//	"Flying
//	 If a creature you control that entered this turn would deal damage,
//	 it deals twice that much damage instead."
//
// Gratuitous Violence's replacement (CR 614.1a) narrowed to a creature
// that entered the battlefield this turn. The source's type and
// controller are its last-known characteristics on the damage event, as
// for every catalog doubler; "entered this turn" is the turn's entry
// tally (Game.EnteredThisTurn), not summoning sickness, so a creature
// that entered on an opponent's turn is not doubled on yours, and a
// creature that has been on the battlefield since an earlier turn is not
// doubled however it got haste. Any damage counts, combat or not, to any
// player or permanent, and it is still dealt by the creature, not by
// Neriv (the printed ruling). Damage is divided or assigned before it is
// doubled (the printed ruling), which is what a replacement on each
// damage event does. With another replacement or prevention effect, the
// player or controller being dealt damage orders them (CR 616.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "66290104-ac65-4eb5-bf00-89b944249e07",
		Name:            "Neriv, Heart of the Storm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			RowLabel: "If a creature you control that entered this turn would deal damage, it deals twice that much damage instead.",
			Watches:  []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				ch, ok := damageSourceCharacteristics(ev, g)
				return ok && hasFold(ch.Types, "Creature") && ch.Controller == src.Controller &&
					g.EnteredThisTurn(ev.DamageSource)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Neriv, Heart of the Storm: twice that much damage",
		}},
	})
}
