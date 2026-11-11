package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eumidian Wastewaker — Creature — Insect Cleric {2}{B}{B}, 4/4
// (Foundations Commander):
//
//	"Whenever this creature attacks, you and defending player each discard
//	 a card or sacrifice a permanent. You draw a card for each land card
//	 put into a graveyard this way.
//	 Encore {6}{B}{B} (...)"
//
// Two players answer in turn, you first and then the defending player (the
// ruling). Each is asked "discard a card or sacrifice a permanent"
// (PickOption); a player who cannot discard (no cards in hand) must
// sacrifice and is not asked, and one with nothing to sacrifice must
// discard. The chosen half is the engine's own discard or sacrifice prompt,
// keyed so the table stays a restore point, and what each player put into a
// graveyard rides the chain as data. When both have answered, the draw
// counts the land CARDS among them that really are in a graveyard: a land
// token sacrificed, or a card a replacement sent elsewhere, is not counted
// (the rulings). "Defending player" is the player behind the attack target,
// a planeswalker's controller included. Encore is the shared keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72a570a2-0a22-4680-ac88-a1724faf17de",
		Name:         "Eumidian Wastewaker",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Encore("{6}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Eumidian Wastewaker — you and defending player each discard a card or sacrifice a permanent; you draw a card for each land card put into a graveyard this way",
				wastewakerAttacks),
		},
	})
}

const (
	wastewakerDiscardOption   = "Discard a card"
	wastewakerSacrificeOption = "Sacrifice a permanent"
)

// wastewakerAttacks starts the chain with the controller, the defending
// player next.
func wastewakerAttacks(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var defender uuid.UUID
	if item.Trigger != nil {
		defender = b17DefendingPlayer(g, item.Trigger.Event)
	}
	if defender == uuid.Nil || defender == item.Controller || !defendingPlayerStillIn(g, defender) {
		defender = uuid.Nil
	}
	return wastewakerAsk(ctx, item.Controller, defender, nil)
}

// wastewakerAsk asks `who` discard-or-sacrifice, then `next` (when not
// nil), carrying the cards put into graveyards so far.
func wastewakerAsk(ctx *Context, who, next uuid.UUID, collected []uuid.UUID) error {
	g := ctx.Game
	p := g.PlayerByIDForEffect(who)
	if p == nil || p.Eliminated {
		return wastewakerNext(ctx, next, collected)
	}
	canDiscard := p.Hand != nil && p.Hand.Size() > 0
	canSacrifice := len(permanentsControlledByMatching(g, who, func(*game.Game, uuid.UUID, game.Card) bool { return true })) > 0
	carry := append([]uuid.UUID{who, next}, collected...)
	switch {
	case canDiscard && canSacrifice:
		return PickOption{
			Player:   who,
			Question: "Eumidian Wastewaker — discard a card or sacrifice a permanent",
			Options:  []game.ChoiceOption{{Label: wastewakerDiscardOption}, {Label: wastewakerSacrificeOption}},
			ThenKey:  wastewakerChoice,
			Carry:    carry,
		}.Apply(ctx)
	case canDiscard:
		return wastewakerDiscard(ctx, carry)
	case canSacrifice:
		return wastewakerSacrifice(ctx, carry)
	}
	return wastewakerNext(ctx, next, collected)
}

// The three keyed continuations of the chain refer to each other through
// wastewakerAsk, so they are registered from init rather than as
// package-level vars (an initialization cycle otherwise). On-disk keys:
// never renamed, never reused.
var (
	// wastewakerChoice runs the chosen half.
	wastewakerChoice game.OptionPickThen
	// wastewakerDiscarded and wastewakerSacrificed carry on to the next
	// player once the card or permanent is gone.
	wastewakerDiscarded  game.DiscardThen
	wastewakerSacrificed game.PromptRunThen
)

func init() {
	wastewakerChoice = OptionPickThen("option-pick/eumidian-wastewaker-choice", func(ctx *Context, r game.OptionPicked) error {
		if r.Option == nil || len(r.Carry) < 2 {
			return nil
		}
		if r.Option.Label == wastewakerSacrificeOption {
			return wastewakerSacrifice(ctx, r.Carry)
		}
		return wastewakerDiscard(ctx, r.Carry)
	})
	wastewakerDiscarded = game.RegisterDiscardThen("choose-cards/eumidian-wastewaker-discard", func(g *game.Game, r game.DiscardsDone) error {
		if r.Item == nil || len(r.Carry) < 2 {
			return nil
		}
		collected := append(append([]uuid.UUID(nil), r.Carry[2:]...), r.Discarded.Cards()...)
		return wastewakerNext(NewContext(g, r.Item), r.Carry[1], collected)
	})
	wastewakerSacrificed = RunThen("prompt-run/eumidian-wastewaker-sacrifice", func(ctx *Context, r game.RunDone) error {
		if len(r.Carry) < 2 {
			return nil
		}
		collected := append(append([]uuid.UUID(nil), r.Carry[2:]...), r.Sacrifices().Cards()...)
		return wastewakerNext(ctx, r.Carry[1], collected)
	})
}

func wastewakerDiscard(ctx *Context, carry []uuid.UUID) error {
	return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   carry[0],
		Source:   ctx.Source(),
		N:        1,
		Question: "Eumidian Wastewaker — discard a card",
		Asked:    game.NewAskedItem(ctx.Item),
	}, wastewakerDiscarded, carry)
}

func wastewakerSacrifice(ctx *Context, carry []uuid.UUID) error {
	return ctx.PlayersSacrificeThen([]uuid.UUID{carry[0]}, sacrificeSpec("a permanent"),
		"Eumidian Wastewaker — sacrifice a permanent", wastewakerSacrificed, ctx.RunData(carry))
}

// wastewakerNext asks the next player, or, with none left, draws a card for
// each land card now in a graveyard.
func wastewakerNext(ctx *Context, next uuid.UUID, collected []uuid.UUID) error {
	if next != uuid.Nil {
		return wastewakerAsk(ctx, next, uuid.Nil, collected)
	}
	lands := 0
	for _, id := range collected {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() && !c.IsToken() {
			lands++
		}
	}
	if lands == 0 {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: lands}.Apply(ctx)
}
