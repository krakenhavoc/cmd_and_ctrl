package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pay_unless_payments_6189.go — #6189: the cumulative-upkeep payments
// the prompt could not take. Each one is a game.PayAction kind
// (game/pay_unless_action.go) with the printed words the prompt shows.
//
//	702.24a Cumulative upkeep is a triggered ability that imposes an
//	        increasing cost on a permanent. "Cumulative upkeep [cost]"
//	        means "At the beginning of your upkeep, if this permanent is
//	        on the battlefield, put an age counter on this permanent.
//	        Then you may pay [cost] for each age counter on it. If you
//	        don't, sacrifice it." If [cost] has choices associated with
//	        it, each choice is made separately for each age counter, ...
//
// Each payment therefore multiplies by the age count (UpkeepPayment.times),
// and the ones that choose are repeatable: with two age counters the payer
// may name one creature twice.

// ManaAndLifePayment is a mana cost paid together with life (#6189):
// Infernal Darkness' "Cumulative upkeep—Pay {B} and 1 life" is
// ManaAndLifePayment("{B}", 1). Both multiply with the age counters.
func ManaAndLifePayment(mana string, life int) UpkeepPayment {
	return UpkeepPayment{Mana: mana, Life: life}
}

// CounterOnOpponentCreaturePayment is "Put a <kind> counter on a
// <noun> an opponent controls" as a payment (Sheltering Ancient):
// `of` are the permanents that may take one, `n` how many counters this
// payment puts (one per age counter), each on a creature of the payer's
// choice — the same one twice if they like.
func CounterOnOpponentCreaturePayment(n int, kind, noun, plural string, of game.PermanentQuery) UpkeepPayment {
	words := "Put a " + kind + " counter on " + articleFor(noun) + " " + noun + " an opponent controls"
	if n != 1 {
		words = "Put " + NumberWord(n) + " " + kind + " counters on " + plural + " opponents control"
	}
	return UpkeepPayment{
		Action: &game.PayAction{
			Kind: game.PayActionCounterOnOpponentCreature, Count: n,
			Of: of, CounterKind: kind, Repeats: true,
		},
		Words: words, Noun: noun, Plural: plural,
	}
}

// CountersOnThisPayment is "Put a <kind> counter on this creature" as a
// payment (Aboroth): `source` is the permanent, and `n` counters go on
// it at once (one per age counter). It names nothing.
func CountersOnThisPayment(n int, kind string, source uuid.UUID) UpkeepPayment {
	words := "Put a " + kind + " counter on this permanent"
	if n != 1 {
		words = "Put " + NumberWord(n) + " " + kind + " counters on this permanent"
	}
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionCountersOnSource, Count: n, CounterKind: kind, OnSource: source},
		Words:  words,
	}
}

// OpponentGainsLifePayment is "An opponent gains N life" as a payment
// (Wall of Shards): `picks` opponents are named, one per age counter
// (the same opponent may be named twice), each gaining `life`.
func OpponentGainsLifePayment(picks, life int) UpkeepPayment {
	words := "An opponent gains " + strconv.Itoa(life) + " life"
	if picks != 1 {
		words = "Opponents gain " + strconv.Itoa(life) + " life " + NumberWord(picks) + " times"
	}
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionOpponentGainsLife, Count: picks, Amount: life, PicksPlayers: true, Repeats: true},
		Words:  words,
	}
}

// OpponentCreatesTokenPayment is "Have an opponent create a <noun>" as a
// payment (Varchild's War-Riders): `key` is the token's registered row
// (game.RegisterPayActionToken), and `picks` opponents are named, one per
// age counter, each creating one token.
func OpponentCreatesTokenPayment(picks int, key, noun string) UpkeepPayment {
	words := "Have an opponent create " + articleFor(noun) + " " + noun
	if picks != 1 {
		words = "Have opponents create " + NumberWord(picks) + " " + noun + " tokens"
	}
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionOpponentCreatesToken, Count: picks, TokenKey: key, PicksPlayers: true, Repeats: true},
		Words:  words, Noun: noun,
	}
}

// GainControlOfLandPayment is "Gain control of a <noun> you don't
// control" as a payment (Herald of Leshrac): `of` are the permanents
// that may be taken, and `n` distinct ones are named (one per age
// counter).
func GainControlOfLandPayment(n int, noun, plural string, of game.PermanentQuery) UpkeepPayment {
	words := "Gain control of " + articleFor(noun) + " " + noun + " you don't control"
	if n != 1 {
		words = "Gain control of " + NumberWord(n) + " " + plural + " you don't control"
	}
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionGainControlOfLand, Count: n, Of: of},
		Words:  words, Noun: noun, Plural: plural,
	}
}

// DrawPayment is "Draw a card" as a payment (Psychic Vortex): `n` cards,
// one per age counter. It names nothing and is always a payment — a draw
// from an empty library is one the player loses for later, not an
// unpayable cost.
func DrawPayment(n int) UpkeepPayment {
	words := "Draw a card"
	if n != 1 {
		words = "Draw " + NumberWord(n) + " cards"
	}
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionDraw, Count: n}, Words: words}
}

// ExileLibraryTopPayment is "Exile the top card of your library" as a
// payment (Thought Lash): `n` cards, one per age counter. A library with
// fewer than `n` cards cannot pay (CR 118.3).
func ExileLibraryTopPayment(n int) UpkeepPayment {
	words := "Exile the top card of your library"
	if n != 1 {
		words = "Exile the top " + NumberWord(n) + " cards of your library"
	}
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionExileLibraryTop, Count: n}, Words: words}
}
