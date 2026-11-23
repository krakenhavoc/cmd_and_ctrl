package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Dragon-Kami Reborn // Dragon-Kami's Egg — an Enchantment — Saga
// {2}{G} (Kamigawa: Neon Dynasty, S82):
//
//	I, II — You gain 2 life. Look at the top three cards of your library.
//	        Exile one of them face down with a hatching counter on it, then
//	        put the rest on the bottom of your library in any order.
//	III   — Exile this Saga, then return it to the battlefield
//	        transformed under your control.
//
// Back face (registered under "<oracle_id>#1"), Dragon-Kami's Egg, an
// Enchantment Creature — Egg 0/1:
//
//	"Whenever this creature or a Dragon you control dies, you may cast a
//	 creature spell from among cards you own in exile with hatching
//	 counters on them without paying its mana cost."
//
// The exile is face down and unseen (game.ExileCardsFaceDownChosenKeyedForEffect):
// nobody looks at the card until the Egg's permission names it, which is
// when its owner becomes its one viewer. The hatching counter is placed from
// the exile's keyed tail, once the card has landed there. The cast is
// Goliath Daydreamer's grant: one permission over every such creature card,
// priced {0}, spent by one cast, offered during the trigger's resolution
// (CR 608.2g). A card not cast keeps its counter for the next death.
//
// No simplification.
func init() {
	const name = "The Dragon-Kami Reborn"
	chapter := func(n int) game.TriggeredAbility {
		return effects.ChapterTrigger(n, effects.SagaChapterLabel(name, n, "gain 2 life; look at the top three cards, exile one face down with a hatching counter, rest on the bottom"),
			dragonKamiRebornLook)
	}
	effects.Register(effects.Spec{
		OracleID:     TheDragonKamiRebornOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1),
			chapter(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile this Saga, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	egg := effects.OnAny([]game.EventKind{game.EventLTB},
		func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
			if effects.ThisDied(ev, source, ch, g) {
				return true
			}
			dead, ok := effects.DiedCreature(ev, g)
			return ok && dead.HasSubtype(source.TextSubtype("Dragon")) && effects.LeftUnderControlOf(ev, dead) == source.Controller
		},
		"Dragon-Kami's Egg — you may cast a creature spell from among cards you own in exile with hatching counters on them without paying its mana cost",
		dragonKamisEggCast)
	egg.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Dragon-Kami's Egg — cast a creature spell from among your exiled cards with hatching counters without paying its mana cost?"}
	effects.Register(effects.Spec{
		OracleID:     TheDragonKamiRebornOracleID + "#1",
		Name:         "Dragon-Kami's Egg",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{egg},
	})
}

// theDragonKamiRebornOracleID is shared with the back face.
const TheDragonKamiRebornOracleID = "c3a46eb3-38d9-4f47-9b71-26c9ea7ef1ce"

// dragonKamiRebornLook is chapters I and II: gain 2 life, then look at the
// top three cards and exile one face down with a hatching counter.
func dragonKamiRebornLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.GainLife{Player: item.Controller, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	looked := g.LookAtTopOfPlayersLibraryForEffect(item.Controller, item.Controller, 3)
	if len(looked) == 0 {
		return nil
	}
	if len(looked) == 1 {
		return dragonKamiRebornExile(g, item.Controller, item.SourceCardID, looked, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "The Dragon-Kami Reborn — exile one of them face down with a hatching counter on it",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    dragonKamiRebornPicked,
		Carry:      looked,
	})
	return nil
}

// dragonKamiRebornPicked is the card chosen: the carry is the cards looked at.
// An on-disk key: never renamed, never reused.
var dragonKamiRebornPicked = game.RegisterChooseCardsThen("choose-cards/the-dragon-kami-reborn-pick", func(g *game.Game, r game.CardsChosen) error {
	return dragonKamiRebornExile(g, r.Chooser, r.Source, r.Carry, r.Picked)
})

// dragonKamiRebornExile exiles the picked card face down, then (tail) puts
// the hatching counter on it and the rest on the bottom in any order.
func dragonKamiRebornExile(g *game.Game, player, source uuid.UUID, looked, picked []uuid.UUID) error {
	took := make(map[uuid.UUID]bool, len(picked))
	for _, id := range picked {
		took[id] = true
	}
	var rest []uuid.UUID
	for _, id := range looked {
		if !took[id] {
			rest = append(rest, id)
		}
	}
	return g.ExileCardsFaceDownChosenKeyedForEffect(player, picked,
		effects.WithCarry(game.KeyedTail{Ref: dragonKamiRebornExiled, Player: player, Source: source}, rest...))
}

// dragonKamiRebornExiled places the counter on what landed in exile, then
// puts the rest of the looked-at cards on the bottom of the library.
var dragonKamiRebornExiled = effects.CardTail("the-dragon-kami-reborn/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			if err := (effects.AddCounter{Target: id, Kind: effects.CounterHatching, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	rest := effects.CardsStillInALibrary(ctx.Game, o.Tail.Carry)
	if len(rest) == 0 {
		return nil
	}
	return ctx.Game.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   o.Tail.Player,
		Source:    o.Tail.Source,
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    effects.LibraryOrderLabel(ctx.Game, o.Tail.Source, "put the rest on the bottom of your library in any order"),
	})
})

// dragonKamisEggCast is the Egg's trigger: a permission over every creature
// card the controller owns in exile with a hatching counter, {0}, one cast.
func dragonKamisEggCast(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	var cards []game.Card
	for _, c := range effects.NeoB19HatchingExileCards(g, controller) {
		if c.IsCreature() {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		Filter:      game.PermissionFilter{CreatureOnly: true},
		LapseOnPass: game.LapseStaysInExile,
		Source:      item.SourceCardID,
		SourceName:  "Dragon-Kami's Egg",
		Label:       "Dragon-Kami's Egg — cast a creature spell with a hatching counter without paying its mana cost",
	}, cards)
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		ids = append(ids, c.InstanceID)
	}
	effects.OfferResolutionCastOver(g, controller, item.SourceCardID, "Dragon-Kami's Egg",
		game.MayCastKeywordFreeCast, effects.DeclineLeaveInExile, ids...)
	return nil
}
