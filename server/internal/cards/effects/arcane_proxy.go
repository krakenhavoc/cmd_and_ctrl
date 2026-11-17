package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcane Proxy — Artifact Creature — Wizard {7}, 4/3 (The Brothers' War,
// #5656): "Prototype {1}{U}{U} — 2/1", "When this creature enters, if you
// cast it, exile target instant or sorcery card with mana value less than
// or equal to this creature's power from your graveyard. Copy that card.
// You may cast the copy without paying its mana cost."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out. "If you
// cast it" is an intervening if (CR 603.4) rechecked as the ability
// resolves. The target is limited by the Proxy's power (2 prototyped, 4
// otherwise) when it is chosen and again as it resolves
// (RelativeToSource). The resolution is Kaervek's exile-and-copy with the
// free cast (CopyAndCast, CR 707.12): only a card that really reached exile
// is copied; the copy is cast, or not, during the resolution (CR 608.2g),
// and a copy not cast ceases to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "792f8235-f23b-4858-bdd4-c7a186b7a470",
		Name:         "Arcane Proxy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				s82BroB21EntersIfCast("Arcane Proxy — exile target instant or sorcery card with mana value less than or equal to this creature's power from your graveyard; copy it and you may cast the copy without paying its mana cost",
					arcaneProxyExileAndCopy),
				RelativeToSource(
					TargetCardInGraveyard("target instant or sorcery card with mana value less than or equal to Arcane Proxy's power from your graveyard",
						YouOwn(), Or(Instant(), Sorcery())),
					ManaValueNoGreaterThanPower())),
		},
	})
}

func arcaneProxyExileAndCopy(ctx *Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(arcaneProxyCopyExiled))
}

// arcaneProxyCopyExiled copies what reached exile and offers the free cast.
var arcaneProxyCopyExiled = CardTail("arcane-proxy/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	return CopyAndCast{Cards: o.IDs, Free: true}.Apply(NewContext(ctx.Game, ctx.Item))
})
