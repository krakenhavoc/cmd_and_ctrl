package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wheel of Misfortune — Sorcery {2}{R}:
//
//	"Each player secretly chooses a number 0 or greater, then all players
//	 reveal those numbers simultaneously and determine the highest and
//	 lowest numbers revealed this way. Wheel of Misfortune deals damage
//	 equal to the highest number to each player who chose that number.
//	 Each player who didn't choose the lowest number discards their hand,
//	 then draws seven cards."
//
// A secret choice round (#5948, ADR 0180 amendment 2026-11-24; #6172):
// every player, the caster first and then in turn order, names a number 0
// or greater with no ceiling (VoteNumber.NoMax). Nobody sees another's
// number until the last is in, then all are revealed together. Only the
// players still in the game count toward the highest and the lowest.
//
// Wheel of Misfortune deals the highest number as damage to each player
// who named it (none for 0); it is still the spell's damage. Then every
// player who did not name the lowest discards their hand, and after all of
// them have, each of them draws seven (Wheel of Fortune's order). When
// every player named the same number nobody discards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     WheelOfMisfortuneOracleID,
		Name:         "Wheel of Misfortune",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var asks []game.VoteAsk
			players := append([]uuid.UUID{ctx.Controller()}, effects.OpponentsInTurnOrder(ctx.Game, ctx.Controller())...)
			for _, p := range players {
				asks = append(asks, game.VoteAsk{
					Voter:    p,
					Question: "Wheel of Misfortune — secretly choose a number 0 or greater (the highest is dealt that much damage; everyone who didn't choose the lowest wheels)",
					Number:   &game.VoteNumber{Min: 0, NoMax: true, SelfDamage: true},
				})
			}
			return effects.SecretChoices{
				Question: "Wheel of Misfortune — secretly choose a number 0 or greater",
				Asks:     asks,
				Then:     wheelOfMisfortuneRevealed,
			}.Apply(ctx)
		},
	})
}

// WheelOfMisfortuneOracleID is the card's oracle_id.
const WheelOfMisfortuneOracleID = "10aff3f6-8347-4e1b-b8e7-c204c9e46657"

// wheelOfMisfortuneRevealed reads the highest and lowest numbers, deals
// the damage, then wheels everyone who didn't name the lowest. An on-disk
// key.
var wheelOfMisfortuneRevealed = effects.VoteResultThen("secret/wheel-of-misfortune", func(ctx *effects.Context, r game.VoteResult) error {
	type named struct {
		player uuid.UUID
		n      int
	}
	var all []named
	hi, lo := -1, -1
	for _, b := range r.Ballots {
		if b.Option != game.VoteNumberIndex || !effects.VoterStillIn(ctx.Game, b.Voter) {
			continue
		}
		all = append(all, named{b.Voter, b.Number})
		if b.Number > hi {
			hi = b.Number
		}
		if lo < 0 || b.Number < lo {
			lo = b.Number
		}
	}
	if len(all) == 0 {
		return nil
	}
	if hi > 0 {
		for _, x := range all {
			if x.n == hi {
				if err := ctx.Game.DealDamageToPlayerForEffect(ctx.Source(), x.player, hi); err != nil {
					return err
				}
			}
		}
	}
	var wheel []uuid.UUID
	for _, x := range all {
		if x.n != lo && effects.VoterStillIn(ctx.Game, x.player) {
			wheel = append(wheel, x.player)
		}
	}
	for _, p := range wheel {
		if _, err := effects.DiscardWholeHand(ctx.Game, p); err != nil {
			return err
		}
	}
	for _, p := range wheel {
		if err := (effects.DrawCards{Player: p, N: 7}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
