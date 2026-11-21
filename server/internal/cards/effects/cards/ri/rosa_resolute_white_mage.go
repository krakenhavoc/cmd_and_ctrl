package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rosa, Resolute White Mage — Legendary Creature — Human Noble Cleric
// {3}{W}, 2/3 (Final Fantasy, #4462):
//
//	"Reach
//	 At the beginning of combat on your turn, put a +1/+1 counter on
//	 target creature you control. It gains lifelink until end of turn."
//
// Luminarch Aspirant's targeted beginning-of-combat trigger with a
// lifelink grant on the same creature. Rosa can target herself. A target
// that has left in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c4fdd20-0396-40b7-8687-43b1035b6956",
		Name:            "Rosa, Resolute White Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Rosa, Resolute White Mage — put a +1/+1 counter on target creature you control; it gains lifelink",
					rosaCounterAndLifelink),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}

// rosaCounterAndLifelink puts the counter on Rosa's target, then gives it
// lifelink until end of turn.
func rosaCounterAndLifelink(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"},
		Label: "Rosa, Resolute White Mage — lifelink until end of turn"}.Apply(ctx)
}
