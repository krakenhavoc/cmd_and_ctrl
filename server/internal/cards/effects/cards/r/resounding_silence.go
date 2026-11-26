package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resounding Silence — Instant {3}{W} (Shards of Alara):
//
//	"Exile target attacking creature.
//	 Cycling {5}{G}{W}{U} ({5}{G}{W}{U}, Discard this card: Draw a card.)
//	 When you cycle this card, exile up to two target attacking creatures."
//
// The spell exiles one attacking creature, legal only while it is attacking as
// the spell resolves (CR 608.2b). The cycle trigger is Choking Tethers'
// graveyard trigger (CR 702.29c) with an "up to two" target clause; each target
// is exiled if it is still an attacking creature, and with none chosen the
// trigger is removed from the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6eefa487-1542-4d24-80bd-4893d9d908c8",
		Name:         "Resounding Silence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileEveryLegalCardTarget(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{5}{G}{W}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.OldB508CycleTrigger("Resounding Silence — cycled: exile up to two target attacking creatures",
				effects.TargetCreature("up to two target attacking creatures", effects.AttackingCreature()).WithCount(0, 2),
				effects.ExileEveryLegalCardTarget),
		},
	})
}
