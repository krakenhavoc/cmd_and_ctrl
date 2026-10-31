package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thor, Asgard's Avenger — Legendary Creature — God Warrior Hero {2}{R}{R},
// 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 If another source you control would deal damage to an opponent or a
//	 permanent an opponent controls, it deals that much damage plus 1
//	 instead."
//
// Hawkeye, Young Avenger's replacement with a constant 1 and without his
// noncombat gate: combat and noncombat damage alike, from any source you
// control (a spell, an ability or a permanent, read from the damage
// event's snapshot so a creature that has left still counts, CR 608.2h)
// except Thor himself. Damage to you or to your permanents is untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44c2fefb-5de3-4406-955a-3fabc46357d0",
		Name:            "Thor, Asgard's Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil || ev.DamageSource == src.InstanceID {
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
			Label: "Thor, Asgard's Avenger: damage from another source plus 1",
		}},
	})
}
