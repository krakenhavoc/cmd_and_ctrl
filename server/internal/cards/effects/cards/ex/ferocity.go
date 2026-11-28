package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ferocity — Enchantment — Aura {1}{G} (Mercadian Masques, S82 OLD batch 685, #5619):
//
//	"Enchant creature
//	 Whenever enchanted creature blocks or becomes blocked, you may put a +1/+1 counter on it."
//
// One trigger over both events: EventBlock (CardID is the blocker) and
// EventBecomesBlocked (CardID is the attacker, once however many blockers,
// CR 509.1h). A creature blocks one attacker, so each fires at most once for
// the host. The counter goes on the creature the event names, if it is still on
// the battlefield, and is optional at resolution (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b1a0f35-3126-4b88-90aa-c56e67850a9d",
		Name:         "Ferocity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source.IsAttachedTo(ev.CardID)
				},
				"Ferocity — you may put a +1/+1 counter on it",
				effects.PutPlusOneCounterOnEventCreature),
				"Ferocity — put a +1/+1 counter on it?"),
		},
	})
}
