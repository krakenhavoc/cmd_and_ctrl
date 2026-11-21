package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jecht, Reluctant Guardian // Braska's Final Aeon — a transforming card
// (Final Fantasy, #4462).
//
// Front, Jecht, Reluctant Guardian — Legendary Creature — Human Warrior
// {3}{B}, 4/3:
//
//	"Menace
//	 Whenever Jecht deals combat damage to a player, you may exile it, then
//	 return it to the battlefield transformed under its owner's control."
//
// Back, Braska's Final Aeon — Legendary Enchantment Creature — Saga
// Nightmare, 7/7:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Jecht Beam — Each opponent discards a card and you draw a card.
//	 III — Ultimate Jecht Shot — Each opponent sacrifices two creatures of
//	 their choice.
//	 Menace"
//
// The transform is the exile-and-return verb (CR 712.14a): the Aeon is a
// new object that enters as a Saga, so it starts on chapter I.
func init() {
	const jechtOracle = "b6cebf8d-37e1-42de-b3f1-1c174ab17ac8"
	effects.Register(effects.Spec{
		OracleID:        jechtOracle,
		Name:            "Jecht, Reluctant Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Jecht, Reluctant Guardian — you may exile it, then return it to the battlefield transformed under its owner's control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Jecht, Reluctant Guardian — exile it, then return it to the battlefield transformed?",
						Then:     jechtTransform,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	jechtBeam := func(g *game.Game, item *game.StackItem) error {
		return effects.Do(
			effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
			effects.DrawCards{N: 1},
		)(g, item)
	}
	effects.Register(effects.Spec{
		OracleID:        jechtOracle + "#1",
		Name:            "Braska's Final Aeon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("Braska's Final Aeon", 1, "Jecht Beam — each opponent discards a card and you draw a card"), jechtBeam),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("Braska's Final Aeon", 2, "Jecht Beam — each opponent discards a card and you draw a card"), jechtBeam),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Braska's Final Aeon", 3, "Ultimate Jecht Shot — each opponent sacrifices two creatures of their choice"),
				effects.FinB13EachOpponentSacrificesN(2, "Ultimate Jecht Shot — choose two creatures to sacrifice")),
		},
	})
}

// jechtTransform is the "yes": exile Jecht, then return it transformed under
// its owner's control. An on-disk key.
var jechtTransform = effects.MayYes("may/jecht-reluctant-guardian-transform", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.ExileAndReturnTransformed{Target: ctx.Source()}.Apply(ctx)
})
