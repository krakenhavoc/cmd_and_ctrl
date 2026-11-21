package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wumpus Aberration — Creature — Eldrazi Beast {3}{G}, 6/6 (Modern Horizons 3):
//
//	"Devoid
//	 When you cast this spell, if {C} wasn't spent to cast it, target opponent
//	 may put a creature card from their hand onto the battlefield.
//	 Trample"
//
// The target opponent chooses and the creature enters under their control.
func init() {
	const label = "Wumpus Aberration — target opponent may put a creature card from their hand onto the battlefield"
	cast := effects.WhenYouCastThisSpell(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			return effects.PutFromHandOntoBattlefield{
				Player:   t.ID,
				Match:    effects.Creature(),
				Optional: true,
				Label:    "Wumpus Aberration — you may put a creature card from your hand onto the battlefield",
			}.Apply(ctx)
		}
		return nil
	})
	cast.AppliesTo = effects.GapB51NoColorlessSpent
	cast.Targets = effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:        "c116801e-9893-46e8-a967-352b5fa27e86",
		Name:            "Wumpus Aberration",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "trample"},
		ReadsManaSpent:  game.ReadsOwnPayment,
		WantsColors:     []string{"C"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}
