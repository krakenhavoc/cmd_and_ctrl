package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krovod Haunch — Artifact — Food Equipment {W} (Murders at Karlov
// Manor, #4469):
//
//	"Equipped creature gets +2/+0.
//	 {2}, {T}, Sacrifice this Equipment: You gain 3 life.
//	 When this Equipment is put into a graveyard from the battlefield,
//	 you may pay {1}{W}. If you do, create two 1/1 white Dog creature
//	 tokens.
//	 Equip {2}"
//
// The pump is Lead Pipe's PumpAttached. The life ability is the ordinary
// Food ability with the Equipment's own words, so sacrificing it is
// sacrificing a Food. The graveyard trigger is Chromatic Star's
// (WhenThisDies reads any permanent put into a graveyard from the
// battlefield): it fires however the Haunch got there, eaten, destroyed
// or sacrificed to something else, and as it resolves its controller
// may pay {1}{W} (Digsite Conservator's MayPay) for two Dogs. Equip is
// CR 702.6a.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85fdfa3c-4912-496a-bbcb-1ae13a34e917",
		Name:         "Krovod Haunch",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Krovod Haunch — you may pay {1}{W} to create two 1/1 white Dogs", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}{W}",
					Question: "Krovod Haunch — pay {1}{W} to create two 1/1 white Dog creature tokens?",
					Then:     krovodHaunchDogs,
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}, Sacrifice this Equipment: You gain 3 life.",
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue}, // ADR 0142
				Effect:  b36GainLife(3),
			},
			EquipAbility("{2}"),
		},
	})
}

// krovodHaunchDogs is the paid half, keyed (#4264).
var krovodHaunchDogs = PayUnlessBranches("pay-unless/krovod-haunch-dogs",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Template: TokenCard("1/1 white Dog"), N: 2}.Apply(ctx)
	}, nil)
