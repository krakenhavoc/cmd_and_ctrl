package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Immard, the Stormcleaver — Legendary Creature — Human Soldier {1}{U}{R}{W}, 4/4:
//
//	"Whenever Immard, the Stormcleaver enters or attacks, put a charge counter
//	 on it or remove one from it. When you remove a counter this way, choose
//	 one —
//	 • Immard deals 4 damage to any target.
//	 • Immard gains lifelink and indestructible until end of turn."
//
// One trigger for both events. The controller picks put or remove as it
// resolves; "remove" is only possible with a charge counter on Immard (CR
// 122.1), and only a counter actually removed starts the reflexive trigger
// (CR 603.12), which chooses its mode and any target as it goes on the stack.
//
// No simplification.
func init() {
	t := effects.WhenThisEntersOrAttacks("Immard, the Stormcleaver — put a charge counter on it or remove one from it", immardPutOrRemove)
	effects.Register(effects.Spec{
		OracleID:     "104f65d3-bedb-4dc5-a5dd-8ab9fe796af4",
		Name:         "Immard, the Stormcleaver",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

func immardPutOrRemove(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return effects.PickOption{
		Question: "Immard, the Stormcleaver — put a charge counter on it or remove one from it?",
		Options: []game.ChoiceOption{
			{Label: "Put a charge counter on it"},
			{Label: "Remove a charge counter from it"},
		},
		ThenKey: immardAnswered,
		Carry:   []uuid.UUID{item.SourceCardID},
	}.Apply(ctx)
}

// immardAnswered is the choice, keyed. An on-disk key.
var immardAnswered = effects.OptionPickThen("option-pick/immard-the-stormcleaver-charge", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	id := r.Carry[0]
	if r.Index != 1 {
		return effects.AddCounter{Target: id, Kind: game.CounterCharge, N: 1}.Apply(ctx)
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.Counters[game.CounterCharge] <= 0 {
		return nil
	}
	if err := (effects.AddCounter{Target: id, Kind: game.CounterCharge, N: -1}).Apply(ctx); err != nil {
		return err
	}
	return effects.WhenYouDo("Immard, the Stormcleaver — choose one", immardModes).Apply(ctx)
})

var immardModes = game.ReflexiveModalBody("immard-the-stormcleaver/modes", func(_ uuid.UUID, _ game.EffectParams) *game.ModeSpec {
	return effects.ChooseOne(
		effects.ModeDoing("Immard deals 4 damage to any target.", effects.TargetAny(), func(item *game.StackItem, ctx *effects.Context, occ int) error {
			t, ok := effects.ModeTarget(ctx, occ)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
		}),
		effects.ModeDoing("Immard gains lifelink and indestructible until end of turn.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.GrantKeywordUntilEOT{
				Target:   item.SourceCardID,
				Keywords: []string{"lifelink", "indestructible"},
				Label:    "Immard, the Stormcleaver — lifelink and indestructible until end of turn",
			}.Apply(ctx)
		}),
	)
})
