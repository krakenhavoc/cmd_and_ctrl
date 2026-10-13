package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neyali, Suns' Vanguard — Legendary Creature — Human Rebel
// {2}{R}{W}, 3/3 (EDHREC rank 3426):
//
//	"Attacking tokens you control have double strike.
//	 Whenever one or more tokens you control attack a player, exile
//	 the top card of your library. During any turn you attacked with
//	 a token, you may play that card."
//
// The Boros token commander. The first printed ability is a layer 6
// static gated on Card.AttackingTarget (AttackingCreaturesHave, declared
// DependsOnAttackingStatus), so a token has double strike the moment it
// is declared or put onto the battlefield attacking and no longer when
// it stops attacking. The second is ONE triggered ability, "whenever one
// or more tokens you control attack A PLAYER" — one declaration is one
// batch and the guard declines every later event of it aimed at the same
// player (see docs/adding-cards.md), so a wide attack on one player
// fires it once, while "it triggers for each player you are attacking
// with one or more tokens" (ruling) is the guard's player dimension
// (OncePerBatchPerPlayer, CR 603.2c / #784). Its body (b32NeyaliAttack)
// exiles the top card of your library if a token is attacking a PLAYER
// as the trigger resolves (a token declared against a planeswalker first
// and one against a player second still count). The permission to play
// it is a conditional exile play (#3192, ADR 0066's second 2026-10-13
// amendment): it lasts for as long as the card stays exiled and is live
// only during a turn its holder attacked with a token
// (game.CondAttackedWithToken, read off the per-turn tally as each
// attacker is declared, so a token that died in combat still counts).
// On a turn no token attacked it is dormant, not gone, and it does not
// depend on Neyali, or her trigger, still being there.
func init() {
	Register(Spec{
		OracleID:     "4012b400-7dcd-43d6-8806-39a3cb743d8f",
		Name:         "Neyali, Suns' Vanguard",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			AttackingCreaturesHave("Attacking tokens you control have double strike.",
				func(c *game.Card) bool { return IsToken(*c) }, "double strike"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32TokenYouControlAttacked(ev, source, g)
			}, b32NeyaliLabel, b32NeyaliAttack)),
		},
	})
}
