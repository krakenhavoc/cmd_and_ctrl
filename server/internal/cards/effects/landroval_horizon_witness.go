package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Landroval, Horizon Witness — Legendary Creature — Bird Noble {4}{W}, 3/4
// (The Hobbit Eternal, HOC):
//
//	"Flying
//	 Whenever two or more creatures you control attack a player, target
//	 attacking creature without flying gains flying until end of turn."
//
// Horn of the Mark's trigger with a target (#4454). It is one trigger per
// defending player, not one per attacker: EventAttack fires per creature,
// so the ability is OncePerBatchPerPlayer, and the COUNT is read off the
// board at the lock-in (b41AttackersAgainst), where Landroval herself
// counts if she attacks. The pick is made as the trigger goes on the stack,
// after every attacker is declared, and may be any attacking creature
// without flying, an opponent's included, as printed. Oteclan Levitator's
// "gains flying" is firstTargetGainsUntilEndOfTurn, with no pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "467c9fe6-d1d3-4242-bc0c-e522b4f21d56",
		Name:            "Landroval, Horizon Witness",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OncePerBatchPerPlayer(On(game.EventAttack, twoOrMoreOfYourCreaturesAttackThatPlayer,
					"Landroval, Horizon Witness — target attacking creature without flying gains flying until end of turn",
					firstTargetGainsUntilEndOfTurn(0, "Landroval, Horizon Witness — flying until end of turn", "flying"))),
				TargetCreature("target attacking creature without flying", AttackingCreature(), WithoutKeyword("flying"))),
		},
	})
}

// twoOrMoreOfYourCreaturesAttackThatPlayer is the condition of "Whenever
// two or more creatures you control attack a player": the declaring player
// is the source's controller and at least two of their creatures, the
// source among them if it attacks, are attacking the defender this
// declaration names.
func twoOrMoreOfYourCreaturesAttackThatPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	defender := b17DefendingPlayer(g, ev)
	if defender == uuid.Nil {
		return false
	}
	return b41AttackersAgainst(g, source.Controller, defender) >= 2
}
