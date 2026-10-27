package effects

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_choice.go — the two card-facing shapes of "an OPPONENT
// chooses while this resolves" (#568, CR 608.2).
//
// MayChoice (may_choice.go) already addresses a yes/no to any seat,
// which covers Combustible Gearhulk, Charismatic Conqueror and Painful
// Quandary. What it cannot say is a choice among three consequences,
// or a choice made in two steps by two different players. Those are
// the two shapes here, and both are built on game.PendingChoiceOptionPick
// plus #552's chaining rather than on kinds of their own:
//
//   - PickOption — "choose one of the following", addressed to an
//     opponent. Torment of Hailfire's "loses 3 life unless that player
//     sacrifices a nonland permanent of their choice or discards a
//     card".
//   - PileSplit — "an opponent separates those cards into two piles.
//     Put one pile into your hand and the other into your graveyard."
//     Fact or Fiction.
//
// # The list is built from what the chooser can DO
//
// CR 608.2's "as much as possible" is enforced at queue time, not at
// answer time: an option the chooser cannot take is never offered. A
// player with no nonland permanent is not shown "sacrifice a nonland
// permanent"; a player with an empty hand is not shown "discard a
// card". That is the rule the engine's own contract depends on — the
// enumerator marks the FIRST option always-legal, so the first option
// must be one that always works.

// PickOption asks Player to choose one of Options while an effect
// resolves, and runs the registered continuation ThenKey with the index
// they picked.
//
// Anything printed after the choice goes in ThenKey, not after this
// primitive returns: Apply only queues the prompt.
//
// A prompt with no options queues nothing and runs ThenKey with
// game.NoChoiceIndex, so a card whose whole option list turned out to be
// impossible still gets to finish. A chooser who has left the game
// (CR 800.4a) is the same case.
type PickOption struct {
	// Player is asked. Zero means the resolving effect's controller.
	Player uuid.UUID

	// From owns the material the options are about, when that is not
	// the chooser — the redaction pass reads it to decide what the
	// chooser may see of a pile that is not theirs. Zero means the
	// chooser's own.
	From uuid.UUID

	// Question is the prompt's header, written the way the card is.
	Question string

	// Options are the branches in the card's printed order. The FIRST
	// must be one the chooser can always take.
	Options []game.ChoiceOption

	// ThenKey is everything printed after the choice, a registered
	// continuation (OptionPickThen), with Carry the IDs it is handed
	// (#2854). It receives the index of the chosen option, or
	// game.NoChoiceIndex when no question could be asked; it runs with
	// g.mu held and may queue further choices. An option that costs mana
	// (ChoiceOption.ManaCost) and that the chooser cannot pay is not
	// offered, and ThenKey still receives the index the card printed. The
	// prompt holds no closure, so the table stays a restore point while
	// it waits. Required (#4253 retired the closure form).
	ThenKey game.OptionPickThen
	Carry   []uuid.UUID
	// WithItem hands the keyed continuation a Context rebuilt from the
	// resolving item (carried as data, game.AskedItem) instead of the
	// stand-in built from the controller and source: for a continuation
	// that reads the item's own words, targets or source stamp (#4253
	// slice 8). Ignored without ThenKey.
	WithItem bool
	// Words and Nums are a keyed continuation's other plain data (#4253
	// slice 10c): the walk of lands a player activates mana abilities of
	// keeps its place in them. Handed back as OptionPicked.Words and Nums.
	Words []string
	Nums  []int

	// Tail is a registered tail the keyed answer hands on
	// (game.OptionPicked.Tail): the rest of the card after a chain of
	// picks, as data. Read only with ThenKey. (#4253 sweep, second pass.)
	Tail game.KeyedTail
}

// OptionPickThen registers a keyed option pick's continuation, as a
// card file's package-level var (#2854). The body is handed a Context
// rebuilt from values — the controller of the effect that asked as its
// controller, the card that asked as its source — never from the stack
// item it began with, which has long resolved; and the answer, with
// Carry as the card put it on the prompt. r.Option is nil, and r.Index
// is -1, when nobody chose (the chooser left the game). The key is an
// on-disk identity ("option-pick/<card>-<what>"): never renamed, never
// reused.
func OptionPickThen(key string, body func(ctx *Context, r game.OptionPicked) error) game.OptionPickThen {
	return game.RegisterOptionPickThen(key, func(g *game.Game, r game.OptionPicked) error {
		controller := r.Chooser
		if len(r.Carry) > 0 {
			// PickOption.Apply puts the asking effect's controller
			// first; the card's own Carry follows it.
			controller, r.Carry = r.Carry[0], r.Carry[1:]
		}
		if r.Asked != nil {
			return body(NewContext(g, r.Asked), r)
		}
		return body(NewContext(g, &game.StackItem{
			Kind:         game.StackItemSpell,
			Controller:   controller,
			Owner:        controller,
			SourceCardID: r.Source,
		}), r)
	})
}

func (p PickOption) Apply(ctx *Context) error {
	chooser := p.Player
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	if p.ThenKey.Key() == "" {
		return fmt.Errorf("%w: PickOption with no registered continuation (OptionPickThen)", game.ErrInvalidParam)
	}
	carry := append([]uuid.UUID{ctx.Controller()}, p.Carry...)
	var asked *game.AskedItem
	if p.WithItem {
		asked = game.NewAskedItem(ctx.Item)
	}
	queued := ctx.Game.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:    chooser,
		FromPlayer: p.From,
		Source:     ctx.Source(),
		Question:   p.Question,
		Options:    p.Options,
		ThenKey:    p.ThenKey,
		Carry:      carry,
		Asked:      asked,
		Words:      p.Words,
		Nums:       p.Nums,
		Tail:       p.Tail,
	})
	if queued != uuid.Nil {
		return nil
	}
	return ctx.Game.RunOptionPickThenForEffect(p.ThenKey, game.OptionPicked{
		Chooser: chooser, Source: ctx.Source(), Index: game.NoChoiceIndex, Carry: carry, Asked: asked.StackItem(),
		Words: append([]string(nil), p.Words...), Nums: append([]int(nil), p.Nums...), Tail: p.Tail,
	})
}

// PileSplit is "an opponent separates those cards into two piles; you
// take one of them" — two chained prompts to two different seats.
//
// The cards must already be REVEALED (RevealTopOfLibrary, or
// RevealCards) before this is called. The splitter is being asked
// about cards they do not own, and protocol's redaction pass shows
// them only what they are entitled to see: an unrevealed pool would
// reach them as an empty prompt. That is deliberate — the alternative
// is an engine that hands one seat a handle on another's library.
type PileSplit struct {
	// Splitter separates the cards. Zero is not meaningful; every
	// printed card of this family names an opponent.
	Splitter uuid.UUID

	// Chooser takes one of the two piles. Zero means the resolving
	// effect's controller, which is every printed card of this family.
	Chooser uuid.UUID

	// Owner owns the cards. Zero means the Chooser — Fact or Fiction
	// splits the controller's own library.
	Owner uuid.UUID

	// SplitQuestion / PickQuestion are the two prompts' headers.
	SplitQuestion, PickQuestion string

	// Cards are the revealed cards being separated, in the order the
	// table saw them.
	Cards []uuid.UUID

	// ThenTail is everything printed after the split, a tail registered
	// with PileTail (#4253 slice 10b): it is told the pile the chooser
	// TOOK and the pile they left, in that order, and runs with g.mu
	// held. Both prompts are keyed, so a table waiting on either is a
	// restore point.
	ThenTail game.KeyedTail

	// FaceDown is the "face-down pile and a face-up pile" split
	// (#2147): the cards were LOOKED at, not revealed, so the caller
	// must have made the splitter the only knower
	// (LookAtTopOfPlayersLibraryForEffect). The splitter's picked cards
	// are the face-down pile; the rest are revealed when the split is
	// answered. See game.PileSplitPrompt.FaceDown.
	FaceDown bool
}

func (p PileSplit) Apply(ctx *Context) error {
	chooser := p.Chooser
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	if p.ThenTail.IsZero() {
		return fmt.Errorf("%w: PileSplit with no registered tail (PileTail)", game.ErrInvalidParam)
	}
	ctx.Game.QueuePileSplitForEffect(game.PileSplitPrompt{
		Splitter:      p.Splitter,
		Chooser:       chooser,
		Owner:         p.Owner,
		Source:        ctx.Source(),
		SplitQuestion: p.SplitQuestion,
		PickQuestion:  p.PickQuestion,
		Cards:         p.Cards,
		FaceDown:      p.FaceDown,
		ThenTail:      p.ThenTail,
	})
	return nil
}

// PileTail registers a pile split's keyed "then" (PileSplit.ThenTail): the
// body is told the pile taken and the pile left, each in the order the
// table saw them, and runs against the Context of CardTail. The key is an
// on-disk identity (ledgered "tail <key>"): never renamed, never reused.
// (#4253 slice 10b.)
func PileTail(key string, body func(ctx *Context, taken, left []uuid.UUID, o game.TailOutcome) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		n := min(max(o.Amount, 0), len(o.IDs))
		return body(ctx, o.IDs[:n:n], o.IDs[n:], o)
	})
}

// RevealPick is "an opponent chooses N of those cards" — Intuition's
// one, Gifts Ungiven's two (#1214, CR 608.2 / CR 701.20).
//
// The cards must already be REVEALED, for PileSplit's reason and with
// the same consequence: the chooser is being asked about cards they do
// not own, and protocol's redaction pass shows them exactly what they
// have been made a knower of. An unrevealed pool reaches them as an
// empty prompt.
//
// `Keyed` (RegisterRevealPick) receives the cards they PICKED and the
// cards they LEFT, each in the order the set was revealed, because every
// card of this family does something with both halves. It runs even when
// nothing could be asked — an empty set, or a chooser who has left the
// game — with nothing picked and everything left. It is a registered
// continuation (#4253 sweep, second pass) and the asking item rides the
// prompt as data, so a table waiting on the opponent's pick is a restore
// point.
type RevealPick struct {
	// Player picks. Zero is not meaningful; every printed card of this
	// family names an opponent.
	Player uuid.UUID

	// Owner owns the cards. Zero means the resolving effect's
	// controller, which is every printed card of this family.
	Owner uuid.UUID

	// Question is the prompt's header, written the way the card is.
	Question string

	// Cards are the revealed cards, in the order the table saw them.
	Cards []uuid.UUID

	// Min / Max bound the pick. Max <= 0 means all of them. Both are
	// clamped to the set's size (CR 608.2's "as much as possible"), so
	// "chooses two" over a set of one is a choice of one.
	Min, Max int

	// Keyed is everything printed after the choice, registered with
	// RegisterRevealPick; Nums are numbers it reads, fixed as the effect
	// resolved (Thassa's Oracle's X). Runs with g.mu held, against the
	// asking item.
	Keyed RevealPickThen
	Nums  []int
}

// RevealPickThen is a reveal pick's registered continuation
// (RegisterRevealPick).
type RevealPickThen struct {
	then game.ChooseCardsThen
	body func(ctx *Context, picked, left []uuid.UUID, nums []int) error
}

// RegisterRevealPick registers RevealPick's continuation under key, an
// on-disk identity ("choose-cards/<card>-<what>", ledgered "choose <key>"):
// never renamed, never reused. `body` is handed the cards picked and left,
// each in the order the set was revealed, and RevealPick.Nums. Call it
// once, from a package-level var. It must capture nothing.
func RegisterRevealPick(key string, body func(ctx *Context, picked, left []uuid.UUID, nums []int) error) RevealPickThen {
	then := game.RegisterChooseCardsThen(key, func(g *game.Game, r game.CardsChosen) error {
		ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
		picked, left := revealPartition(r.Candidates, r.Picked)
		return body(ctx, picked, left, r.Nums)
	})
	return RevealPickThen{then: then, body: body}
}

// revealPartition splits the revealed set into the cards picked and the
// cards left, each in the order the set was revealed.
func revealPartition(all, picked []uuid.UUID) (in, out []uuid.UUID) {
	for _, id := range all {
		if slices.Contains(picked, id) {
			in = append(in, id)
		} else {
			out = append(out, id)
		}
	}
	return in, out
}

func (p RevealPick) Apply(ctx *Context) error {
	if p.Keyed.body == nil {
		return fmt.Errorf("%w: RevealPick with no registered continuation (RegisterRevealPick)", game.ErrInvalidParam)
	}
	owner := p.Owner
	if owner == uuid.Nil {
		owner = ctx.Controller()
	}
	queued, err := ctx.Game.RevealPickThenForEffect(game.RevealPickPrompt{
		Chooser:  p.Player,
		Owner:    owner,
		Source:   ctx.Source(),
		Question: p.Question,
		Cards:    p.Cards,
		Min:      p.Min,
		Max:      p.Max,
		ThenKey:  p.Keyed.then,
		Asked:    game.NewAskedItem(ctx.Item),
		Nums:     p.Nums,
	}, nil)
	if err != nil || queued > 0 {
		return err
	}
	// Nobody could be asked: nothing is picked and every card is left.
	return p.Keyed.body(ctx, nil, append([]uuid.UUID(nil), p.Cards...), append([]int(nil), p.Nums...))
}

// ChoosePermanents is the resolution-time pick over one or more
// players' boards (#1214, CR 608.2): "for each player, you choose from
// among the permanents that player controls …" (Tragic Arrogance) and
// "sacrifice any number of lands" (Scapeshift).
//
// ONE primitive for both engine kinds, because the card does not know
// which one it is asking: the engine decides per leg, from whether the
// board being walked belongs to the chooser
// (game.PermanentsPickedThenForEffect). Tragic Arrogance walks every
// player including its own controller, so one printed sentence is both.
//
// It does NOT target. A permanent with hexproof, shroud or protection
// can be chosen this way, and the effect needs no legal choice to
// resolve — which is the difference from writing the same clause as a
// target spec, and the reason these cards were waiting on a prompt
// rather than on a predicate.
type ChoosePermanents struct {
	// Player answers every leg. Zero means the resolving effect's
	// controller.
	Player uuid.UUID

	// Of are the players whose boards are walked, in ask order. Empty
	// means the chooser's own board alone, which is the "choose N of
	// your own permanents" shape.
	Of []uuid.UUID

	// Question is each prompt's header.
	Question string

	// Candidates computes ONE seat's offer — the permanents on offer
	// and the bounds. Returning nothing skips that seat. Runs with
	// g.mu held and must not write.
	Candidates func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int)

	// Keyed is the choice's set rule, its sacrifice and everything printed
	// after it, as a REGISTERED definition (RegisterPermanentsPick, #4290):
	// PermanentsDef.Validate is #1017's set-level legality hook, applied to
	// every leg (Tragic Arrogance's "an artifact, a creature, an
	// enchantment, and a planeswalker" is a rule about the SET);
	// PermanentsDef.Sacrifice sacrifices everything chosen, across every
	// leg, before Then runs ("target opponent sacrifices the creature you
	// choose"; the permanents leave through the one sacrifice path, so a
	// sacrificed commander goes to the graveyard like any other permanent,
	// CR 903.9a, ADR 0115, and Then really does run after they have gone);
	// and Then is the rest. The legs hold no closure, so a table waiting on
	// any seat's pick is a restore point.
	//
	// Carry and Nums are the plain data its body reads beside the
	// resolving item, which rides the run (Context.RunData).
	Keyed PermanentsThen
	Carry []uuid.UUID
	Nums  []int
}

func (p ChoosePermanents) Apply(ctx *Context) error {
	chooser := p.Player
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	of := p.Of
	if len(of) == 0 {
		of = []uuid.UUID{chooser}
	}
	if !p.Keyed.run.Valid() {
		return fmt.Errorf("%w: ChoosePermanents with no registered pick (RegisterPermanentsPick)", game.ErrInvalidParam)
	}
	_, err := ctx.Game.PermanentsPickedThenKeyedForEffect(game.PermanentPickPrompt{
		Chooser:    chooser,
		Source:     ctx.Source(),
		Question:   p.Question,
		Of:         of,
		Candidates: p.Candidates,
	}, p.Keyed.run, ctx.RunData(p.Carry, p.Nums...))
	return err
}

// NonlandPermanentsControlledBy lists the nonland permanents a player
// controls, in battlefield order — the candidate set behind "sacrifices
// a nonland permanent of their choice".
//
// Here rather than in a card file because it is the shape of the
// clause, not of the card: an option list has to be built from what the
// chooser can actually do (see the file comment), and every card that
// offers a sacrifice branch needs the same answer.
//
// Caller must hold g.mu — it is an effect-time read.
func NonlandPermanentsControlledBy(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller != playerID || c.IsLand() {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// SacrificeChoice queues "sacrifice one of these permanents" as a
// chained card pick, and runs ThenTail once the permanent has actually
// gone.
//
// game.PlayerSacrificesForEffect is the prompt for a sacrifice that is
// the END of an effect (Grave Pact fans one out per player and nothing
// follows). This one is for a sacrifice in the MIDDLE of one — Torment
// of Hailfire repeats X times, and the next repetition must not be
// asked until this one has finished — so it needs a continuation, and
// it gets one the way QueueDiscardChoiceForEffect does: by riding
// PendingChoiceChooseCards, whose keyed frame carries a tail and
// re-checks the live zone on submit.
//
// It is not a second sacrifice system. The permanent leaves through
// SacrificePermanentForEffect, the one sacrifice path, so EventSacrifice
// and the dies-triggers are identical either way; what differs is only
// which prompt kind asked, and the CHOOSE-CARDS kind is the one with a
// continuation.
//
// A player with no candidate permanent is not prompted and ThenTail runs
// immediately — a mandatory sacrifice with nothing to sacrifice does
// nothing (CR 701.21a), and the rest of the effect still happens.
type SacrificeChoice struct {
	// Player sacrifices. Required.
	Player uuid.UUID

	// Candidates are the permanents they may choose from, computed by
	// the caller (NonlandPermanentsControlledBy and friends) so the
	// clause's predicate stays in the card file.
	Candidates []uuid.UUID

	// Question is the prompt's header.
	Question string

	// ThenTail is the rest of the effect, a registered tail (#4253 slice
	// 10b). It runs once the permanent has gone, or immediately when
	// there was nothing to sacrifice. The choice is a keyed pick
	// (sacrificeChoicePicked) carrying the tail, so a table waiting on it
	// is a restore point. It still asks when there is one candidate. Zero
	// means nothing follows.
	ThenTail game.KeyedTail
}

// sacrificeChoicePicked is the keyed answer to a SacrificeChoice: the picked
// permanent goes through the sacrifice's continuation, and the prompt's tail
// runs once it has gone. The answer is not read for "if you do": the prompt
// is mandatory. An on-disk key: never renamed, never reused.
var sacrificeChoicePicked = game.RegisterChooseCardsThen("choose-cards/sacrifice-choice", func(g *game.Game, r game.CardsChosen) error {
	// The prompt's tail rides the sacrifice as the Next of a keyed tail
	// (#4253), told nothing, as it always was.
	tail := r.Tail
	return g.SacrificeAllKeyedForEffect(uuid.Nil, r.Picked, game.KeyedTail{Ref: sacrificeChoiceDone, Next: &tail})
})

// sacrificeChoiceDone is the sacrifice settled: the choice's own tail runs,
// told nothing. An on-disk key: never renamed, never reused.
var sacrificeChoiceDone = game.RegisterTail("then/sacrifice-choice-done", func(g *game.Game, o game.TailOutcome) error {
	if o.Tail.Next == nil {
		return nil
	}
	return o.Tail.Next.Run(g, nil)
})

func (s SacrificeChoice) Apply(ctx *Context) error {
	// #3275, CR 101.2: a permanent that can't be sacrificed is not a
	// legal choice.
	var candidates []uuid.UUID
	for _, id := range s.Candidates {
		if !ctx.Game.CantBeSacrificedForEffect(id) {
			candidates = append(candidates, id)
		}
	}
	s.Candidates = candidates
	if len(s.Candidates) == 0 {
		return s.ThenTail.Run(ctx.Game, nil)
	}
	// Re-checked against the live battlefield on submit: the prompt is
	// asynchronous and a permanent can leave between the question and
	// the answer. #993: the picked permanent goes through the
	// sacrifice's CONTINUATION (sacrificeChoicePicked), so the tail
	// really does run "once the permanent has gone" — Chain of Vapor
	// once offered its copy while a sacrificed land's owner was still
	// mid-prompt. The answer is not read: the prompt is mandatory (Min
	// 1), so a clause hanging off it is "then", not "if you do".
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    s.Player,
		FromPlayer: s.Player,
		Source:     ctx.Source(),
		Question:   s.Question,
		Cards:      s.Candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    sacrificeChoicePicked,
		Tail:       s.ThenTail,
	})
	return nil
}

// PermanentsControlledBy lists every permanent a player controls, in
// battlefield order — the candidate set behind a bare "sacrifice a
// permanent" (Read the Runes), which unlike the nonland clause above
// includes lands.
//
// Caller must hold g.mu — it is an effect-time read.
func PermanentsControlledBy(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == playerID {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// tapOrUntapPermanent is "tap or untap <that permanent>" once the
// permanent is known: the choice between the two actions is made as
// the effect resolves, through an option pick. Either branch is always
// legal — tapping a tapped permanent and untapping an untapped one
// both do nothing — so the first option is one the chooser can always
// take, as the prompt requires. Shared by Merrow Reejerey's trigger
// and Gandalf the Grey's first bullet.
func tapOrUntapPermanent(ctx *Context, cardName string, target uuid.UUID) error {
	return PickOption{
		Question: cardName + " — tap or untap that permanent",
		Options: []game.ChoiceOption{
			{Label: "Tap it"},
			{Label: "Untap it"},
		},
		ThenKey: tapOrUntapAnswer,
		Carry:   []uuid.UUID{target},
	}.Apply(ctx)
}

// tapOrUntapAnswer is the tap-or-untap choice answered, keyed (#4253): the
// permanent rides the prompt as its one carried ID.
var tapOrUntapAnswer = OptionPickThen("option-pick/tap-or-untap-that-permanent", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	target := r.Carry[0]
	switch r.Index {
	case 0:
		return TapTarget{Target: target}.Apply(ctx)
	case 1:
		return UntapTarget{Target: target}.Apply(ctx)
	default:
		return nil
	}
})
