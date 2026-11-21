package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ox Drover — Creature — Human Peasant {3}{W}, 4/4 (Wilds of Eldraine
// Commander, #5368):
//
//	"Vigilance
//	 This creature can't be blocked by Oxen.
//	 Whenever this creature enters or attacks, target opponent creates a
//	 2/4 white Ox creature token and you draw a card."
//
// One ability with two conditions (WhenThisEntersOrAttacks), targeting an
// opponent as it goes on the stack. The Ox is created under that
// opponent's control, then I draw; if the target is gone the whole ability
// is countered on resolution (CR 608.2b), so I draw nothing either. The
// block rule refuses every blocker with the Ox subtype, the Oxen this card
// hands out included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "70f26726-0373-4e50-a1de-37392ca94890",
		Name:            "Ox Drover",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.Subtype("Ox"), "Oxen"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Ox Drover — target opponent creates a 2/4 white Ox creature token and you draw a card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						opp := effects.TargetedPlayer(ctx)
						if opp == uuid.Nil {
							return nil
						}
						if err := (effects.CreateToken{Controller: opp, Template: effects.TokenCard("2/4 white Ox"), N: 1}).Apply(ctx); err != nil {
							return err
						}
						return effects.DrawCards{N: 1}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
