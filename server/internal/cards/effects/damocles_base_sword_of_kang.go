package effects

import (
	"github.com/google/uuid"

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
var damoclesBaseChoice = NewVillainousChoice("damocles-base",
	"Damocles Base — sacrifice a nontoken creature, or lose 2 life and the controller draws two cards?",
	VillainousOption{
		Label:    "Sacrifice a nontoken creature of your choice",
		Harm:     2,
		Possible: villainousCanSacrifice(damoclesNontokenCreature()),
		Body:     villainousSacrifice("a nontoken creature", damoclesNontokenCreature()),
	},
	VillainousOption{
		Label:    "Lose 2 life; the controller of Damocles Base draws two cards",
		LifeCost: 2,
		Harm:     1,
		Body: func(ctx *Context, f Facing) error {
			ctx.RecordStep(damoclesBaseRecord, []uuid.UUID{f.Player})
			return damoclesBaseLoseAndDraw.Run(ctx)
		},
	})

// damoclesBaseRecord is the player who lost the life.
const damoclesBaseRecord = "damocles-base/facing"

// damoclesBaseLoseAndDraw is the life loss and then the draw: the draw waits
// for a prompt the loss asks (a CR 616 order; CR 608.2c, #4155).
var damoclesBaseLoseAndDraw = RegisterThenSteps("then-steps/damocles-base-lose-and-draw",
	Instruction(func(ctx *Context) error {
		r := ctx.StepRecord(damoclesBaseRecord)
		if len(r.IDs) == 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.IDs[0], -2)
	}),
	DrawCards{N: 2},
)

// damoclesNontokenCreature is "a nontoken creature".
func damoclesNontokenCreature() CardPredicate {
	return And(Creature(), Not(IsTokenPredicate()))
}

func init() {
	Register(Spec{
		OracleID:        "544c9ad2-86f9-483e-a53c-f64c8602c393",
		Name:            "Damocles Base, Sword of Kang",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Damocles Base — that player faces a villainous choice",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return FaceVillainousChoice{
						Players: []uuid.UUID{item.Trigger.Event.Target},
						Choice:  damoclesBaseChoice,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Crew 3",
				Cost:    CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Damocles Base, Sword of Kang"),
			},
		},
	})
}
