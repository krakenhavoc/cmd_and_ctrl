package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malfegor — Legendary Creature — Demon Dragon {2}{B}{B}{R}{R}, 6/6 (Commander
// 2011, #5574):
//
//	"Flying
//	 When Malfegor enters, discard your hand. Each opponent sacrifices a
//	 creature of their choice for each card discarded this way."
//
// Two steps of one ability: the controller discards their whole hand, the
// number of cards that really left it is recorded, and the sacrifices follow
// once any discard prompt (madness) has settled. Each opponent chooses that
// many creatures in APNAP order and they go together; an opponent with fewer
// sacrifices what they have, and an empty hand means nobody sacrifices.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "52e1546a-63eb-48d7-acc4-81b912aabc90",
		Name:            "Malfegor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Malfegor — discard your hand; each opponent sacrifices a creature for each card discarded this way",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						n, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
						if err != nil {
							return err
						}
						ctx.RecordStep(malfegorDiscarded, nil, n)
						return nil
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.S82OldB456EachOpponentSacrificesN(ctx, ctx.StepRecord(malfegorDiscarded).N(),
							"Malfegor — choose a creature to sacrifice for each card discarded")
					}),
				))),
		},
	})
}

// malfegorDiscarded is the number of cards discarded this way.
const malfegorDiscarded = "malfegor/discarded"
