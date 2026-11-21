package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alluring Suitor // Deadly Dancer — {2}{R} Creature — Vampire 2/3 //
// Creature — Vampire 3/3 (Innistad: Crimson Vow, #5495):
//
//	Front: "When you attack with exactly two creatures, transform this
//	        creature."
//	Back:  "Trample
//	        When this creature transforms into Deadly Dancer, add {R}{R}.
//	        Until end of turn, you don't lose this mana as steps and phases
//	        end.
//	        {R}{R}: This creature and another target creature each get
//	        +1/+0 until end of turn."
//
// The front trigger is Love on the Battlefield's "exactly two" attack
// reader, one trigger per declaration (CR 603.2c). The Suitor need not be
// one of the attackers. It transforms in place (CR 701.27f ignores the
// instruction if it already turned over since the trigger went on the
// stack). The mana is Savage Ventmaw's keep-until-end-of-turn mana. The
// pump is one resolution on the Dancer and the chosen creature: if the
// target is illegal the ability does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	const oracle = "e98e8129-3e0e-45cc-b00c-1387cc592023"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Alluring Suitor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, effects.LoveOnTheBattlefieldExactlyTwo,
				"Alluring Suitor — transform this creature", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Deadly Dancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTransform, effects.ThisTransformedInto(1),
				"Deadly Dancer — add {R}{R}; you don't lose this mana as steps and phases end",
				effects.Do(effects.AddMana{Produced: "{R}{R}", Riders: []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()}})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{R}: This creature and another target creature each get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}{R}"),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Deadly Dancer — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Deadly Dancer — +1/+0"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
