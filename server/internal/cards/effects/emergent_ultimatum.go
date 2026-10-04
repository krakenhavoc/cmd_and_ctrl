package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emergent Ultimatum — Sorcery {B}{B}{G}{G}{G}{U}{U}:
//
//	"Search your library for up to three monocolored cards with
//	 different names and exile them. An opponent chooses one of those
//	 cards. Shuffle that card into your library. You may cast the other
//	 cards without paying their mana costs. Exile Emergent Ultimatum."
//
// Gifts Ungiven's search and Intuition's opponent pick, ending in
// Cascade's free cast:
//
//   - The search takes up to three cards for which every one is
//     monocolored (exactly one colour: a colourless card and a gold card
//     both fail) and no two share a name (DifferentNames, the set rule
//     Gifts Ungiven uses). They go to EXILE face up. The library is not
//     shuffled by the search: the only shuffle the card prints is the
//     one that takes the chosen card back in.
//   - "An opponent chooses" names no seat, so the caster picks which
//     opponent (ChoosePlayer, Slithermuse's shape), and that opponent
//     picks exactly one of the exiled cards (RevealPick), seeing all of
//     them because exile is public.
//   - That card is tucked into its owner's library and the library is
//     shuffled. The others get a free-cast GRANT (free_cast_grants.go,
//     ADR 0066's posture on every "you may cast it" a resolution
//     offers): {0}, flash timing because the printed cast happens during
//     resolution (CR 608.2g), good until the caster's next priority
//     pass. A card not cast stays in exile, as printed.
//   - The spell exiles itself last, so it is not left in the graveyard.
//
// A found set of one is chosen by default and shuffled back, with
// nothing left to cast; a search that finds nothing skips the middle and
// still exiles the spell.
//
// No simplification beyond the cast-as-a-grant posture every such card
// shares.
func init() {
	Register(Spec{
		OracleID:     "0eecdfb3-3b05-4051-a660-060ff6df80ef",
		Name:         "Emergent Ultimatum",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return emergentUltimatumSearch(item, ctx)
		},
	})
}

const emergentUltimatumName = "Emergent Ultimatum"

// emergentUltimatumSearch is the search, and from its continuation the
// rest. Everything the later steps need is a scalar or an ID, so an undo
// across a prompt resolves against the restored game.
func emergentUltimatumSearch(item *game.StackItem, ctx *Context) error {
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:   ctx.Controller(),
		Source:   ctx.Source(),
		Pred:     func(c game.Card) bool { return len(c.EffectiveColors()) == 1 },
		Dest:     game.ZoneExile,
		Limit:    3,
		Reveal:   true,
		Reason:   "Emergent Ultimatum — up to three monocolored cards with different names, to exile",
		Validate: DifferentNames,
		Then: func(g *game.Game, found []uuid.UUID) error {
			return emergentUltimatumChoose(NewContext(g, item), found)
		},
	})
}

// emergentUltimatumChoose has an opponent pick one of the exiled cards.
func emergentUltimatumChoose(ctx *Context, found []uuid.UUID) error {
	if len(found) == 0 {
		return emergentUltimatumExileSelf(ctx)
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Emergent Ultimatum — choose an opponent to pick one of the exiled cards",
		Then: func(ctx *Context) error {
			opp := ctx.ChosenPlayer()
			if opp == uuid.Nil {
				return emergentUltimatumExileSelf(ctx)
			}
			return RevealPick{
				Player:   opp,
				Owner:    ctx.Controller(),
				Question: "Emergent Ultimatum — choose one of these cards; it is shuffled into their library and they may cast the others for free",
				Cards:    found,
				Min:      1,
				Max:      1,
				Then:     emergentUltimatumSettle,
			}.Apply(ctx)
		},
	}.Apply(ctx)
}

// emergentUltimatumSettle shuffles the chosen card into its owner's
// library, then offers the other cards, then exiles the spell.
func emergentUltimatumSettle(ctx *Context, picked, left []uuid.UUID) error {
	item := ctx.Item
	controller, source := ctx.Controller(), ctx.Source()
	finish := func(g *game.Game, _ bool) error {
		grantFreeCasts(g, controller, source, emergentUltimatumName, game.TimingFlash, game.LapseStaysInExile, left)
		return emergentUltimatumExileSelf(NewContext(g, item))
	}
	if len(picked) == 0 {
		return finish(ctx.Game, false)
	}
	return ctx.Game.TuckToLibraryThenForEffect(picked[0], game.TuckOptions{}, func(g *game.Game, ok bool) error {
		if err := g.ShuffleLibraryForEffect(controller); err != nil {
			return err
		}
		return finish(g, ok)
	})
}

// emergentUltimatumExileSelf is "Exile Emergent Ultimatum."
func emergentUltimatumExileSelf(ctx *Context) error {
	return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
}
