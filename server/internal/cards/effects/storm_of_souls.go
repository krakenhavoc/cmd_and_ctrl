package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm of Souls — Sorcery {4}{W}{W} (Bloomburrow Commander):
//
//	"Return all creature cards from your graveyard to the battlefield. Each
//	 of them is a 1/1 Spirit with flying in addition to its other types.
//	 Exile Storm of Souls."
//
// Every creature card in your graveyard returns in one simultaneous entry
// (CR 614.12; ReturnFromGraveyardTogether), the set read as the spell
// resolves. "Each of them is a 1/1 Spirit with flying in addition to its
// other types" says how they enter (CR 614.1c), so it rides the entry as an
// overwrite (ADR 0210, Excava's shape): the Spirit subtype in layer 4, flying
// in layer 6 and the base 1/1 in layer 7b, pinned to that object. "Spirit" is
// a word of the card's text (CR 612.2). The spell then exiles itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7022ae9a-2589-493b-a050-a230fd02f80b",
		Name:         "Storm of Souls",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (ReturnFromGraveyardTogether{
				Targets:    graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() }),
				Controller: ctx.Controller(),
				EntersAs: &game.EntersAs{
					AddSubtypes: ctx.TextSubtypes("Spirit"),
					Keywords:    []string{"flying"},
					BasePT:      &game.EntersAsPT{Power: 1, Toughness: 1},
					Label:       "Storm of Souls — a 1/1 Spirit with flying in addition to its other types",
				},
			}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
