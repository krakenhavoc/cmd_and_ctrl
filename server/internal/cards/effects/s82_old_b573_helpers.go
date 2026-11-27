package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b573_helpers.go — shared shapes of S82 OLD batch 573 (Guildpact,
// #5599): Dune-Brood Nephilim's Sand token row and Moratorium Stone's
// "exile all cards and permanents with that name" body.

// The token row is added from this file's package-level initializer so the
// shared table's column alignment is not rewritten and it exists before any
// card file's init() asks for it. Dune-Brood Nephilim's Sand was never
// printed as a token card (see knownUnmatchedTokens).
var _ = registerOldB573Tokens()

func registerOldB573Tokens() bool {
	for key, card := range map[string]game.Card{
		"1/1 colorless Sand": {Name: "Sand", TypeLine: "Token Creature — Sand", Power: 1, Toughness: 1},
	} {
		if _, dup := TokenTable[key]; dup {
			panic("s82_old_b573_helpers: token key already in the table: " + key)
		}
		TokenTable[key] = card
	}
	return true
}

// ExileTargetCardAndEveryCardAndPermanentNamedLikeIt is Moratorium Stone's
// second ability: "Exile target nonland card from a graveyard, all other
// cards from graveyards with the same name as that card, and all permanents
// with that name." The names are the target's as it is in its graveyard
// (CR 201.2); the cards and permanents are found as the ability resolves and
// leave together. A target that has left its graveyard makes the ability do
// nothing (CR 608.2b).
func ExileTargetCardAndEveryCardAndPermanentNamedLikeIt(ctx *Context) error {
	target, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	same := SharesANameWith(ExtractionNames(c))
	ids := []uuid.UUID{target}
	for _, seat := range ctx.Game.Seats {
		if seat == nil || seat.Graveyard == nil {
			continue
		}
		for i := range seat.Graveyard.Cards {
			if o := seat.Graveyard.Cards[i]; o.InstanceID != target && same(o) {
				ids = append(ids, o.InstanceID)
			}
		}
	}
	if ctx.Game.Battlefield != nil {
		for i := range ctx.Game.Battlefield.Cards {
			if o := ctx.Game.Battlefield.Cards[i]; same(o) {
				ids = append(ids, o.InstanceID)
			}
		}
	}
	return ExileCardsThen(ctx, ids, game.TailRef{})
}
