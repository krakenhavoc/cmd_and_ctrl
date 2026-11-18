package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const theShatteredStatesEraOracleID = "99473461-3c6f-4f7f-8ba2-d6c090ddb567"

func init() {
	const name = "The Shattered States Era"
	Register(Spec{
		OracleID:     theShatteredStatesEraOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "gain control of target creature until end of turn, untap it, haste"),
				TargetCreature("target creature"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := FirstLegalBattlefieldTarget(ctx)
					if target == [16]byte{} {
						return nil
					}
					if err := (GainControl{Target: target, Duration: DurationUntilEndOfTurn(ctx), Label: name + " — gain control until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"}, Label: name + " — haste"}.Apply(ctx)
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "creatures you control get +1/+0"),
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: name + " — +1/+0"}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:        theShatteredStatesEraOracleID + "#1",
		Name:            "Nameless Conqueror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
	})
}
