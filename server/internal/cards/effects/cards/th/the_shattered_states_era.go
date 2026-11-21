package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Shattered States Era // Nameless Conqueror — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {4}{R}:
//
//	"I — Gain control of target creature until end of turn. Untap it. It
//	     gains haste until end of turn.
//	 II — Creatures you control get +1/+0 until end of turn.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Human Samurai, 3/3: "Trample, haste".
//
// Chapter I is Act of Treason's three clauses in order. Chapter II's
// "creatures you control" is the set at resolution (CR 611.2c).
//
// No simplification.
const TheShatteredStatesEraOracleID = "99473461-3c6f-4f7f-8ba2-d6c090ddb567"

func init() {
	const name = "The Shattered States Era"
	effects.Register(effects.Spec{
		OracleID:     TheShatteredStatesEraOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "gain control of target creature until end of turn, untap it, haste"),
				effects.TargetCreature("target creature"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalBattlefieldTarget(ctx)
					if target == [16]byte{} {
						return nil
					}
					if err := (effects.GainControl{Target: target, Duration: effects.DurationUntilEndOfTurn(ctx), Label: name + " — gain control until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"}, Label: name + " — haste"}.Apply(ctx)
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "creatures you control get +1/+0"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: name + " — +1/+0"}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        TheShatteredStatesEraOracleID + "#1",
		Name:            "Nameless Conqueror",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
	})
}
