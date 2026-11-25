package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundercloud Shaman — Creature — Giant Shaman {3}{R}{R}, 4/4 (Commander
// 2015, S82 OLD #5546):
//
//	"When this creature enters, it deals damage equal to the number of Giants
//	 you control to each non-Giant creature."
//
// The Giants are counted once, as the trigger resolves (CR 608.2h), the
// Shaman itself among them, and that number is dealt to every creature that
// isn't a Giant, yours and theirs, in one damage instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb812ea7-3301-4774-86a2-64c4ed6766fc",
		Name:         "Thundercloud Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thundercloud Shaman — damage equal to the number of Giants you control to each non-Giant creature",
				effects.Do(effects.Instruction(thundercloudShamanBlast))),
		},
	})
}

func thundercloudShamanBlast(ctx *effects.Context) error {
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.HasSubtypeWord("Giant") {
			n++
		}
	}
	if n <= 0 {
		return nil
	}
	return effects.DamageEachMatching(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() && !c.HasSubtypeWord("Giant") }, n)
}
