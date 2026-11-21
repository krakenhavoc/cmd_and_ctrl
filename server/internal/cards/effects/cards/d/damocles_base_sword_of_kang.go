package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Damocles Base, Sword of Kang — Legendary Artifact — Vehicle {4}{B}, 5/5:
//
//	"Flying, deathtouch
//	 Whenever Damocles Base deals combat damage to a player, that player
//	 faces a villainous choice — They sacrifice a nontoken creature of
//	 their choice, or they lose 2 life and you draw two cards.
//	 Crew 3"
//
// A villainous choice (CR 701.55, #3140), faced by the player it hit.
// Both options are always on offer (CR 701.55b): a player with no
// nontoken creature may pick the sacrifice and lose nothing, which is
// the printed rule and not a gap. The damaged player is read off the
// triggering damage event. Crew 3 is the ordinary CR 702.122 ability.
//
// No simplifications.
var DamoclesBaseChoice = effects.NewVillainousChoice("damocles-base",
	"Damocles Base — sacrifice a nontoken creature, or lose 2 life and the controller draws two cards?",
	effects.VillainousOption{
		Label:    "Sacrifice a nontoken creature of your choice",
		Harm:     2,
		Possible: effects.VillainousCanSacrifice(damoclesNontokenCreature()),
		Body:     effects.VillainousSacrifice("a nontoken creature", damoclesNontokenCreature()),
	},
	effects.VillainousOption{
		Label:    "Lose 2 life; the controller of Damocles Base draws two cards",
		LifeCost: 2,
		Harm:     1,
		Body: func(ctx *effects.Context, f effects.Facing) error {
			ctx.RecordStep(damoclesBaseRecord, []uuid.UUID{f.Player})
			return damoclesBaseLoseAndDraw.Run(ctx)
		},
	})

// damoclesBaseRecord is the player who lost the life.
const damoclesBaseRecord = "damocles-base/facing"

// damoclesBaseLoseAndDraw is the life loss and then the draw: the draw waits
// for a prompt the loss asks (a CR 616 order; CR 608.2c, #4155).
var damoclesBaseLoseAndDraw = effects.RegisterThenSteps("then-steps/damocles-base-lose-and-draw",
	effects.Instruction(func(ctx *effects.Context) error {
		r := ctx.StepRecord(damoclesBaseRecord)
		if len(r.IDs) == 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.IDs[0], -2)
	}),
	effects.DrawCards{N: 2},
)

// damoclesNontokenCreature is "a nontoken creature".
func damoclesNontokenCreature() effects.CardPredicate {
	return effects.And(effects.Creature(), effects.Not(effects.IsTokenPredicate()))
}

func init() {
	effects.Register(effects.Spec{
		OracleID:        "544c9ad2-86f9-483e-a53c-f64c8602c393",
		Name:            "Damocles Base, Sword of Kang",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Damocles Base — that player faces a villainous choice",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.FaceVillainousChoice{
						Players: []uuid.UUID{item.Trigger.Event.Target},
						Choice:  DamoclesBaseChoice,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Crew 3",
				Cost:    effects.CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Damocles Base, Sword of Kang"),
			},
		},
	})
}
