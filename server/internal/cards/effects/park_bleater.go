package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Park Bleater — Creature — Goat Employee {2}{W}, 2/2:
//
//	"Whenever another creature you own enters, you get {TK}.
//	 {W}, {T}: You may put a sticker on a creature you own that entered
//	 this turn."
//
// "You own" reads the entering creature's owner; the ticket goes to Park
// Bleater's controller, the "you" of its ability. The sticker is the
// engine's keyed questions (ADR 0208 §4) over the creatures its controller
// owns that entered this turn (Game.EnteredThisTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02f31e00-646d-4971-8704-1583bbfcddd6",
		Name:         "Park Bleater",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, anotherCreatureYouOwnEntered, "Park Bleater — you get {TK}", Do(GetTickets{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: You may put a sticker on a creature you own that entered this turn.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(PutASticker{On: StickerOnCreatureYouOwnThatEnteredThisTurn}),
		}},
	})
}

// anotherCreatureYouOwnEntered is "whenever another creature you own
// enters".
func anotherCreatureYouOwnEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller && c.IsCreature()
}
