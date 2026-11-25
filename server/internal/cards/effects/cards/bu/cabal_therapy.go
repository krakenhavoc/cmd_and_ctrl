package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cabal Therapy — Sorcery {B} (Eternal Masters, S82 OLD, #5542):
//
//	"Choose a nonland card name. Target player reveals their hand and discards
//	 all cards with that name.
//	 Flashback—Sacrifice a creature. (You may cast this card from your
//	 graveyard for its flashback cost. Then exile it.)"
//
// The target player is chosen on cast; the name is asked as the spell resolves,
// for any card name (CR 201.2: free text), through the keyed card-name prompt
// (the order printed: name first, then the reveal). The player reveals their
// hand and discards every nonland card with that name, all at once and not by
// their choice. A land named does nothing. The flashback is Dread Return's
// sacrifice-only cost (no mana), the creature chosen when the spell is cast
// from the graveyard; the card is exiled afterwards (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "018107d1-ebf7-4657-bc91-c8374af55cc5",
		Name:             "Cabal Therapy",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetPlayer("target player"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.FlashbackSacrifice(1, "a creature", effects.Creature())},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim := effects.FirstLegalTargetPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Cabal Therapy — choose a nonland card name",
				effects.WithCarry(ctx.ItemTail(cabalTherapyNamed), victim))
			return nil
		},
	})
}

// cabalTherapyNamed is the reveal and the discard once the name (the tail's
// Word) is known; the player is the tail's Carry. An on-disk key.
var cabalTherapyNamed = effects.CardTail("then/s82-old-b279-cabal-therapy-named", func(ctx *effects.Context, o game.TailOutcome) error {
	victim := effects.TailCarry0(o)
	if victim == uuid.Nil {
		return nil
	}
	return effects.RevealHandAndDiscardAllNonlandNamed(ctx, victim, o.Tail.Word)
})
