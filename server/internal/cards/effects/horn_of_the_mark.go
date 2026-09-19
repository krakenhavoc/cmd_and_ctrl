package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horn of the Mark — Legendary Artifact {2}:
//
//	"Whenever two or more creatures you control attack a player, look
//	 at the top five cards of your library. You may reveal a creature
//	 card from among them and put it into your hand. Put the rest on
//	 the bottom of your library in a random order."
//
// The card #952 was filed for, and the one that pays for the
// primitive: look at N, take a named subset to hand, route the rest.
// It is LookAtTopThenMayTakeToHand and nothing else
// (take_from_library.go) — the whole of the second and third
// sentences, with the library-to-hand move going through the engine's
// named door rather than borrowing "return it to its owner's hand".
//
// Three things about the trigger that are the rule rather than
// defensive coding:
//
//   - "TWO OR MORE creatures … attack a PLAYER" is one trigger per
//     defending player, not one per attacker. EventAttack fires per
//     creature (a three-creature swing is three events), so the
//     ability is OncePerBatchPerPlayer: the CR 603.2c key with the
//     defender as its second dimension, so an attack split across two
//     opponents that puts two creatures on each triggers twice and one
//     that puts three on one opponent triggers once.
//   - The COUNT is read off the board at the lock-in, not off the
//     event. CR 508.1 declares attackers as one turn-based action, and
//     the engine announces the whole declaration in one batch
//     (#830/#859), so by the time the first EventAttack is harvested
//     every attacker already carries its AttackingTarget. Counting
//     events instead would see one.
//   - The Horn is an artifact and never attacks, so "creatures you
//     control" is measured against the Horn's controller
//     (`ev.Actor == source.Controller`), which is also what makes a
//     stolen Horn tax the thief's attacks and not the owner's.
//
// The look is PRIVATE — only its controller sees the five — and the
// card taken is REVEALED, which is exactly what the printed sentence
// says and what TakeFromLibraryToHand.Reveal does. The bottom is a
// RANDOM order because the card prints "in a random order"; there is
// no caveat here, unlike Goblin Ringleader's "in any order".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b836c32-84f0-41ae-b7d9-67b92c743c60",
		Name:         "Horn of the Mark",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Actor != source.Controller {
						return false
					}
					defender := b17DefendingPlayer(g, ev)
					return defender != uuid.Nil &&
						b41AttackersAgainst(g, source.Controller, defender) >= 2
				},
				"Horn of the Mark — look at the top five cards of your library",
				LookAtTopThenMayTakeToHand(5, Creature(), 1,
					"Horn of the Mark — you may reveal a creature card from among them and put it into your hand"))),
		},
	})
}

// b41AttackersAgainst counts the creatures `attacker` controls that are
// attacking `player` right now — the "two or more creatures you control
// attack a player" count, measured against combat state rather than
// against the event stream.
//
// An attack at a planeswalker or a battle that player controls counts
// as attacking them (CR 506.2, CR 508.1d), which is what
// DefendingPlayerForAttackForEffect resolves and what keeps the count
// agreeing with the trigger's own per-player key.
func b41AttackersAgainst(g *game.Game, attacker, player uuid.UUID) int {
	n := 0
	for _, id := range b13AttackingCreaturesYouControl(g, attacker) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if g.DefendingPlayerForAttackForEffect(c.AttackingTarget) == player {
			n++
		}
	}
	return n
}
