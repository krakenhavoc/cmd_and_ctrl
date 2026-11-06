package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karlov Watchdog — Creature — Dog {3}{W}, 3/2 (Murders at Karlov Manor,
// #5241):
//
//	"Vigilance
//	 Permanents your opponents control can't be turned face up during
//	 your turn.
//	 Whenever you attack with three or more creatures, creatures you
//	 control get +1/+1 until end of turn."
//
// The middle line is a TurnFaceUpRestriction (game/turn_face_up_gate.go):
// read live from the battlefield, it closes the turn-face-up special
// action for a permanent an opponent controls while it is the Watchdog's
// controller's turn, whether that action would pay a disguise or morph
// cost or a manifested or cloaked creature card's mana cost (the
// ruling), and an effect that would turn one face up does nothing to it
// (CR 101.2). It is checked when the action is taken, not stored, so the
// Watchdog leaving lifts it at once.
//
// The attack trigger fires once per combat, on the declaration that
// gives you three or more attacking creatures, so an additional combat
// can trigger it again. "Creatures you control" is locked in as it
// resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73f44500-fe83-40fa-b6f2-ed5673d0edcc",
		Name:            "Karlov Watchdog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		TurnFaceUpRestrictions: []game.TurnFaceUpRestriction{
			OpponentsPermanentsCantTurnFaceUpDuringYourTurn("Permanents your opponents control can't be turned face up during your turn."),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, karlovWatchdogAttackedWithThree, karlovWatchdogAttackLabel, Do(BoostUntilEOT{
				Match:     And(Creature(), YouControl()),
				Power:     1,
				Toughness: 1,
				Label:     "Karlov Watchdog — +1/+1",
			}))),
		},
	})
}

// karlovWatchdogAttackLabel is the attack trigger's stack label, named
// because the once-per-combat dedup matches on it.
const karlovWatchdogAttackLabel = "Karlov Watchdog — you attacked with three or more creatures: creatures you control get +1/+1"

// OpponentsPermanentsCantTurnFaceUpDuringYourTurn is "Permanents your
// opponents control can't be turned face up during your turn": a
// face-down permanent controlled by anyone but the source's controller,
// while the source's controller is the active player.
func OpponentsPermanentsCantTurnFaceUpDuringYourTurn(label string) game.TurnFaceUpRestriction {
	return game.TurnFaceUpRestriction{
		Label: label,
		Forbids: func(q game.TurnFaceUpQuery) bool {
			you := q.Source.Controller
			return you != uuid.Nil && q.Card.Controller != you && IsYourTurn(q.Game, you)
		},
	}
}

// karlovWatchdogAttackedWithThree is "whenever you attack with three or
// more creatures": the source's controller declared attackers and now
// has at least three attacking creatures, and the ability has not
// triggered yet in this combat. Combat phases are counted by the turn's phase
// ordinal (Turn.PhaseOrdinal: the first combat is 1, an added one 2), so
// an additional combat can trigger it again, and a creature declared
// late in the same combat cannot. A file restored without the ordinal
// reads as the first combat.
func karlovWatchdogAttackedWithThree(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) || len(b13AttackingCreaturesYouControl(g, ev.Actor)) < 3 {
		return false
	}
	return g.TriggeredThisTurn(source.InstanceID, karlovWatchdogAttackLabel) < max(g.Turn.PhaseOrdinal, 1)
}
