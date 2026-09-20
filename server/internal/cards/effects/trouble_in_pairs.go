package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trouble in Pairs — Enchantment {2}{W}{W}:
//
//	"If an opponent would begin an extra turn, that player skips that
//	 turn instead.
//	 Whenever an opponent attacks you with two or more creatures,
//	 draws their second card each turn, or casts their second spell
//	 each turn, you draw a card."
//
// The white stax enchantment that taxes everything a Commander table
// does twice. The second sentence is one printed ability with three
// trigger conditions; it is declared here as three entries rather
// than one watching three event kinds, because they key on different
// per-event data and only the attack half wants the CR 603.1 batch
// collapse. No event can satisfy two of them, so the behaviour is
// the same and each condition reads as its own sentence.
//
// All three counters come off tallies the engine already keeps and
// bumps BEFORE the harvester sees the event, so "their second" is a
// tally of exactly two — the trigger fires once, on the draw or cast
// that reaches two, and not on every one after it. The attack half is
// OncePerBatch: the engine emits one EventAttack per declared
// attacker, so without it a five-creature alpha strike would draw
// four extra cards.
//
// "Attacks YOU" is the defending player, not merely "attacks", which
// is what separates this from Firemane Commando — an opponent who
// swings the whole team at somebody else gives you nothing. An attack
// at your planeswalker or battle counts as an attack at you.
//
// One clause is dropped, and it is inert rather than unimplemented:
// nothing in this engine can take an extra turn, so a replacement
// that skipped one would never be consulted. It comes back with extra
// turns (#753).
func init() {
	Register(Spec{
		OracleID:     "f349f58b-8cc8-45e4-9565-2b46fdf976c9",
		Name:         "Trouble in Pairs",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The clause that stops an opponent taking an extra turn does nothing yet, because no card in the game can take one.",
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentAttackedYouWithAtLeast(ev, source, g, 2, troubleInPairsAttackLabel)
			}, troubleInPairsAttackLabel, Do(DrawCards{N: 1}))),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentDrewTheirSecondCardThisTurn(ev, source, g)
			}, troubleInPairsDrawLabel, Do(DrawCards{N: 1})),
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentCastTheirSecondSpellThisTurn(ev, source, g)
			}, troubleInPairsCastLabel, Do(DrawCards{N: 1})),
		},
	})
}

// The three stack labels. The attack one is also the per-turn key the
// batch guard reads, so it has to be stable.
const (
	troubleInPairsAttackLabel = "Trouble in Pairs — an opponent attacked you with two or more creatures: draw a card"
	troubleInPairsDrawLabel   = "Trouble in Pairs — an opponent drew their second card: draw a card"
	troubleInPairsCastLabel   = "Trouble in Pairs — an opponent cast their second spell: draw a card"
)
