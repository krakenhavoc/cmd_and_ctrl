package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien, Monsters' Advocate — Legendary Planeswalker — Vivien {3}{G}{G},
// loyalty 3 (Ikoria: Lair of Behemoths, S82 OLD batch 120, #5512):
//
//	"You may look at the top card of your library any time.
//	 You may cast creature spells from the top of your library.
//	 +1: Create a 3/3 green Beast creature token. Put your choice of a reach
//	 counter, a vigilance counter, or a trample counter on it.
//	 −2: When you next cast a creature spell this turn, search your library for
//	 a creature card with lesser mana value, put it onto the battlefield, then
//	 shuffle."
//
// The first two lines are Mm'menon's pair: the top card is visible to its
// controller (CR 401.5) and a cast-only permission over creature spells from
// the top. The +1 makes the Beast through the token door with a keyed tail,
// which asks the controller for the counter once the token has landed and puts
// it on that token; the keyword counters grant their keyword through the
// layers. The −2 creates a delayed trigger (CR 603.7) for the next creature
// spell its controller casts this turn, which fires once; it searches for a
// creature card whose mana value is less than that spell's, read as last
// known if the spell has left the stack (CR 608.2h, {X} counted as chosen),
// puts it onto the battlefield and shuffles, finding nothing if there is no
// such card.
//
// No simplification.
func init() {
	castFromTop := effects.PlayFromTopOfYourLibrary(game.PermissionFilter{CardType: "creature"},
		"Cast a creature spell from the top of your library (Vivien, Monsters' Advocate)")
	castFromTop.CastOnly = true
	effects.Register(effects.Spec{
		OracleID:          "7256b03e-4fd5-4a0a-9b20-7b88870edfec",
		Name:              "Vivien, Monsters' Advocate",
		Completeness:      effects.CompletenessFull,
		StartingLoyalty:   3,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions:   []game.CastPermission{castFromTop},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 3/3 green Beast creature token. Put your choice of a reach counter, a vigilance counter, or a trample counter on it.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Cost:    effects.LoyaltyCost(1),
				Effect:  vivienMonstersAdvocateBeast,
			},
			{
				Label: "−2: When you next cast a creature spell this turn, search your library for a creature card with lesser mana value, put it onto the battlefield, then shuffle.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.WhenYouNextCast(
						"Vivien, Monsters' Advocate — search your library for a creature card with lesser mana value and put it onto the battlefield",
						game.CastFilter{Types: []string{"Creature"}}, vivienMonstersAdvocateSearch).Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// vivienMonstersAdvocateBeast makes the Beast and hands it to the counter
// question.
func vivienMonstersAdvocateBeast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     ctx.Source(),
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(effects.TokenCard("3/3 green Beast")),
			Count:    1,
		}},
	}, ctx.ItemTail(vivienMonstersAdvocateAsk))
}

// vivienMonstersAdvocateAsk asks which counter the new Beast gets. An
// on-disk key: never renamed, never reused.
var vivienMonstersAdvocateAsk = effects.CardTail("vivien-monsters-advocate/beast-counter-choice", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.PickOption{
		Question: "Vivien, Monsters' Advocate — put a reach counter, a vigilance counter, or a trample counter on the Beast",
		Options:  []game.ChoiceOption{{Label: "Reach"}, {Label: "Vigilance"}, {Label: "Trample"}},
		ThenKey:  vivienMonstersAdvocateCounter,
		Carry:    []uuid.UUID{o.IDs[0]},
	}.Apply(ctx)
})

// vivienMonstersAdvocateCounter puts the chosen counter on the Beast. An
// on-disk key: never renamed, never reused.
var vivienMonstersAdvocateCounter = effects.OptionPickThen("option-pick/vivien-monsters-advocate-counter", func(ctx *effects.Context, r game.OptionPicked) error {
	kinds := []string{game.CounterReach, game.CounterVigilance, game.CounterTrample}
	if r.Index < 0 || r.Index >= len(kinds) || len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return effects.AddCounter{Target: r.Carry[0], Kind: kinds[r.Index], N: 1}.Apply(ctx)
})

// vivienMonstersAdvocateSearch is the delayed trigger's body: the creature
// card with lesser mana value than the spell that was cast. An on-disk key.
var vivienMonstersAdvocateSearch = game.SimpleDelayedBody("vivien-monsters-advocate/search-lesser-creature",
	func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		spells := ctx.PayloadCards()
		if len(spells) == 0 {
			return nil
		}
		mv, ok := g.LastKnownSpellManaValueForEffect(spells[0])
		if !ok {
			return nil
		}
		return effects.SearchLibrary{
			Player:    item.Controller,
			Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() < mv },
			Dest:      game.ZoneBattlefield,
			Limit:     1,
			Shuffle:   true,
			Reason:    "Vivien, Monsters' Advocate — a creature card with lesser mana value",
			Source:    item.SourceCardID,
		}.Apply(ctx)
	})
