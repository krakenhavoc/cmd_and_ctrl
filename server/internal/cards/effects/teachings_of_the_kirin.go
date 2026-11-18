package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     teachingsOfTheKirinOracleID,
		Name:         "Teachings of the Kirin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("Teachings of the Kirin", 1, "mill three cards, then create a 1/1 colorless Spirit"),
				Do(MillCards{N: 3}, CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: 1})),
			ChapterTriggerTargeting(2, SagaChapterLabel("Teachings of the Kirin", 2, "put a +1/+1 counter on target creature you control"),
				TargetCreature("target creature you control", YouControl()),
				PlusOneCounterOnTheTarget),
			ChapterTrigger(3, SagaChapterLabel("Teachings of the Kirin", 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})

	attack := WheneverThisAttacks("Kirin-Touched Orochi — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	attack.Modes = ChooseOne(
		ModeDoing("Exile target creature card from a graveyard. When you do, create a 1/1 colorless Spirit creature token.",
			TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				return kirinOrochiExile(ctx, occ, kirinOrochiSpiritTail)
			}),
		ModeDoing("Exile target noncreature card from a graveyard. When you do, put a +1/+1 counter on target creature you control.",
			TargetCardInGraveyard("target noncreature card from a graveyard", Noncreature()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				return kirinOrochiExile(ctx, occ, kirinOrochiCounterTail)
			}),
	)
	Register(Spec{
		OracleID:     teachingsOfTheKirinOracleID + "#1",
		Name:         "Kirin-Touched Orochi",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}

// teachingsOfTheKirinOracleID is shared with the back face ("#1").
const teachingsOfTheKirinOracleID = "cb7edb5e-3a90-41c7-ae52-dc564086bf22"

func kirinOrochiExile(ctx *Context, occ int, tail game.TailRef) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(tail)}.Apply(ctx)
}

var kirinOrochiSpiritBody = game.ReflexiveBody("s82-neo-b18/kirin-touched-orochi-spirit",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: 1}.Apply(NewContext(g, item))
	}),
	constTargets(func() *game.TargetSpec { return nil }))

var kirinOrochiCounterBody = game.ReflexiveBody("s82-neo-b18/kirin-touched-orochi-counter",
	simpleBody(PlusOneCounterOnTheTarget),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target creature you control", YouControl())
	}))

var kirinOrochiSpiritTail = CardTail("tail/s82-neo-b18-kirin-orochi-spirit", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return WhenYouDo("Kirin-Touched Orochi — create a 1/1 colorless Spirit creature token", kirinOrochiSpiritBody).Apply(ctx)
})

var kirinOrochiCounterTail = CardTail("tail/s82-neo-b18-kirin-orochi-counter", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return WhenYouDo("Kirin-Touched Orochi — put a +1/+1 counter on target creature you control", kirinOrochiCounterBody).Apply(ctx)
})
