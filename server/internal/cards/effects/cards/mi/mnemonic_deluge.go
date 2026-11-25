package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mnemonic Deluge — Sorcery {6}{U}{U}{U} (Commander Legends):
//
//	"Exile target instant or sorcery card from a graveyard. Copy that card
//	 three times. You may cast the copies without paying their mana costs.
//	 Exile Mnemonic Deluge."
//
// Psionic Ritual's resolution with three copies: the card is exiled through
// the CR 614 window and only a card that really reached exile is copied
// (CR 707.12), three times, each copy offered a free cast while the Deluge
// is still resolving (CR 608.2g). Copies not cast cease to exist
// (CR 707.10a); the card itself stays in exile. The Deluge exiles itself
// last, whether or not a card was exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56de4cdc-4db0-48ec-8a16-f66086507233",
		Name:         "Mnemonic Deluge",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target instant or sorcery card from a graveyard",
			effects.Or(effects.Instant(), effects.Sorcery())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, effects.NewContext(ctx.Game, item).ItemTail(mnemonicDelugeTail))
		},
	})
}

// mnemonicDelugeTail copies what reached exile three times with a free-cast
// offer for each, then exiles the Deluge. An on-disk key.
var mnemonicDelugeTail = effects.CardTail("copy-and-cast/mnemonic-deluge-three-copies", func(ctx *effects.Context, o game.TailOutcome) error {
	g, item := ctx.Game, ctx.Item
	ctx = effects.NewContext(g, item)
	var thrice []uuid.UUID
	for _, id := range o.IDs {
		thrice = append(thrice, id, id, id)
	}
	if err := (effects.CopyAndCast{Cards: thrice, Free: true}).Apply(ctx); err != nil {
		return err
	}
	return effects.ExileTarget{Target: item.SourceCardID}.Apply(ctx)
})
