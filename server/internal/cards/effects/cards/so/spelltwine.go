package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spelltwine — Sorcery {5}{U} (Commander 2017, #5533):
//
//	"Exile target instant or sorcery card from your graveyard and target
//	 instant or sorcery card from an opponent's graveyard. Copy those
//	 cards. Cast the copies if able without paying their mana costs.
//	 Exile Spelltwine."
//
// Mnemonic Deluge's resolution over two target clauses: both cards are
// exiled through the CR 614 window, and only a card that really reached
// exile is copied (CR 707.12). The copies are cast while Spelltwine is
// still resolving (CR 608.2g), each with a free-cast offer; a copy not cast
// ceases to exist (CR 707.10a). "If able" is the offer itself: a copy that
// cannot be cast is simply not cast. If one target is illegal on
// resolution the other is still exiled and copied (CR 608.2b). Spelltwine
// exiles itself last, whatever else happened.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "307e68e0-5095-474c-af02-bfcb2e463763",
		Name:         "Spelltwine",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
				effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
			effects.TargetCardInGraveyard("target instant or sorcery card from an opponent's graveyard",
				effects.Or(effects.Instant(), effects.Sorcery()), effects.Not(effects.YouOwn())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return effects.ExileTarget{Target: item.SourceCardID}.Apply(ctx)
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, effects.NewContext(ctx.Game, item).ItemTail(spelltwineTail))
		},
	})
}

// spelltwineTail copies each card that reached exile, offers the free casts,
// then exiles Spelltwine. An on-disk key.
var spelltwineTail = effects.CardTail("copy-and-cast/spelltwine-both-then-exile-self", func(ctx *effects.Context, o game.TailOutcome) error {
	item := ctx.Item
	ctx = effects.NewContext(ctx.Game, item)
	if len(o.IDs) > 0 {
		if err := (effects.CopyAndCast{Cards: o.IDs, Free: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ExileTarget{Target: item.SourceCardID}.Apply(ctx)
})
