package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm of Memories — Sorcery {2}{R}{R}{R} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn.)
//	 Exile an instant or sorcery card with mana value 3 or less from your
//	 graveyard at random. You may cast it without paying its mana cost.
//	 If that spell would be put into a graveyard, exile it instead."
//
// Storm is the engine keyword (CR 702.40a), so each copy is its own
// resolution of the same instruction. The card is drawn uniformly from
// the qualifying cards in your graveyard as each copy resolves, through
// the keyed random stream (Arcane Bombardment's pick), so a copy that
// resolves after an earlier one exiled a card cannot pick that card
// again. No target: an empty pool simply exiles nothing.
//
// The cast is the free-cast grant Creative Technique makes over the
// exiled card (ADR 0099 §4): {0}, so X is 0 and additional costs are
// still paid, offered during the resolution (CR 608.2g, #4250), which is
// the ruling's "go on the stack as each copy resolves, and each resolves
// before more copies resolve". "If that spell would be put into a
// graveyard, exile it instead" is the permission's
// ExileInsteadOfGraveyard. A declined card stays in exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f8ca795-d6fd-4e3e-911c-c621a942acbb",
		Name:         "Storm of Memories",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g, you := ctx.Game, ctx.Controller()
			pick := g.ChooseAtRandomForEffect(randomDraw(ctx), graveyardIDs(g, you, func(c game.Card) bool {
				return (c.IsInstant() || c.IsSorcery()) && c.ManaValue() <= 3
			}), 1)
			if len(pick) == 0 {
				return nil
			}
			const name = "Storm of Memories"
			perm := g.FreeCastFromExilePermission(you, item.SourceCardID, name)
			perm.AltCostKey = game.AltCostKeyGrantedCost
			perm.ExileInsteadOfGraveyard = true
			return exileWithPermissionThenOffer(g, pick[0], item.SourceCardID, perm, name)
		},
	})
}
