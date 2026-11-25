package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reverend Mother Jessica — Legendary Creature — Human Noble Advisor {2}{U},
// 2/2:
//
//	"Prowess, prowess
//	 Whenever Reverend Mother Jessica attacks, for each opponent, tap up to one
//	 target creature that player controls with lesser power and goad it."
//
// Two instances of prowess (CR 702.108b: each triggers separately). The attack
// trigger has one "up to one" clause per possible opponent
// (PerOpponentClauses' shape), each bound to that opponent by turn order and
// narrowed to creatures with power less than Jessica's, judged as the trigger is
// put on the stack and again on resolution against her power as it then is, or
// as it last was (CR 608.2h, RelativeToSource). Each pick still legal is tapped
// and goaded (CR 701.15a): until your next turn it attacks each combat if able
// and attacks a player other than you if able. A creature already tapped is
// goaded all the same.
//
// No simplification.
func init() {
	trigger := effects.WheneverThisAttacks("Reverend Mother Jessica — for each opponent, tap up to one target creature that player controls with lesser power and goad it",
		reverendMotherJessicaTapAndGoad)
	var clauses []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		clauses = append(clauses, effects.RelativeToSource(
			effects.TargetPermanent("up to one target creature controlled by the "+effects.NthOpponentWords[n]+" opponent after you in turn order with lesser power",
				effects.Creature(), effects.ControlledByNthOpponent(n)).WithCount(0, 1),
			effects.LesserPower()))
	}
	trigger.Targets = effects.Clauses(clauses[0], clauses[1:]...)
	effects.Register(effects.Spec{
		OracleID:        "eb947d87-fa06-438c-93d2-7d681468b2f1",
		Name:            "Reverend Mother Jessica",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess, game.KeywordProwess},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}

func reverendMotherJessicaTapAndGoad(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var picked []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !effects.OnBattlefield(g, t.ID) {
			continue
		}
		if err := g.TapTargetForEffect(t.ID); err != nil {
			return err
		}
		picked = append(picked, t.ID)
	}
	if len(picked) == 0 {
		return nil
	}
	return effects.GoadAndScheduleClear(ctx, picked)
}
