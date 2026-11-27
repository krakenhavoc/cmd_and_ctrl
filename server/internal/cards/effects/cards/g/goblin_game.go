package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Game — Sorcery {5}{R}{R} (Planeshift, S82 OLD b642, #5615):
//
//	"Each player hides at least one item, then all players reveal them
//	 simultaneously. Each player loses life equal to the number of items they
//	 revealed. The player who revealed the fewest items then loses half their life,
//	 rounded up. If two or more players are tied for fewest, each loses half their
//	 life, rounded up."
//
// The items are a number: a secret choice round (Wheel of Misfortune's, #5948) in
// which every player, the caster first and then in turn order, secretly names a
// number 1 or greater with no ceiling. Nobody sees another's number until the last
// is in, then all are revealed together. Every player loses that much life; the
// lowest number (every player who named it, on a tie) then loses half their life
// as it stands after that loss, rounded up (CR 107.1a). Only players still in the
// game count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a10c945-ad8b-47b6-8cbd-7b4e2612a9b0",
		Name:         "Goblin Game",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var asks []game.VoteAsk
			players := append([]uuid.UUID{ctx.Controller()}, effects.OpponentsInTurnOrder(ctx.Game, ctx.Controller())...)
			for _, p := range players {
				asks = append(asks, game.VoteAsk{
					Voter:    p,
					Question: "Goblin Game — secretly hide at least one item (choose a number 1 or greater; you lose that much life, and the fewest also lose half their life)",
					Number:   &game.VoteNumber{Min: 1, NoMax: true, SelfDamage: true},
				})
			}
			return effects.SecretChoices{
				Question: "Goblin Game — secretly hide at least one item",
				Asks:     asks,
				Then:     goblinGameRevealed,
			}.Apply(ctx)
		},
	})
}

// goblinGameRevealed is Goblin Game's resolution once every number is revealed.
// An on-disk key.
var goblinGameRevealed = effects.VoteResultThen("secret/goblin-game", func(ctx *effects.Context, r game.VoteResult) error {
	type named struct {
		player uuid.UUID
		n      int
	}
	var all []named
	fewest := -1
	for _, b := range r.Ballots {
		if b.Option != game.VoteNumberIndex || !effects.VoterStillIn(ctx.Game, b.Voter) {
			continue
		}
		all = append(all, named{b.Voter, b.Number})
		if fewest < 0 || b.Number < fewest {
			fewest = b.Number
		}
	}
	for _, x := range all {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), x.player, -x.n); err != nil {
			return err
		}
	}
	for _, x := range all {
		if x.n != fewest {
			continue
		}
		p := ctx.Game.PlayerByIDForEffect(x.player)
		if p == nil {
			continue
		}
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), x.player, -effects.OldB642HalfLifeRoundedUp(p.Life)); err != nil {
			return err
		}
	}
	return nil
})
