package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jabari's Influence — Instant {3}{W}{W} (Mirage):
//
//	"Cast this spell only after combat.
//	 Gain control of target nonartifact, nonblack creature that attacked you
//	 this turn and put a -1/-0 counter on it."
//
// Cast only once the combat phase is over (the postcombat main phase and the end
// step). The target must be a nonartifact, nonblack creature declared as an
// attacker this turn with its defending player the caster (the attack record
// names the object, CR 400.7), checked at announce and again at resolution
// (CR 608.2b). Control is gained with no duration, so for as long as the game
// lasts (CR 611.2a), and the -1/-0 counter goes on afterwards; a creature that
// has left is neither taken nor counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8930b1ce-80ac-48b6-a368-ccff9fd01d38",
		Name:         "Jabari's Influence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			switch g.Turn.Step {
			case game.StepPostcombatMain, game.StepEnd:
				return true
			}
			return false
		},
		CastConditionLabel: "Cast this spell only after combat.",
		Targets: effects.TargetCreature("target nonartifact, nonblack creature that attacked you this turn",
			effects.Not(effects.Artifact()), effects.NonBlack(), effects.OldB804AttackedYouThisTurn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GainControl{Target: id, Duration: game.IndefiniteDuration(), Label: "Jabari's Influence — gain control"}).Apply(ctx); err != nil {
				return err
			}
			return effects.AddCounter{Target: id, Kind: "-1/-0", N: 1}.Apply(ctx)
		},
	})
}
