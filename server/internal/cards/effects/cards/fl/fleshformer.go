package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fleshformer — Creature — Human Wizard {2}{B}, 2/2 (Conflux):
//
//	"{W}{U}{B}{R}{G}: This creature gets +2/+2 and gains fear until end of
//	 turn. Target creature gets -2/-2 until end of turn. Activate only
//	 during your turn."
//
// One activated ability with one target. The target is chosen as it is
// activated; on resolution Fleshformer pumps itself only while it is still
// the permanent that activated (CR 400.7), and the -2/-2 lands on the target
// if it is still legal (CR 608.2b). Fear is a layer-6 grant the engine reads
// when blockers are declared. Activation is limited to your own turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67b8b67d-fda3-4a62-9bfb-51a171638502",
		Name:         "Fleshformer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{W}{U}{B}{R}{G}: This creature gets +2/+2 and gains fear until end of turn. Target creature gets -2/-2 until end of turn. Activate only during your turn.",
			Purpose:   fleshformerPurpose(),
			Cost:      effects.ManaCost("{W}{U}{B}{R}{G}"),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.SourceIsStillThisPermanent(g, item) {
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: "Fleshformer — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"fear"}, Label: "Fleshformer — fear"}).Apply(ctx); err != nil {
						return err
					}
				}
				if id := effects.FirstLegalBattlefieldTarget(ctx); id != uuid.Nil {
					return effects.BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Fleshformer — -2/-2"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// fleshformerPurpose declares the -2/-2 on the target (#4302) beside the pump
// and the evasion Fleshformer gives itself.
func fleshformerPurpose() game.Purpose {
	p := effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))
	p.Answers = game.AnswerPump | game.AnswerCombatGrant
	return p
}
