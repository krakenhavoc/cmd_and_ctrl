package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Endless Ranks of HYDRA — Sorcery {3}{B} (Marvel Super Heroes
// Commander, #4455):
//
//	"For each opponent, you create a 2/1 black Villain creature token
//	 with menace.
//	 Whenever your commander enters or attacks, you may pay {1}{B}. If
//	 you do, return this card from your graveyard to your hand."
//
// The spell makes one Villain per opponent still in the game. The second
// ability works only from the graveyard (a trigger of a sorcery has no
// other place to be): it watches a commander its controller owns entering
// the battlefield or being declared as an attacker, and offers the
// {1}{B} payment as it resolves. The payment is real mana (a decline
// returns nothing); the return checks the card is still in the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fb76cdf-438b-48f5-aceb-db637381d052",
		Name:         "Endless Ranks of HYDRA",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("2/1 black Villain with menace"), N: len(ctx.Opponents())}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{endlessRanksOfHydraReturn()},
	})
}

// endlessRanksOfHydraReturn is the graveyard half of Endless Ranks of
// HYDRA.
func endlessRanksOfHydraReturn() game.TriggeredAbility {
	t := OnAny([]game.EventKind{game.EventETB, game.EventAttack}, msc10YourCommanderEntersOrAttacks,
		"Endless Ranks of HYDRA — you may pay {1}{B} to return this card to your hand",
		func(g *game.Game, item *game.StackItem) error {
			return MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}{B}",
				Question: "Endless Ranks of HYDRA — pay {1}{B} to return it from your graveyard to your hand?",
				Then:     endlessRanksOfHydraPaid,
			}.Apply(NewContext(g, item))
		})
	t.Zones = []game.ZoneKind{game.ZoneGraveyard}
	return t
}

var endlessRanksOfHydraPaid = PayUnlessBranches("pay-unless/endless-ranks-of-hydra-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return returnThisCardFromYourGraveyardToYourHand(ctx.Game, ctx.Item)
	}, nil)

// msc10YourCommanderEntersOrAttacks is "whenever your commander enters
// or attacks": the creature of the event is a commander the source's
// owner owns.
func msc10YourCommanderEntersOrAttacks(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB && ev.Kind != game.EventAttack {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCommander && c.Owner == source.Owner
}
