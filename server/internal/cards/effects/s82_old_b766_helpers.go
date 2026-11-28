package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b766_helpers.go — shared pieces of S82 OLD batch 766 (#5628):
// Tempest card draw and library.

// RecordTheEventsTargetAsThePlayer is the fill-in Build of a trigger whose
// effect acts on the player the event names as its target (Shocker's "that
// player"): it stamps that player onto Params.Player as the trigger goes on the
// stack, so the effect still has them if the source has left.
func RecordTheEventsTargetAsThePlayer(label string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Player = ev.Target
		return item
	}
}

// ThatPlayerDiscardsTheirHandThenDrawsThatMany is "that player discards all the
// cards in their hand, then draws that many cards" for the player a fill-in
// Build stamped onto Params.Player. The draw is a second step that reads the
// count from the step record, so a discard that stops for a prompt (a madness
// offer) draws as many as were really discarded when it resumes (CR 608.2c).
func ThatPlayerDiscardsTheirHandThenDrawsThatMany(record string) Effect {
	return Do(
		Instruction(func(ctx *Context) error {
			n, err := DiscardWholeHand(ctx.Game, ctx.Item.Params.Player)
			ctx.RecordStep(record, nil, n)
			return err
		}),
		Instruction(func(ctx *Context) error {
			return DrawCards{Player: ctx.Item.Params.Player, N: ctx.StepRecord(record).N()}.Apply(ctx)
		}),
	)
}

// BoobyTrapEnters is "As this artifact enters, choose an opponent and a card
// name other than a basic land card name" (Booby Trap): the controller picks
// both, the opponent stored as the chosen player and the name as the chosen
// name, the two prompts queued in a row.
func BoobyTrapEnters(label string) func(*game.Card, *Context) error {
	chooseOpponent := ChoosePlayerAsEnters(label, Opponents)
	return func(card *game.Card, ctx *Context) error {
		if err := chooseOpponent(card, ctx); err != nil {
			return err
		}
		ctx.Game.QueueCardNamePromptForEffect(game.CardNamePrompt{
			Chooser:          card.Controller,
			Source:           card.InstanceID,
			Reason:           label + " — choose a card name other than a basic land card name",
			NotBasicLandName: true,
		})
		return nil
	}
}

// boobyTrapDamage is "if you do, this artifact deals 10 damage to that player":
// it runs only when the artifact really was sacrificed, with the drawer carried
// from the trigger. The damage is dealt by the artifact as it last existed (CR
// 608.2h), and only to a player still in the game. An on-disk key: never
// renamed, never reused.
var boobyTrapDamage = CardTail("then/s82-old-b766-booby-trap-damage", func(ctx *Context, o game.TailOutcome) error {
	victim := TailCarry0(o)
	if !OneWentThisWay(o) || victim == uuid.Nil {
		return nil
	}
	if p := ctx.PlayerByID(victim); p == nil || p.Eliminated {
		return nil
	}
	d := DealDamage{Source: ctx.Source(), Target: victim, Amount: 10}
	if ref, ok := ctx.SourceRef(); ok {
		d.SourceObject = &ref
	}
	return d.Apply(ctx)
})

// BoobyTrapStrikes is the effect of Booby Trap's trigger: sacrifice this
// artifact, and if you do, it deals 10 damage to the player who drew the named
// card (stamped on the trigger as it went on the stack).
func BoobyTrapStrikes(g *game.Game, item *game.StackItem) error {
	if !OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	return SacrificePermanent{Target: item.SourceCardID, ThenTail: WithCarry(ctx.ItemTail(boobyTrapDamage), item.Params.Player)}.Apply(ctx)
}

// SacredGuideReveal is Sacred Guide's ability: reveal cards from the top of the
// library until a white card is revealed, put that card into the hand, and exile
// every other card revealed this way. No white card means every revealed card is
// exiled. The exile follows the move to hand (CR 608.2c), through a tail so a
// move that stops for a prompt is a restore point.
func SacredGuideReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := item.Controller
	white := ctx.TextColor("W")
	run, hit := RevealUntil(ctx, player, func(c game.Card) bool { return c.HasColor(white) },
		"Sacred Guide — reveal until a white card")
	var rest []uuid.UUID
	for _, id := range run {
		if id != hit {
			rest = append(rest, id)
		}
	}
	if hit == uuid.Nil {
		return sacredGuideExileRest(g, player, item.SourceCardID, rest)
	}
	return g.TakeFromLibraryToHandKeyedForEffect(player, []uuid.UUID{hit},
		game.KeyedTail{Ref: sacredGuideTaken, Player: player, Source: item.SourceCardID, Carry: rest})
}

// sacredGuideTaken exiles the other revealed cards (the Carry) once the white
// card's move to hand has settled. An on-disk key: never renamed, never reused.
var sacredGuideTaken = game.RegisterTail("s82-old-b766/sacred-guide-taken", func(g *game.Game, o game.TailOutcome) error {
	return sacredGuideExileRest(g, o.Tail.Player, o.Tail.Source, o.Tail.Carry)
})

func sacredGuideExileRest(g *game.Game, player, source uuid.UUID, rest []uuid.UUID) error {
	left := CardsStillInALibrary(g, rest)
	if len(left) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(left, game.KeyedTail{Player: player, Source: source})
}

// FromAnArtifactCreatureEnchantmentOrLandSource is "target activated ability
// from an artifact, creature, enchantment, or land" (Interdict): the ability's
// source is one of those four, read as the object it is, or last was. A mana
// ability never reaches the stack, which is the printed "(Mana abilities can't
// be targeted.)".
func FromAnArtifactCreatureEnchantmentOrLandSource() StackItemPredicate {
	return func(g *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		if item.Kind != game.StackItemActivated {
			return false
		}
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		return ok && (c.IsArtifact() || c.IsCreature() || c.IsEnchantment() || c.IsLand())
	}
}

// InterdictResolves is the whole of Interdict: counter the targeted activated
// ability, then its source permanent's activated abilities (mana abilities
// too) can't be activated this turn, then draw a card. The permanent is the one
// the ability came from; if it has left, or come back as a new object, it is
// not locked (CR 400.7). The target must still be legal for any of this to
// happen (CR 608.2b).
func InterdictResolves(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	ref, hasRef := ctx.Game.SourceObjectForEffect(ctx.Game.StackItemForEffect(t.ID))
	if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
		return err
	}
	if hasRef && OnBattlefield(ctx.Game, ref.ID) && ctx.Game.ObjectEpochForEffect(ref.ID) == ref.Epoch {
		if err := (ScopedEffectFor{
			Target:   ref.ID,
			Mods:     []game.Mod{game.AddRestrictionsMod(game.CantActivate | game.CantActivateMana)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Interdict — its activated abilities can't be activated this turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
}
