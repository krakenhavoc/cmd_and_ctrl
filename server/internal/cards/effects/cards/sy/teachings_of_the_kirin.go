package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teachings of the Kirin // Kirin-Touched Orochi — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493).
//
//	I   — Mill three cards. Create a 1/1 colorless Spirit creature token.
//	II  — Put a +1/+1 counter on target creature you control.
//	III — Exile this Saga, then return it to the battlefield transformed
//	      under your control.
//
// Kirin-Touched Orochi, Enchantment Creature — Snake Monk 1/1: "Whenever
// this creature attacks, choose one —
//   - Exile target creature card from a graveyard. When you do, create a
//     1/1 colorless Spirit creature token.
//   - Exile target noncreature card from a graveyard. When you do, put a
//     +1/+1 counter on target creature you control."
//
// The "when you do" is a reflexive trigger (CR 603.12) created only if
// the card really reached exile; its own target (the counter's creature)
// is chosen as it goes on the stack.
func init() {
	effects.Register(effects.Spec{
		OracleID:     TeachingsOfTheKirinOracleID,
		Name:         "Teachings of the Kirin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("Teachings of the Kirin", 1, "mill three cards, then create a 1/1 colorless Spirit"),
				effects.Do(effects.MillCards{N: 3}, effects.CreateToken{Template: effects.TokenCard("1/1 colorless Spirit"), N: 1})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel("Teachings of the Kirin", 2, "put a +1/+1 counter on target creature you control"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.PlusOneCounterOnTheTarget),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Teachings of the Kirin", 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})

	attack := effects.WheneverThisAttacks("Kirin-Touched Orochi — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	attack.Modes = effects.ChooseOne(
		effects.ModeDoing("Exile target creature card from a graveyard. When you do, create a 1/1 colorless Spirit creature token.",
			effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				return kirinOrochiExile(ctx, occ, kirinOrochiSpiritTail)
			}),
		effects.ModeDoing("Exile target noncreature card from a graveyard. When you do, put a +1/+1 counter on target creature you control.",
			effects.TargetCardInGraveyard("target noncreature card from a graveyard", effects.Noncreature()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				return kirinOrochiExile(ctx, occ, kirinOrochiCounterTail)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     TeachingsOfTheKirinOracleID + "#1",
		Name:         "Kirin-Touched Orochi",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}

// teachingsOfTheKirinOracleID is shared with the back face ("#1").
const TeachingsOfTheKirinOracleID = "cb7edb5e-3a90-41c7-ae52-dc564086bf22"

func kirinOrochiExile(ctx *effects.Context, occ int, tail game.TailRef) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return effects.ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(tail)}.Apply(ctx)
}

var kirinOrochiSpiritBody = game.ReflexiveBody("s82-neo-b18/kirin-touched-orochi-spirit",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Spirit"), N: 1}.Apply(effects.NewContext(g, item))
	}),
	effects.ConstTargets(func() *game.TargetSpec { return nil }))

var kirinOrochiCounterBody = game.ReflexiveBody("s82-neo-b18/kirin-touched-orochi-counter",
	effects.SimpleBody(effects.PlusOneCounterOnTheTarget),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you control", effects.YouControl())
	}))

var kirinOrochiSpiritTail = effects.CardTail("tail/s82-neo-b18-kirin-orochi-spirit", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return effects.WhenYouDo("Kirin-Touched Orochi — create a 1/1 colorless Spirit creature token", kirinOrochiSpiritBody).Apply(ctx)
})

var kirinOrochiCounterTail = effects.CardTail("tail/s82-neo-b18-kirin-orochi-counter", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return effects.WhenYouDo("Kirin-Touched Orochi — put a +1/+1 counter on target creature you control", kirinOrochiCounterBody).Apply(ctx)
})
