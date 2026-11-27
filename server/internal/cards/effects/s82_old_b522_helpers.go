package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b522_helpers.go — Eventide, +1/+1 and other counters (S82 older
// sets, batch 522, #5590).

// S82OldB522WithCounter is "creature with a <kind> counter on it".
func S82OldB522WithCounter(kind string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters[kind] > 0 }
}

// S82OldB522DoubleEachKindOfCounter is "Double the number of each kind of
// counter on target permanent" (Gilder Bairn): the first legal target gets as
// many more counters of each kind as it has now, the kinds read once before
// any is added.
func S82OldB522DoubleEachKindOfCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	for _, kind := range MshA6KindsWithCounters(c.Counters) {
		if err := (AddCounter{Target: id, Kind: kind, N: c.Counters[kind]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB522ChooseAnOpponentEntersWithMinusCounters is Canker Abomination's
// "As this creature enters, choose an opponent. This creature enters with a
// -1/-1 counter on it for each creature that player controls": the controller
// is asked as the creature arrives, the answer is stored on it as its chosen
// player, and the counters land before anything else can act (the shape of the
// as-enters choices, see S82OldB113EntersWithChoiceOfCounters).
func S82OldB522ChooseAnOpponentEntersWithMinusCounters(name string) func(*game.Card, *Context) error {
	return func(card *game.Card, ctx *Context) error {
		g := ctx.Game
		var options []game.ChoiceOption
		for _, id := range Opponents.Seats(ctx, card.Controller) {
			label := "Seat ?"
			if p := g.PlayerByIDForEffect(id); p != nil {
				label = p.Name
				if label == "" {
					label = "Seat " + strconv.Itoa(p.Seat+1)
				}
			}
			options = append(options, game.ChoiceOption{Label: label, Player: id})
		}
		if len(options) == 0 {
			return nil
		}
		g.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  card.Controller,
			Source:   card.InstanceID,
			Question: name + " — choose an opponent",
			Options:  options,
			ThenKey:  s82OldB522CankerThen,
			Carry:    []uuid.UUID{card.Controller, card.InstanceID},
		})
		return nil
	}
}

// s82OldB522CankerThen stores the chosen opponent on the creature and puts a
// -1/-1 counter on it for each creature that player controls, if it is still
// on the battlefield (CR 608.2b). An on-disk key: never renamed, never reused.
var s82OldB522CankerThen = OptionPickThen("option-pick/s82-old-b522-canker-abomination", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Option == nil || !OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	id, opp := r.Carry[0], r.Option.Player
	if err := ctx.Game.SetChosenPlayerForEffect(id, opp); err != nil {
		return err
	}
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == opp && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterMinusOne, N: n}.Apply(ctx)
})

// S82OldB522EntersWithMinusCountersUnlessCastAnotherSpellOfColor is "This
// creature enters with N -1/-1 counters on it unless you've cast another
// <colour> spell this turn" (Hotheaded Giant). The tally counts the Giant's own
// cast, so it is taken off when this very card was cast as a spell.
func S82OldB522EntersWithMinusCountersUnlessCastAnotherSpellOfColor(letter string, n int, words string) game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:  "Enters with " + counterCountPhrase(n, game.CounterMinusOne) + " on it unless you've cast another " + words + " spell this turn.",
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID != src.InstanceID {
				return false
			}
			cast := g.CastTallyFor(src.Controller).SpellsCastOfColor(src.TextColor(letter))
			if GapB138WasCast(g, src.InstanceID) {
				cast--
			}
			return cast <= 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterMinusOne, n)
			return nil
		},
	}
}

// S82OldB522EachPlayerReturnsEachCreatureWithMinusCounter is Pyrrhic Revival's
// "Each player returns each creature card from their graveyard to the
// battlefield with an additional -1/-1 counter on it": every creature card in
// every graveyard enters at once under its owner's control.
func S82OldB522EachPlayerReturnsEachCreatureWithMinusCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	return PutCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, ids,
		game.ZoneEntryOptions{EntersWithCounters: map[string]int{game.CounterMinusOne: 1}}, game.TailRef{}, 0)
}
