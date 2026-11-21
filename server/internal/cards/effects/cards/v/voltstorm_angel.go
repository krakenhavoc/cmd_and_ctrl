package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Voltstorm Angel — Creature — Angel {3}{W}{W}, 4/4 (Modern Horizons 3):
//
//	"Flying
//	 When this creature enters, you get {E}{E}{E} (three energy counters).
//	 At the beginning of combat on your turn, you may pay {E}{E}. When you
//	 do, choose one —
//	 • This creature gains vigilance and lifelink until end of turn.
//	 • Other creatures you control get +1/+1 until end of turn."
//
// The entry is Aether Chaser's. The combat trigger is Conduit Goblin's step
// trigger with the energy paid as it resolves (ADR 0129 §3), and "when you
// do" is a modal reflexive trigger (Hylda of the Icy Crown's shape, CR 603.12,
// 700.2b): the mode is chosen as it goes on the stack, so it can be answered
// before it resolves. The first bullet is the Angel's as it was (a source
// that has left gets nothing, CR 400.7), the second reaches the other
// creatures you control as the trigger resolves (CR 611.2c).
//
// No simplification.
func init() {
	combat := effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
		"Voltstorm Angel — you may pay {E}{E}; when you do, choose one",
		effects.MayPayEnergyThen("Voltstorm Angel", 2, "choose a mode", voltstormAngelPaid))
	effects.Register(effects.Spec{
		OracleID:        "836609f7-48b8-47b0-9af6-94150b048251",
		Name:            "Voltstorm Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Voltstorm Angel", 3),
			combat,
		},
	})
}

// voltstormAngelPaid raises the reflexive modal trigger once the energy is
// paid, keyed. An on-disk key.
var voltstormAngelPaid = effects.EnergyPaid("pay-unless/gap-b55-voltstorm-angel", func(g *game.Game, item *game.StackItem) error {
	return effects.ReflexiveTrigger{
		Label: "Voltstorm Angel — choose one",
		Body:  voltstormAngelBody,
	}.Apply(effects.NewContext(g, item))
})

var voltstormAngelBody = game.ReflexiveModalBody("gap-b55-voltstorm-angel/modes", func(sourceID uuid.UUID, _ game.EffectParams) *game.ModeSpec {
	return effects.ChooseOne(
		effects.ModeDoing("This creature gains vigilance and lifelink until end of turn.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   ctx.Source(),
					Keywords: []string{"vigilance", "lifelink"},
					Label:    "Voltstorm Angel — vigilance and lifelink until end of turn",
				}.Apply(ctx)
			}),
		effects.ModeDoing("Other creatures you control get +1/+1 until end of turn.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(sourceID)),
					Power:     1,
					Toughness: 1,
					Label:     "Voltstorm Angel — other creatures get +1/+1",
				}.Apply(ctx)
			}),
	)
})
