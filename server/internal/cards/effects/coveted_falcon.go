package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "4c495e94-3194-47c7-8129-dd0a0f915daa",
		Name:             "Coveted Falcon",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Coveted Falcon — gain control of target permanent you own but don't control",
					Do(Instruction(func(ctx *Context) error {
						t, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						return GainControl{Target: t.ID, Duration: game.IndefiniteDuration(), Label: "Coveted Falcon — gain control"}.Apply(ctx)
					}))),
				TargetPermanent("target permanent you own but don't control", YouOwn(), Not(YouControl()))),
			Targeting(
				WhenThisIsTurnedFaceUp("Coveted Falcon — target opponent gains control of any number of target permanents you control; draw a card for each",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
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
						return DrawCards{N: gained}.Apply(ctx)
					}),
				TargetPlayer("target opponent", Opponent()).Then(
					TargetPermanent("any number of target permanents you control", YouControl()).WithCount(0, 0))),
		},
	})
}
