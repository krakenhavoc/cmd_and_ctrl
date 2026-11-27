package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b677_helpers.go — S82 OLD batch 677 (Nemesis: card draw and library,
// #5618).

// S82OldB677SearchForUpToThreeNamed is "you may search your library for up to
// three cards named <name>, reveal them, put them into your hand, then shuffle"
// (Nesting Wurm, Skyshroud Sentinel): the "you may" and the count are the
// search's own (CR 701.23b), none being a legal answer.
func S82OldB677SearchForUpToThreeNamed(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:    item.Controller,
			Predicate: func(c game.Card) bool { return c.Name == name },
			Dest:      game.ZoneHand,
			Limit:     3,
			Reveal:    true,
			Shuffle:   true,
			Optional:  true,
			Reason:    name + " — you may search for up to three cards named " + name,
		}.Apply(NewContext(g, item))
	}
}

// S82OldB677PackHunt is Pack Hunt's body: search for up to three cards with the
// same name as the target creature, read as the spell resolves (a target that
// is gone makes the spell fizzle, CR 608.2b, so there is no name to read).
func S82OldB677PackHunt(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	target, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	name := target.Effective().Name
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return game.CardNameMatches(c, name) },
		Dest:      game.ZoneHand,
		Limit:     3,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Pack Hunt — search for up to three cards named " + name,
	}.Apply(ctx)
}

// S82OldB677EyeOfYawgmoth is "Reveal a number of cards from the top of your
// library equal to the sacrificed creature's power. Put one into your hand and
// exile the rest": the power is the sacrificed creature's last-known, floored
// at zero (CR 107.1b), the reveal takes what the library has, and the
// controller picks the card that goes to hand.
func S82OldB677EyeOfYawgmoth(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var revealed []uuid.UUID
	if err := (RevealTopOfLibrary{N: ctx.SacrificedPower(), Reason: "Eye of Yawgmoth — reveal cards equal to the sacrificed creature's power", Revealed: &revealed}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	return RevealPick{
		Player:   ctx.Controller(),
		Question: "Eye of Yawgmoth — put one of the revealed cards into your hand (exile the rest)",
		Cards:    revealed,
		Min:      1,
		Max:      1,
		Keyed:    s82OldB677EyePick,
	}.Apply(ctx)
}

// s82OldB677EyePick takes the picked card and, once it has landed, exiles the
// rest. An on-disk key.
var s82OldB677EyePick = RegisterRevealPick("choose-cards/s82-old-b677-eye-of-yawgmoth-pick",
	func(ctx *Context, picked, left []uuid.UUID, _ []int) error {
		return ctx.Game.TakeFromLibraryToHandKeyedForEffect(ctx.Controller(), picked,
			game.KeyedTail{Ref: s82OldB677EyeTaken, Player: ctx.Controller(), Source: ctx.Source(), Carry: left})
	})

// s82OldB677EyeTaken exiles the unpicked revealed cards still in a library.
var s82OldB677EyeTaken = CardTail("s82-old-b677/eye-of-yawgmoth-taken", func(ctx *Context, o game.TailOutcome) error {
	left := CardsStillInALibrary(ctx.Game, o.Tail.Carry)
	if len(left) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(left, game.KeyedTail{})
})

// S82OldB677DiviningWitch is "Choose a card name. Exile the top six cards of
// your library, then reveal cards from the top of your library until you reveal
// a card with the chosen name. Put that card into your hand and exile all other
// cards revealed this way." The name is chosen as the ability resolves and is
// not checked against any list; the six exiled cards are not "revealed this
// way", so they are exiled and not looked for.
func S82OldB677DiviningWitch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
		"Divining Witch — choose a card name", ctx.ItemTail(s82OldB677WitchNamed))
	return nil
}

// s82OldB677WitchNamed exiles the top six once the name (the tail's Word) is
// chosen. An on-disk key.
var s82OldB677WitchNamed = CardTail("then/s82-old-b677-divining-witch-named", func(ctx *Context, o game.TailOutcome) error {
	player := o.Tail.Player
	return ctx.Game.ExileTopKeyedForEffect(player, player, 6,
		game.KeyedTail{Ref: s82OldB677WitchExiled, Player: player, Source: ctx.Source(), Word: o.Tail.Word})
})

// s82OldB677WitchExiled reveals from the top until the chosen name, takes that
// card, and exiles the other cards revealed. An on-disk key.
var s82OldB677WitchExiled = CardTail("s82-old-b677/divining-witch-exiled", func(ctx *Context, o game.TailOutcome) error {
	player, name := o.Tail.Player, o.Tail.Word
	run, hit := RevealUntil(ctx, player, func(c game.Card) bool {
		return game.CardNameMatches(c, name)
	}, "Divining Witch — reveal until a card named "+name)
	var rest []uuid.UUID
	for _, id := range run {
		if id != hit {
			rest = append(rest, id)
		}
	}
	if hit == uuid.Nil {
		return exileStillInALibraryB677(ctx.Game, rest)
	}
	return ctx.Game.TakeFromLibraryToHandKeyedForEffect(player, []uuid.UUID{hit},
		game.KeyedTail{Ref: s82OldB677WitchTaken, Player: player, Source: ctx.Source(), Carry: rest})
})

var s82OldB677WitchTaken = CardTail("s82-old-b677/divining-witch-taken", func(ctx *Context, o game.TailOutcome) error {
	return exileStillInALibraryB677(ctx.Game, o.Tail.Carry)
})

func exileStillInALibraryB677(g *game.Game, ids []uuid.UUID) error {
	left := CardsStillInALibrary(g, ids)
	if len(left) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(left, game.KeyedTail{})
}

// S82OldB677DefiantVanguardBlocks is Defiant Vanguard's trigger: "When this
// creature blocks, at end of combat, destroy it and all creatures it blocked
// this turn." The creatures it blocked are those it blocks as the trigger is
// put on the stack, each with its CR 400.7 epoch, and the Vanguard itself is
// the first object; an object that has left and returned by end of combat is
// a new one and is not destroyed.
func S82OldB677DefiantVanguardBlocks(label string) game.TriggeredAbility {
	t := On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return SelfBlocksOnce(ev, source)
	}, label, func(g *game.Game, item *game.StackItem) error {
		return ScheduleDelayedTrigger{
			At:     game.StepEndCombat,
			Label:  label,
			Body:   s82OldB677DestroyListedBody,
			Params: game.EffectParams{Objects: item.Params.Objects},
		}.Apply(NewContext(g, item))
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		ids := append([]uuid.UUID{source.InstanceID}, source.BlockedAttackers()...)
		for _, id := range ids {
			item.Params.Objects = append(item.Params.Objects, game.ObjectRef{ID: id, Epoch: g.ObjectEpochForEffect(id)})
		}
		return item
	}
	return t
}

// s82OldB677DestroyListedBody destroys every listed object that is still on
// the battlefield as the same object (CR 400.7).
var s82OldB677DestroyListedBody = game.DelayedBody("s82-old-b677/destroy-listed-objects",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		for _, ref := range p.Objects {
			if !OnBattlefield(g, ref.ID) || g.ObjectEpochForEffect(ref.ID) != ref.Epoch {
				continue
			}
			if err := (DestroyTarget{Target: ref.ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
