package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ancient Brass Dragon — Creature — Elder Dragon {5}{B}{B}, 7/6 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, roll a d20. When
//	 you do, put any number of target creature cards with total mana value X or
//	 less from graveyards onto the battlefield under your control, where X is
//	 the result."
//
// The roll's natural result is X (CR 706.3a). "When you do" is a reflexive
// trigger (CR 603.12) created by the roll's tail, so its targets are chosen as
// it goes on the stack, after the roll: any number of creature cards in any
// graveyards whose mana values add up to X or less (a target set rule, whose
// limit rides the trigger as its Params.Amount). They enter together, under
// your control, and a target that has left its graveyard is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d79fb4e-2ca0-404a-9663-1f5abce6e733",
		Name:            "Ancient Brass Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Ancient Brass Dragon — roll a d20; when you do, put creature cards with total mana value X or less onto the battlefield",
				effects.Do(effects.RollDice{Sides: 20, N: 1, Then: ancientBrassDragonRolled})),
		},
	})
}

// ancientBrassDragonRolled is the roll's tail: it creates the reflexive
// trigger carrying X. An on-disk key: never renamed, never reused.
var ancientBrassDragonRolled = effects.DiceTail("dice/ancient-brass-dragon-rolled", func(ctx *effects.Context, rolls []int) error {
	return effects.ReflexiveTrigger{
		Label:  "Ancient Brass Dragon — put any number of target creature cards with total mana value X or less from graveyards onto the battlefield under your control",
		Body:   ancientBrassDragonReanimate,
		Params: game.EffectParams{Amount: rolls[0]},
	}.Apply(ctx)
})

// ancientBrassDragonReanimate is the reflexive trigger's body and its target
// clause, a pure function of the source and the roll. An on-disk key.
var ancientBrassDragonReanimate = game.ReflexiveBody("ancient-brass-dragon/reanimate",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				ids = append(ids, t.ID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		return effects.ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
	}),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCardInGraveyard("any number of target creature cards from graveyards with total mana value X or less", effects.Creature()).
			WithCount(0, 0).WithTotal(effects.TotalManaValueAtMost(p.Amount))
	})
