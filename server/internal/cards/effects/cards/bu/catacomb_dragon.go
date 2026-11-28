package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Catacomb Dragon — Creature — Dragon {4}{B}{B}, 4/4 (Mirage, S82 OLD b807):
//
//	"Flying
//	 Whenever this creature becomes blocked by a nonartifact, non-Dragon
//	 creature, that creature gets -X/-0 until end of turn, where X is half the
//	 creature's power, rounded down."
//
// One block event per blocker–attacker pair (CR 509.3b, 509.3d); the trigger
// is the attacker's side only (this creature is the one blocked), so it
// triggers once for each qualifying blocker. The blocker's type and
// subtypes are read through the layers as the block is declared, and X is
// half its power as the trigger resolves, rounded down toward negative
// infinity (CR 107.1a): a blocker with power 5 gets -2/-0, one with power 1
// gets -0/-0. A blocker that has left does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a5e9f72b-7628-4265-8b43-592a0bf678b3",
		Name:            "Catacomb Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Target != source.InstanceID {
					return false
				}
				blocker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.OldB807NonartifactNonDragon(g, blocker)
			}, "Catacomb Dragon — that creature gets -X/-0 until end of turn, X half its power rounded down",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					blocker := item.Trigger.Event.CardID
					if !effects.OnBattlefield(g, blocker) {
						return nil
					}
					x := effects.OldB807HalfRoundedDown(effects.B31CurrentPowerOf(g, blocker))
					return effects.BoostUntilEOT{Target: blocker, Power: -x, Label: "Catacomb Dragon — -X/-0"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
