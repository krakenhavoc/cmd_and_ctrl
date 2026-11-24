package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yusri, Fortune's Flame — Legendary Creature — Efreet {1}{U}{R}, 2/3
// (Modern Horizons 2, #5501):
//
//	"Flying
//	 Whenever Yusri attacks, choose a number between 1 and 5. Flip that many
//	 coins. For each flip you win, draw a card. For each flip you lose,
//	 Yusri deals 2 damage to you. If you won five flips this way, you may
//	 cast spells from your hand this turn without paying their mana costs."
//
// The number is chosen as the trigger resolves (a keyed number prompt, so a
// table waiting on it is a restore point) and then handed to a called coin
// flip of that many coins. Each win draws a card and each loss is 2 damage
// from Yusri to the controller, whatever order the flips fall in. Five wins
// open a standing, end-of-turn {0} cast permission over the controller's
// hand: it covers cards drawn later this turn and keeps each spell's own
// timing, and a land is never cast. The damage reads Yusri as it last
// existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c0a0a7d-b12f-47c8-a6a6-d046f999ec6b",
		Name:            "Yusri, Fortune's Flame",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Yusri — choose a number between 1 and 5, flip that many coins, draw for each win, take 2 for each loss",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChooseNumber{
						Min: 1, Max: 5,
						Question: "Yusri, Fortune's Flame — choose a number between 1 and 5",
						Unit:     game.PayAmountOther,
						Goal:     func(*effects.Context) int { return 3 },
						Then:     yusriNumberChosen,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// yusriNumberChosen is the chosen number's branch: flip that many coins.
var yusriNumberChosen = effects.NumberThen("number/yusri-fortunes-flame-coins",
	func(ctx *effects.Context, r game.NumberAnswer) error {
		if r.Number < 1 {
			return nil
		}
		ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
			Flipper:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Yusri, Fortune's Flame — call the coin flips",
			Coins:    r.Number,
			ThenKey:  yusriFlipped,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	})

// yusriFlipped is the settled flips: a card for each win, 2 damage for each
// loss, and five wins open the free casts.
var yusriFlipped = effects.CoinFlipThen("coin-flip/yusri-fortunes-flame", func(ctx *effects.Context, r game.CoinFlipped) error {
	wins, losses := 0, 0
	for _, w := range r.Result.Won {
		if w {
			wins++
		} else {
			losses++
		}
	}
	if wins > 0 {
		if err := (effects.DrawCards{Player: ctx.Controller(), N: wins}).Apply(ctx); err != nil {
			return err
		}
	}
	if losses > 0 {
		if err := (effects.DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 2 * losses}).Apply(ctx); err != nil {
			return err
		}
	}
	if wins == 5 {
		ctx.Game.GrantCastPermissionForEffect(game.CastPermission{
			Player:     ctx.Controller(),
			Zone:       game.ZoneHand,
			Scope:      game.ScopeStanding,
			Cost:       "{0}",
			AltCostKey: game.AltCostKeyGrantedCost,
			CastOnly:   true,
			Filter:     game.PermissionFilter{NonLandOnly: true},
			Source:     ctx.Source(),
			Label:      "Yusri, Fortune's Flame — cast it without paying its mana cost",
		})
	}
	return nil
})
