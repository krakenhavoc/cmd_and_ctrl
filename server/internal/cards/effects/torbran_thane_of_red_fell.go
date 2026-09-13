package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torbran, Thane of Red Fell — Legendary Creature — Dwarf Noble,
// {1}{R}{R}{R}, 2/4 (EDHREC rank 1003):
//
//	"If a red source you control would deal damage to an opponent or
//	 a permanent an opponent controls, it deals that much damage plus
//	 2 instead."
//
// The mono-red go-wide commander: every 1/1 Goblin hits for three.
// Twinflame Tyrant's replacement with an addition in place of a
// doubling and one more gate — the source must be RED, read off the
// source card wherever it is (a red creature that dies mid-combat is
// still red in the graveyard when its damage event arrives). Combat
// and non-combat damage alike, as printed; damage to yourself or to
// your own permanents is untouched. Torbran is himself a red source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c3495bf-02e7-4ad9-949d-92eb3d2b662a",
		Name:         "Torbran, Thane of Red Fell",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.DamageSource)
				if !ok || dealer.Controller != src.Controller || !dealer.HasColor("R") {
					return false
				}
				return damageHitsAnOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount += 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Torbran, Thane of Red Fell: +2 damage",
		}},
	})
}
