package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warmonger — Creature — Minotaur Monger {3}{R}, 3/3:
//
//	"{2}: This creature deals 1 damage to each creature without flying
//	 and each player. Any player may activate this ability."
//
// An any-player row (CR 602.2, 602.1b): whoever activates it pays the
// {2} out of their own pool (CR 602.1a). Warmonger is the source of the
// damage (CR 120.3), read as it last existed if it has left (CR
// 113.7a), and it is a creature without flying, so it hits itself.
// "Without flying" is read post-layer as the ability resolves, and
// every player is hit, the activator included
// (effects_damage_each_creature_and_player.go).
//
// Its Purpose is a creature sweep, which buys a player who does not
// control it nothing the bot prices (ADR 0106 owner decision 2: draws
// and the controller's life loss), so the bot does not reach across the
// table for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bae92332-1a0b-476b-8719-190e8d8cc03a",
		Name:         "Warmonger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}: This creature deals 1 damage to each creature without flying and each player. Any player may activate this ability.",
			// ADR 0126 §6: only creatures without flying.
			Purpose:   game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:      ManaCost("{2}"),
			AnyPlayer: true,
			Effect:    thisDealsDamageToEachCreatureMatchingAndEachPlayer(WithoutKeyword("flying"), 1),
		}},
	})
}
