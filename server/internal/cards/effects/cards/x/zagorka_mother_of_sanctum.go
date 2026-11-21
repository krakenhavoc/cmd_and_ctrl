package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zagorka, Mother of Sanctum — {1}{G}{U}{R} Legendary Creature — Human
// Peasant 4/2 (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Vigilance, trample
//	 At the beginning of your end step, each player may create a tapped
//	 land token named Sanctum with "{T}: Add one mana of any color." Each
//	 opponent who does can't attack you during their next turn."
//
// One trigger that asks each player in turn order from the active player
// (a yes/no each, a chain of keyed prompts): a "yes" creates the Sanctum
// tapped under that player's control, and when that player is an opponent
// of Zagorka's controller they can't attack that controller (the player
// only, not their planeswalkers) during their next turn, a window that
// opens as the answer is given and covers a creature that enters later
// (CR 508.1c). A player who declines, and Zagorka's controller whichever
// way they answer, are not restricted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e444c824-6b9c-4b72-8aa9-a86489c47cbe",
		Name:            "Zagorka, Mother of Sanctum",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Zagorka, Mother of Sanctum — each player may create a tapped Sanctum land token; each opponent who does can't attack you during their next turn",
				func(g *game.Game, item *game.StackItem) error {
					return zagorkaAsk(effects.NewContext(g, item), effects.SeatsFromActive(g))
				}),
		},
	})
}

// zagorkaAsk asks the first of `players` and carries the rest.
func zagorkaAsk(ctx *effects.Context, players []uuid.UUID) error {
	if len(players) == 0 {
		return nil
	}
	return effects.MayChoice{
		Player:   players[0],
		Question: "Zagorka, Mother of Sanctum — create a tapped Sanctum land token?",
		Then:     zagorkaAnswer,
		Carry:    players[1:],
	}.Apply(ctx)
}

// zagorkaAnswer is one player's answer, then the next player's. An
// on-disk key: never renamed, never reused. Registered in init because
// the answer asks the next player and so names the key again.
var zagorkaAnswer game.ConfirmThen

func init() {
	zagorkaAnswer = effects.MayAnswer("confirm/zagorka-sanctum", func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if r.Accepted {
			if err := zagorkaSanctum(ctx, r.Chooser); err != nil {
				return err
			}
		}
		return zagorkaAsk(ctx, r.Carry)
	})
}

func zagorkaSanctum(ctx *effects.Context, player uuid.UUID) error {
	if _, err := ctx.Game.CreateTokensForEffect(player, ctx.TextToken(effects.SanctumToken()), 1, game.TokenEntryOptions{Tapped: true}); err != nil {
		return err
	}
	if player != ctx.Controller() {
		ctx.Game.GrantCantAttackPlayerScopedForEffect(player, ctx.Controller(), game.CantAttackScope{PlayerOnly: true},
			"Zagorka, Mother of Sanctum — can't attack you during their next turn", ctx.Item.SourceCardID)
	}
	return nil
}
