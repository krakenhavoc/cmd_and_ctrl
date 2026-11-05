package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Galvanizing Tenor — Creature — Klingon Performer {3}{R}, 4/3 (Star Trek
// Commander, #4974):
//
//	"Menace
//	 Battlefront — Whenever this creature attacks, goad up to X target
//	 creatures, where X is the number of opponents you attacked this
//	 turn. (Until your next turn, those creatures attack each combat if
//	 able and attack a player other than you if able.)"
//
// Battlefront is an ability word. X is counted as the trigger goes on the
// stack, when its targets are chosen, from this turn's attack records
// (trekB37OpponentsAttackedThisTurn): the attack that triggered it is
// already recorded, so attacking two opponents is up to two targets. The
// target group is built then, the way Batroc the Leaper's "up to X" is; an
// X of zero (the Tenor attacked only a planeswalker or a battle) asks for
// nothing. Any creature may be chosen, yours included. The goad is the
// ordinary one (CR 701.15a): until your next turn, for each target still
// legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	attack := WheneverThisAttacks("Galvanizing Tenor — goad up to X target creatures", galvanizingTenorGoad)
	attack.TargetsFrom = func(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		x := trekB37OpponentsAttackedThisTurn(g, source.Controller)
		if x <= 0 {
			return nil
		}
		return TargetCreature("up to X target creatures").WithCount(0, x)
	}
	Register(Spec{
		OracleID:        "d974fee0-7d20-465c-a7e3-bc361e241cd4",
		Name:            "Galvanizing Tenor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}

// galvanizingTenorGoad goads every target still legal as the ability
// resolves.
func galvanizingTenorGoad(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return goadAndScheduleClear(ctx, ids)
}
