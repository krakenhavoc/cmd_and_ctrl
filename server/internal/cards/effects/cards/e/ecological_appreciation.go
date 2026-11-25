package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ecological Appreciation — Sorcery {X}{2}{G}:
//
//	"Search your library and graveyard for up to four creature cards with
//	 different names that each have mana value X or less and reveal them.
//	 An opponent chooses two of those cards. Shuffle the chosen cards into
//	 your library and put the rest onto the battlefield. Exile Ecological
//	 Appreciation."
//
// Gifts Ungiven's split (Elemental Teachings' shape) over two zones
// (#6156):
//
//   - The search names both zones, so both are searched (BothZones, CR
//     701.23a) and no zone question is asked. The found cards are revealed
//     and stay where they were found (HoldFound, CR 701.20a) while the
//     opponent chooses: a graveyard card does not leave the graveyard
//     before the instruction that moves it.
//   - "An opponent" names no seat: the caster picks which opponent
//     (ChoosePlayer) and that opponent picks two (RevealPick, clamped to
//     what was found, CR 608.2's "as much as possible").
//   - The chosen cards are shuffled into the library: one found in the
//     graveyard is put into the library first, then the library is
//     shuffled. The rest enter the battlefield together, from whichever
//     zone each was in, under the caster's control. Then the spell exiles
//     itself.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     EcologicalAppreciationOracleID,
		Name:         "Ecological Appreciation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player: ctx.Controller(),
				Source: ctx.Source(),
				Pred: func(c game.Card) bool {
					return c.IsCreature() && c.ManaValue() <= x
				},
				Dest:          game.ZoneLibrary,
				HoldFound:     true,
				AlsoGraveyard: true,
				BothZones:     true,
				Limit:         4,
				Reveal:        true,
				Reason:        "Ecological Appreciation — up to four creature cards with different names and mana value X or less",
				Rule:          effects.SearchRuleDifferentNames,
				ThenTail:      ctx.ItemTail(ecologicalAppreciationChooseTail),
			})
		},
	})
}

// EcologicalAppreciationOracleID is the card's oracle_id.
const EcologicalAppreciationOracleID = "0635ab65-fc1a-4e43-9e11-3a1ad3093830"

// ecologicalAppreciationChooseTail has the caster name the opponent who
// chooses. An on-disk key.
var ecologicalAppreciationChooseTail = effects.CardTail("ecological-appreciation/choose", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Ecological Appreciation — choose an opponent to pick two of the revealed cards to shuffle into your library",
		ThenKey:  ecologicalAppreciationOpponent,
		Carry:    o.IDs,
	}.Apply(ctx)
})

// ecologicalAppreciationOpponent is the answer: that opponent picks two.
// Nobody to name chooses nothing, so every found card enters. An on-disk
// key.
var ecologicalAppreciationOpponent = effects.ChoosePlayerThen("option-pick/ecological-appreciation-opponent", func(ctx *effects.Context, chosen uuid.UUID, found []uuid.UUID) error {
	if chosen == uuid.Nil {
		return ecologicalAppreciationSettle(ctx, nil, found)
	}
	return effects.RevealPick{
		Player:   chosen,
		Owner:    ctx.Controller(),
		Question: "Ecological Appreciation — choose two of these cards; they are shuffled into their library and the rest enter the battlefield",
		Cards:    found,
		Min:      2,
		Max:      2,
		Keyed:    ecologicalAppreciationPicked,
	}.Apply(ctx)
})

var ecologicalAppreciationPicked = effects.RegisterRevealPick("choose-cards/ecological-appreciation-settle", func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
	return ecologicalAppreciationSettle(ctx, picked, left)
})

// ecologicalAppreciationSettle shuffles the chosen cards into the library
// (tucking any still in the graveyard first), then puts the rest onto the
// battlefield.
func ecologicalAppreciationSettle(ctx *effects.Context, picked, left []uuid.UUID) error {
	me := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	var tuck []uuid.UUID
	for _, id := range picked {
		if me != nil && me.Graveyard.Contains(id) {
			tuck = append(tuck, id)
		}
	}
	tail := game.KeyedTail{Ref: ecologicalAppreciationTucked, Player: ctx.Controller(), Source: ctx.Source(), Carry: append([]uuid.UUID(nil), left...)}
	if len(tuck) == 0 {
		return ecologicalAppreciationRest(ctx.Game, ctx.Controller(), ctx.Source(), left)
	}
	return ctx.Game.TuckCardsToLibraryKeyedForEffect(tuck, game.TuckOptions{}, tail)
}

// ecologicalAppreciationTucked is the rest of the spell once the chosen
// graveyard cards' tuck has settled, the cards left riding its Carry. An
// on-disk key.
var ecologicalAppreciationTucked = game.RegisterTail("ecological-appreciation/tucked", func(g *game.Game, o game.TailOutcome) error {
	return ecologicalAppreciationRest(g, o.Tail.Player, o.Tail.Source, o.Tail.Carry)
})

// ecologicalAppreciationRest shuffles the caster's library, then puts the
// cards left onto the battlefield from wherever each still is, then exiles
// the spell.
func ecologicalAppreciationRest(g *game.Game, player, source uuid.UUID, left []uuid.UUID) error {
	if err := g.ShuffleLibraryForEffect(player); err != nil {
		return err
	}
	me := g.PlayerByIDForEffect(player)
	var entries []game.BatchEntry
	for _, id := range left {
		switch {
		case me == nil:
		case me.Library.Contains(id):
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneLibrary})
		case me.Graveyard.Contains(id):
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneGraveyard})
		}
	}
	then := game.KeyedTail{Ref: ecologicalAppreciationExileSelf, Player: player, Source: source}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{Controller: player}, then)
}

// ecologicalAppreciationExileSelf is "Exile Ecological Appreciation." An
// on-disk key.
var ecologicalAppreciationExileSelf = game.RegisterTail("ecological-appreciation/exile-self", func(g *game.Game, o game.TailOutcome) error {
	return g.ExileCardForEffect(o.Tail.Source)
})
