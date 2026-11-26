package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vigorspore Wurm — Creature — Wurm {5}{G}, 6/4 (Guilds of Ravnica, #5525):
//
//	"Undergrowth — When this creature enters, target creature gains vigilance
//	 and gets +X/+X until end of turn, where X is the number of creature cards
//	 in your graveyard.
//	 This creature can't be blocked by more than one creature."
//
// X is counted as the ability resolves (CR 608.2h) and locked for the turn;
// the target is any creature, yours or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be843255-7408-45b9-b0ca-c79342daa353",
		Name:         "Vigorspore Wurm",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Vigorspore Wurm — target creature gains vigilance and gets +X/+X until end of turn, where X is the number of creature cards in your graveyard",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						x := effects.B11CreatureCardsInGraveyard(g, ctx.Controller())
						if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"},
							Label: "Vigorspore Wurm — vigilance until end of turn"}).Apply(ctx); err != nil {
							return err
						}
						return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x,
							Label: "Vigorspore Wurm — +X/+X until end of turn"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
