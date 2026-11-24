package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Neera, Wild Mage — Legendary Creature — Human Elf Shaman {4}{U}{R}, 2/7
// (Commander Legends: Battle for Baldur's Gate, S82 GAP2 b05 #5948, #6094):
//
//	"Whenever you cast a spell, you may put it on the bottom of its owner's
//	 library. If you do, reveal cards from the top of your library until
//	 you reveal a nonland card. You may cast that card without paying its
//	 mana cost. Then put all revealed cards not cast this way on the bottom
//	 of your library in a random order. This ability triggers only once
//	 each turn."
//
// The trigger is once each turn (TriggersOnlyOnceEachTurn), on any spell
// its controller casts; the "you may" is asked as it resolves (CR 603.5).
// The spell is moved, not countered (CR 701.6a), to its owner's library
// bottom (PutSpellInLibraryKeyedForEffect): a spell that can't be countered
// still moves, and the rest runs only when it really got there ("if you
// do": a spell already gone, or a commander its owner sent to the command
// zone, stops it). Then the reveal and the free cast from the library
// (Gap2B05NeeraRevealAndMayCastFree), the rest to the bottom in a random
// order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88fb1dd1-60fb-4f51-ac4d-30b43156a474",
		Name:         "Neera, Wild Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.TriggersOnlyOnceEachTurn(game.TriggeredAbility{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor == source.Controller
			},
			Key: "Neera, Wild Mage — put that spell on the bottom of its owner's library, then reveal until a nonland card and cast it free",
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Neera, Wild Mage — put that spell on the bottom of its owner's library? If you do, reveal until a nonland card; you may cast it free.",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				spell := ctx.Trigger().Event.CardID
				if spell == uuid.Nil {
					return nil
				}
				return g.PutSpellInLibraryKeyedForEffect(spell, game.TuckOptions{ToBottom: true}, ctx.ItemTail(neeraTucked))
			},
		})},
	})
}

// neeraTucked is "if you do": the reveal runs only when the spell reached
// the library. An on-disk key.
var neeraTucked = effects.CardTail("then/neera-wild-mage-tucked", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.Gap2B05NeeraRevealAndMayCastFree(ctx)
})
