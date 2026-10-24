package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tusk and Whiskers — Legendary Creature — Elephant Mouse Performer
// {3}{G}{W}, 4/4:
//
//	"Whenever you put an ability sticker on a creature, put a +1/+1
//	 counter on that creature.
//	 {2}{G}{W}, {T}: You get {TK}. You may put a sticker on a nonland
//	 permanent you own."
//
// The trigger watches EventStickerPlaced (ADR 0208) for an ability
// sticker on an object that is a creature as the sticker goes on. The
// activated ability is the ticket, then the engine's keyed sticker
// questions, in that order (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de45de7e-5bba-4612-8bb7-755710d93ecf",
		Name:         "Tusk and Whiskers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStickerPlaced, youPutAnAbilityStickerOnACreature,
				"Tusk and Whiskers — put a +1/+1 counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					_, on := PlacedStickerKind(item)
					return AddCounter{Target: on, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{W}, {T}: You get {TK}. You may put a sticker on a nonland permanent you own.",
			Cost:    Plus(ManaCost("{2}{G}{W}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  GetTicketsThenSticker(1),
		}},
	})
}

// youPutAnAbilityStickerOnACreature is "whenever you put an ability
// sticker on a creature".
func youPutAnAbilityStickerOnACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	if kind, _ := game.StickerEventKindText(ev); kind != game.StickerAbility {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}
