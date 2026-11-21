package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coveted Falcon — Artifact Creature — Bird {1}{U}{U} (Murders at Karlov
// Manor, #4469):
//
//	"Flying
//	 Whenever this creature attacks, gain control of target permanent you
//	 own but don't control.
//	 Disguise {1}{U}
//	 When this creature is turned face up, target opponent gains control of
//	 any number of target permanents you control. Draw a card for each one
//	 they gained control of this way."
//
// Both control changes last indefinitely (CR 611.2a: no duration is stated),
// each pinned to the object it took (CR 400.7), so a permanent that leaves
// and returns comes back to its owner. "Gained control of this way" counts
// the permanents the engine actually handed over: one that left the
// battlefield, or an opponent who is no longer a legal target, is not
// counted (CR 608.2b). A token you gave away is still its creator's: "you
// own" is read through its owner (CR 111.2), so the attack trigger can take
// it back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4c495e94-3194-47c7-8129-dd0a0f915daa",
		Name:             "Coveted Falcon",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Coveted Falcon — gain control of target permanent you own but don't control",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						t, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						return effects.GainControl{Target: t.ID, Duration: game.IndefiniteDuration(), Label: "Coveted Falcon — gain control"}.Apply(ctx)
					}))),
				effects.TargetPermanent("target permanent you own but don't control", effects.YouOwn(), effects.Not(effects.YouControl()))),
			effects.Targeting(
				effects.WhenThisIsTurnedFaceUp("Coveted Falcon — target opponent gains control of any number of target permanents you control; draw a card for each",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						opp, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						gained := 0
						for _, t := range ctx.ClauseTargets(1) {
							if !ctx.IsTargetLegal(t) {
								continue
							}
							if g.GainControlForEffect(ctx.Source(), t.ID, opp.ID, game.IndefiniteDuration(), "Coveted Falcon — gain control") {
								gained++
							}
						}
						if gained == 0 {
							return nil
						}
						return effects.DrawCards{N: gained}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent()).Then(
					effects.TargetPermanent("any number of target permanents you control", effects.YouControl()).WithCount(0, 0))),
		},
	})
}
