package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f11ec1c9-347d-4d79-aaa9-5919db5f9044",
		Name:         "Clandestine Meddler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Clandestine Meddler — suspect up to one other target creature you control", SuspectEachLegalTarget),
				Another(UpToOneTargetCreature("up to one other target creature you control", YouControl())),
			),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && attacker.Suspected
			}, "Clandestine Meddler — surveil 1",
				Do(Surveil{N: 1}))),
		},
	})
}
