package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tea, Earl Grey, Hot — Artifact — Food {2} (Star Trek, #4973):
//
//	"When this artifact enters, draw a card.
//	 {2}, {T}, Sacrifice this artifact: Choose one —
//	 • This artifact deals 2 damage to target opponent.
//	 • You gain 3 life."
//
// A Food with its own sacrifice ability rather than the Food token's. The
// mode is chosen and its target picked as the ability is activated
// (CR 602.2b, 601.2b-c), and the sacrifice is a cost. The damage is dealt
// by the artifact, which has left the battlefield by then, so its
// last-known information is the source (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f2e8083-29af-4cdb-a2cb-216b2f3f5616",
		Name:         "Tea, Earl Grey, Hot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tea, Earl Grey, Hot — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Choose one — This artifact deals 2 damage to target opponent; or you gain 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Modes: ChooseOne(
				ModeDoing("This artifact deals 2 damage to target opponent.", TargetPlayer("target opponent", Opponent()),
					modeSourceDealsDamageToTargetPlayer(2)),
				ModeDoing("You gain 3 life.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
					}),
			),
		}},
	})
}
