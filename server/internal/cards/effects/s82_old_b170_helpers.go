package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b170_helpers.go — shared bodies for the S82 older-sets batch 170
// (War of the Spark, #5520): the Bond cycle's exile-and-return clauses, Price
// of Betrayal's counter removal and Deliver Unto Evil's opponent choice.

// S82OldB170IsInstantOrSorcery is "instant and/or sorcery card".
func S82OldB170IsInstantOrSorcery(c game.Card) bool {
	return c.IsInstant() || c.IsSorcery()
}

// s82OldB170BondOfInsightReturn is the mill ended: "Return up to two instant
// and/or sorcery cards from your graveyard to your hand. Exile Bond of
// Insight." The spell is exiled before the graveyard is read, which is the
// paper order's result: a spell that is still resolving is never in the
// graveyard to be chosen. An on-disk key.
var s82OldB170BondOfInsightReturn = game.RegisterTail("s82-old-b170/bond-of-insight-return", func(g *game.Game, o game.TailOutcome) error {
	if err := g.ExileCardForEffect(o.Tail.Source); err != nil {
		return err
	}
	return S82OldB117ChooseFromGraveyardToHand(g, o.Tail.Player, o.Tail.Source,
		"Bond of Insight — return up to two instant and/or sorcery cards from your graveyard to your hand",
		S82OldB170IsInstantOrSorcery, 0, 2)
})

// S82OldB170BondOfInsight is Bond of Insight's whole text: each player mills
// four cards, then the controller returns up to two instant and/or sorcery
// cards, then the spell is exiled.
func S82OldB170BondOfInsight(_ *game.StackItem, ctx *Context) error {
	return MillEachPlayerInTurn(ctx, ApnapPlayers(ctx.Game), ConstantMill(4),
		game.KeyedTail{Ref: s82OldB170BondOfInsightReturn, Player: ctx.Controller(), Source: ctx.Source()})
}

// S82OldB170RemoveCounters is "remove up to <n> counters from <target>": the
// controller names how many of each kind, a permanent's or a player's, and
// every kind chosen comes off. The bot takes -1/-1 counters off its own
// permanent and everything else off another's, and any counter but poison off
// an opponent.
var S82OldB170RemoveCounters = RegisterCounterChoice("number/s82-old-b170-remove-counters",
	func(ctx *Context, target uuid.UUID, kind string, most int) int {
		if ctx.Game.PlayerByIDForEffect(target) != nil {
			if kind == game.CounterPoison {
				return 0
			}
			return most
		}
		mine := false
		if c, ok := ctx.Game.LookupCardForEffect(target); ok {
			mine = c.Controller == ctx.Controller()
		}
		if (kind == game.CounterMinusOne) == mine {
			return most
		}
		return 0
	},
	func(ctx *Context, target uuid.UUID, _ []uuid.UUID, picked map[string]int) error {
		kinds := make([]string, 0, len(picked))
		for kind := range picked {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			n := picked[kind]
			if n <= 0 {
				continue
			}
			if ctx.Game.PlayerByIDForEffect(target) != nil {
				if err := ctx.Game.AddPlayerCounterByForEffect(ctx.Controller(), target, kind, -n); err != nil {
					return err
				}
				continue
			}
			if err := (AddCounter{Target: target, Kind: kind, N: -n}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

// S82OldB170DeliverOpponent is "an opponent" named by the controller (the
// engine skips the question when one opponent is left): the named opponent
// chooses two of the cards the spell targeted.
var S82OldB170DeliverOpponent = OptionPickThen("option-pick/s82-old-b170-deliver-unto-evil-opponent",
	func(ctx *Context, r game.OptionPicked) error {
		if r.Option == nil || r.Option.Player == uuid.Nil || ctx.Game.PlayerByIDForEffect(r.Option.Player) == nil {
			// Nobody is left to choose: the cards stay in the graveyard.
			return nil
		}
		var cards []uuid.UUID
		for _, id := range r.Carry {
			if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
				cards = append(cards, id)
			}
		}
		if len(cards) == 0 {
			return nil
		}
		n := min(2, len(cards))
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    r.Option.Player,
			FromPlayer: ctx.Controller(),
			Source:     r.Source,
			Question:   "Deliver Unto Evil — choose two of these cards to leave in their graveyard; the rest go to their hand",
			Cards:      cards,
			Min:        n,
			Max:        n,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    s82OldB170DeliverReturn,
			Carry:      append([]uuid.UUID{ctx.Controller()}, cards...),
		})
		return nil
	})

// s82OldB170DeliverReturn is the opponent's pick answered: the chosen cards
// stay, every other target still in the graveyard goes to its owner's hand.
// The first carried ID is the spell's controller.
var s82OldB170DeliverReturn = game.RegisterChooseCardsThen("choose-cards/s82-old-b170-deliver-unto-evil-return",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Carry) < 1 {
			return nil
		}
		picked := map[uuid.UUID]bool{}
		for _, id := range r.Picked {
			picked[id] = true
		}
		for _, id := range r.Carry[1:] {
			if picked[id] {
				continue
			}
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			if err := g.ReturnFromGraveyardForEffect(id, game.ZoneHand); err != nil {
				return err
			}
		}
		return nil
	})

// S82OldB170DeliverUntoEvil is Deliver Unto Evil's whole text: the choice or
// the return first, then "Exile Deliver Unto Evil" as its last instruction.
func S82OldB170DeliverUntoEvil(item *game.StackItem, ctx *Context) error {
	if err := s82OldB170DeliverBody(ctx, LegalTargetCardIDs(ctx)); err != nil {
		return err
	}
	return ctx.Game.ExileCardForEffect(item.SourceCardID)
}

// s82OldB170DeliverBody returns the cards with a Bolas planeswalker, and
// otherwise names the opponent who chooses two of them to leave. The choice is
// a prompt: the cards it leaves are known once it is answered.
func s82OldB170DeliverBody(ctx *Context, cards []uuid.UUID) error {
	if len(cards) == 0 {
		return nil
	}
	if B42ControlsPlaneswalkerNamed(ctx.Game, ctx.Controller(), "Bolas") {
		for _, id := range cards {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	ctx.Game.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  ctx.Controller(),
		Among:    Opponents.Seats(ctx, ctx.Controller()),
		Source:   ctx.Source(),
		Question: "Deliver Unto Evil — choose an opponent to choose two of the cards",
		ThenKey:  S82OldB170DeliverOpponent,
		Carry:    append([]uuid.UUID{ctx.Controller()}, cards...),
	})
	return nil
}

// S82OldB170AttackedOrBlocked is "creature that attacked or blocked this
// turn".
func S82OldB170AttackedOrBlocked() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && (g.AttackedThisTurn(c.InstanceID) || g.BlockedThisTurn(c.InstanceID))
	}
}

// S82OldB170ReturnModeTargetToHand is a bullet's "return target <card> from
// your graveyard to your hand": the bullet's own target, if it is still legal
// as the spell resolves (CR 608.2b).
func S82OldB170ReturnModeTargetToHand(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
}
