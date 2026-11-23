package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Davros, Dalek Creator — Legendary Artifact Creature — Alien Scientist
// {1}{U}{B}{R}, 3/4:
//
//	"Menace
//	 At the beginning of your end step, create a 3/3 black Dalek
//	 artifact creature token with menace if an opponent lost 3 or more
//	 life this turn. Then each opponent who lost 3 or more life this
//	 turn faces a villainous choice — You draw a card, or that player
//	 discards a card."
//
// A villainous choice (CR 701.55, #3140), faced by each opponent who lost
// three or more life this turn, in APNAP order, one at a time (CR
// 701.55d). Both halves read the turn's life loss (the per-player tally
// the other "lost life this turn" cards read) as the trigger resolves,
// so they are not an intervening "if": a trigger put on the stack and
// then robbed of its condition still resolves, and simply does less.
// The opponent who picks "you draw a card" gives the controller the
// draw; the other is their own discard from hand.
//
// No simplifications.
var davrosChoice = effects.NewVillainousChoice("davros-dalek-creator",
	"Davros — the controller draws a card, or you discard a card?",
	effects.VillainousOption{
		Label: "The controller of Davros draws a card",
		Harm:  1,
		Body: func(ctx *effects.Context, _ effects.Facing) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	},
	effects.VillainousOption{
		Label:    "Discard a card",
		Harm:     1,
		Possible: func(g *game.Game, facing uuid.UUID) bool { return effects.VillainousHandSize(g, facing) > 0 },
		Body:     effects.VillainousDiscard(1),
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:        "25c0bbc8-b0c3-4166-967f-e8b096487a6c",
		Name:            "Davros, Dalek Creator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Davros — create a Dalek if an opponent lost 3 or more life this turn, then those opponents face a villainous choice",
				effects.Do(
					// The choice waits for a prompt the token's creation asks (CR 608.2c,
					// #4155). The opponents who lost 3 life are found again for it: making
					// a token does not change how much life was lost this turn.
					effects.Instruction(func(ctx *effects.Context) error {
						if len(davrosLosers(ctx)) == 0 {
							return nil
						}
						return effects.CreateToken{Template: effects.TokenCard("3/3 black Dalek artifact with menace"), N: 1}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.FaceVillainousChoice{Players: davrosLosers(ctx), Choice: davrosChoice}.Apply(ctx)
					}),
				)),
		},
	})
}

// davrosLosers are the opponents who lost 3 or more life this turn.
func davrosLosers(ctx *effects.Context) []uuid.UUID {
	var losers []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if effects.B18LifeLostThisTurn(ctx.Game, opp) >= 3 {
			losers = append(losers, opp)
		}
	}
	return losers
}
