package cardse

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "0eecdfb3-3b05-4051-a660-060ff6df80ef",
		Name:         "Emergent Ultimatum",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return emergentUltimatumSearch(ctx)
		},
	})
}

const EmergentUltimatumName = "Emergent Ultimatum"

// emergentUltimatumSearch is the search, and from its continuation the
// rest. Everything the later steps need is a scalar or an ID, so an undo
// across a prompt resolves against the restored game.
func emergentUltimatumSearch(ctx *effects.Context) error {
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:   ctx.Controller(),
		Source:   ctx.Source(),
		Pred:     func(c game.Card) bool { return len(c.EffectiveColors()) == 1 },
		Dest:     game.ZoneExile,
		Limit:    3,
		Reveal:   true,
		Reason:   "Emergent Ultimatum — up to three monocolored cards with different names, to exile",
		Rule:     effects.SearchRuleDifferentNames,
		ThenTail: ctx.ItemTail(emergentUltimatumChooseTail),
	})
}

// emergentUltimatumChooseTail has an opponent pick one of the exiled cards.
// An on-disk key.
var emergentUltimatumChooseTail = effects.CardTail("emergent-ultimatum/choose", func(ctx *effects.Context, o game.TailOutcome) error {
	return emergentUltimatumChoose(ctx, o.IDs)
})

// emergentUltimatumChoose has an opponent pick one of the exiled cards.
// Keyed (#4253): the found cards ride the prompts as data.
func emergentUltimatumChoose(ctx *effects.Context, found []uuid.UUID) error {
	if len(found) == 0 {
		return emergentUltimatumExileSelf(ctx)
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Emergent Ultimatum — choose an opponent to pick one of the exiled cards",
		ThenKey:  EmergentUltimatumOpponent,
		Carry:    found,
	}.Apply(ctx)
}

// emergentUltimatumOpponent is the answer to "choose an opponent": that
// opponent picks one of the found cards. Nobody to name leaves the spell
// to exile itself. An on-disk key: never renamed, never reused. #4253 slice
// 10b: the one keyed ChoosePlayer form (ChoosePlayerThen), the same key.
var EmergentUltimatumOpponent = effects.ChoosePlayerThen("option-pick/emergent-ultimatum-opponent", func(ctx *effects.Context, chosen uuid.UUID, found []uuid.UUID) error {
	if chosen == uuid.Nil {
		return emergentUltimatumExileSelf(ctx)
	}
	queued, err := ctx.Game.RevealPickThenForEffect(game.RevealPickPrompt{
		Chooser:  chosen,
		Owner:    ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Emergent Ultimatum — choose one of these cards; it is shuffled into their library and they may cast the others for free",
		Cards:    found,
		Min:      1,
		Max:      1,
		ThenKey:  EmergentUltimatumPicked,
	})
	if err != nil || queued > 0 {
		return err
	}
	// Nobody could be asked: nothing is picked and every card is left.
	return EmergentUltimatumSettle(ctx, nil, found)
})

// emergentUltimatumPicked is the answer to the opponent's pick: the cards
// they did not choose are the ones the caster may cast. An on-disk key:
// never renamed, never reused.
var EmergentUltimatumPicked = game.RegisterChooseCardsThen("choose-cards/emergent-ultimatum-picked", func(g *game.Game, r game.CardsChosen) error {
	var left []uuid.UUID
	for _, id := range r.Candidates {
		if !slices.Contains(r.Picked, id) {
			left = append(left, id)
		}
	}
	return EmergentUltimatumSettle(effects.KeyedContext(g, r.From, r.Source), r.Picked, left)
})

// emergentUltimatumSettle shuffles the chosen card into its owner's
// library, then offers the other cards, then exiles the spell. The tuck's
// "then" is a registered tail (#4253), the cards left riding its Carry, so
// a CR 614 / 616 / 903.9 prompt the tuck pauses on is a restore point.
func EmergentUltimatumSettle(ctx *effects.Context, picked, left []uuid.UUID) error {
	controller, source := ctx.Controller(), ctx.Source()
	if len(picked) == 0 {
		return emergentUltimatumFinish(ctx.Game, controller, source, left)
	}
	tail := game.KeyedTail{Ref: emergentUltimatumTucked, Player: controller, Source: source, Carry: append([]uuid.UUID(nil), left...)}
	return ctx.Game.TuckCardsToLibraryKeyedForEffect(picked[:1], game.TuckOptions{}, tail)
}

// emergentUltimatumTucked is the rest of the spell once the chosen card's
// tuck has settled: shuffle Player's library, then offer the cards left (the
// Carry) and exile the spell (Source). An on-disk key.
var emergentUltimatumTucked = game.RegisterTail("emergent-ultimatum/tucked", func(g *game.Game, o game.TailOutcome) error {
	if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	return emergentUltimatumFinish(g, o.Tail.Player, o.Tail.Source, o.Tail.Carry)
})

// emergentUltimatumFinish offers the cards left and exiles the spell.
func emergentUltimatumFinish(g *game.Game, controller, source uuid.UUID, left []uuid.UUID) error {
	effects.GrantFreeCasts(g, controller, source, EmergentUltimatumName, game.TimingFlash, game.LapseStaysInExile, left)
	return g.ExileCardForEffect(source)
}

// emergentUltimatumExileSelf is "Exile Emergent Ultimatum."
func emergentUltimatumExileSelf(ctx *effects.Context) error {
	return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
}
