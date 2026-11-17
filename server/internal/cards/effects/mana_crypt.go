package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Crypt — Artifact {0} (Eternal Masters, #5542):
//
//	"At the beginning of your upkeep, flip a coin. If you lose the flip,
//	 this artifact deals 3 damage to you.
//	 {T}: Add {C}{C}."
//
// The upkeep trigger is The Gold Saucer's keyed coin call (CR 705.2),
// with the trigger carried on the prompt so the payoff reads the Crypt
// it came from. A won flip does nothing. A lost flip has the Crypt deal
// 3 damage to the trigger's controller, even when it has left the
// battlefield in response, as it last existed (Juzám Djinn's damage).
// The mana ability is Sol Ring's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c63e4e1-89d2-4bc6-a232-94e75c4b1c8a",
		Name:         "Mana Crypt",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Mana Crypt — flip a coin; if you lose the flip, it deals 3 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Mana Crypt — call the coin flip",
						ThenKey:  manaCryptFlipped,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// manaCryptFlipped is the settled call: a lost flip has the Crypt deal 3
// damage to the trigger's controller. An on-disk key: never renamed,
// never reused.
var manaCryptFlipped = CoinFlipThen("coin-flip/mana-crypt-damage", func(ctx *Context, r game.CoinFlipped) error {
	if wonFirstFlip(r) {
		return nil
	}
	return DealDamage{Source: ctx.Item.SourceCardID, Target: ctx.Controller(), Amount: 3}.Apply(ctx)
})
