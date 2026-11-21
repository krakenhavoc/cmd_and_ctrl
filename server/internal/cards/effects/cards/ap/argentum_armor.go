package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argentum Armor — Artifact — Equipment for {6} (EDHREC rank 2221):
//
//	"Equipped creature gets +6/+6.
//	 Whenever equipped creature attacks, destroy target permanent.
//	 Equip {6}"
//
// Twelve mana before it does anything and worth it anyway, because
// "destroy target permanent" every combat with no restriction is the
// widest removal clause in the format — lands, commanders, whatever
// is holding the game together.
//
// THE FIRST TARGETED ATTACHMENT TRIGGER in the catalog. Targets rides
// game.TriggeredAbility.Targets, so the engine picks the target as
// the ability goes on the stack (CR 603.3d), drops the trigger with no
// prompt if the board is empty, and re-checks the choice at
// resolution (CR 608.2b) — a permanent that gains hexproof in
// response is skipped rather than erroring, through the same
// destroyChosenTargetTrigger four other catalog cards resolve with.
//
// The target clause is "permanent", unqualified, which includes your
// own — the card really is that indiscriminate, and in a four-player
// game the right target is usually not yours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0a1f39b-cda4-4925-83f8-2161f575edfb",
		Name:         "Argentum Armor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(6, 6)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			},
			Targets: effects.TargetPermanent("target permanent", effects.Permanent()),
			Key:     "Argentum Armor — destroy target permanent",
			Effect:  effects.DestroyChosenPermanent,
		}},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{6}"),
		},
	})
}
