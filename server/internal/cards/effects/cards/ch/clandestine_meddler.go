package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clandestine Meddler — Creature — Vampire Rogue {2}{B}:
//
//	"When this creature enters, suspect up to one other target creature
//	 you control. (A suspected creature has menace and can't block.)
//	 Whenever one or more suspected creatures you control attack,
//	 surveil 1."
//
// The second ability is a "one or more" batch: the engine emits one
// attack event per creature, and OncePerBatch fires on the first
// suspected attacker and declines the rest of the same declaration, so a
// three-creature alpha strike surveils once. An unsuspected attacker
// does not use the batch up, because the guard runs only for events the
// condition accepted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f11ec1c9-347d-4d79-aaa9-5919db5f9044",
		Name:         "Clandestine Meddler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Clandestine Meddler — suspect up to one other target creature you control", effects.SuspectEachLegalTarget),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature you control", effects.YouControl())),
			),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && attacker.Suspected
			}, "Clandestine Meddler — surveil 1",
				effects.Do(effects.Surveil{N: 1}))),
		},
	})
}
