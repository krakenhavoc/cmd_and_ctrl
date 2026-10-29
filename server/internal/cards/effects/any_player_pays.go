package effects

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// AnyPlayerPays is "<effect> unless any player pays <Cost>" and "Any
// player may pay <Cost>. If no one does, <effect>" (ADR 0150, #1903):
// the payment is offered to every player still in the game, in APNAP
// order from the active player, the controller included, until one
// pays (game.QueueAnyPlayerPaysKeyedForEffect). The offer blocks the step
// it was made in and takes no standing answer.
//
// The consequences are registered (Then) and run against the resolving
// stack item carried as data, so "if no one does" reads the item's
// targets and X when it finally runs (Wild Might's creature, Soul
// Strings' cards).
type AnyPlayerPays struct {
	// Cost is the payment ("{2}"), with any X already worked out.
	Cost string
	// Question is the prompt header each player sees.
	Question string
	// HelpsController marks a payment the controller wants made (Icy
	// Prison's upkeep): game.AnyPlayerPaysKeyed.HelpsController.
	HelpsController bool

	// The fields below are ADR 0150 PR 2's (§3): payments other than
	// mana, and the "any opponent" and "counter it" shapes.

	// Pay is a payment that is not mana: LifePayment, DamagePayment,
	// SacrificePayment, DiscardPayment, ExileFromGraveyardPayment,
	// EitherPayment (ADR 0187 §5) or NoPayment. Cost is unused
	// when it is set; the offer's words are the payment's, lower-cased
	// ("sacrifice a land", "5 life").
	Pay UpkeepPayment
	// Opponents is "any opponent may" (Innocent Traveler, Distant
	// Memories): the controller is not asked.
	Opponents bool
	// GuardsSource holds the source spell on the stack while the offer
	// is open: a cast trigger whose "if a player does" counters the
	// spell (Dash Hopes, Brain Gorgers).
	GuardsSource bool
	// AcceptLabel and DeclineLabel name the two answers when "Pay" and
	// "Don't pay" read wrong.
	AcceptLabel, DeclineLabel string

	// Then is the offer's REGISTERED consequences (AnyPlayerConsequences,
	// #4264): "if a player does", "if no one does" and the payment worked
	// out per player (Temporal Extortion's half their life, rounded up).
	// Carry is the plain data they are handed. Every link of the offer
	// holds no closure, so a table waiting on any player's answer is a
	// restore point. Required (#4253 retired the closure form).
	Then  game.AnyPlayerThen
	Carry []uuid.UUID
}

// AnyPlayerConsequences registers a keyed any-player offer's
// consequences under `key`, an on-disk identity (never renamed, never
// reused, ledgered "anyplayer <key>"): `onPaid` is "if a player does,
// ...", `onNobody` "if no one does, ..." and `perPayer` the payment
// worked out for each player. Any may be nil, but not all three. The
// bodies run against a Context bound to a copy of the asking item rebuilt
// from data the offer carries (game.AskedItem), because the item itself
// has long resolved by the time anyone answers (and a restore does not
// carry it). An offer queued without one gets a stand-in with the
// effect's controller and its source card. Call it once, from a
// package-level var.
func AnyPlayerConsequences(key string,
	onPaid, onNobody func(ctx *Context, r game.AnyPlayerAnswer) error,
	perPayer game.AnyPlayerPerPayerFunc,
) game.AnyPlayerThen {
	wrap := func(body func(ctx *Context, r game.AnyPlayerAnswer) error) game.AnyPlayerFunc {
		if body == nil {
			return nil
		}
		return func(g *game.Game, r game.AnyPlayerAnswer) error {
			if r.Item != nil {
				return body(NewContext(g, r.Item), r)
			}
			return body(keyedContext(g, r.Controller, r.Source), r)
		}
	}
	return game.RegisterAnyPlayerPays(key, wrap(onPaid), wrap(onNobody), perPayer)
}

// LifePayment is "pay N life" as a payment (ADR 0150 §3, CR 119.4).
func LifePayment(n int) UpkeepPayment {
	a := game.PayAction{Kind: game.PayActionLife, Count: n}
	return UpkeepPayment{Action: &a, Words: game.PayActionWords(a)}
}

// DamagePayment is "have <source> deal N damage to them" as a payment
// (ADR 0150 §3): the offer's source deals it to the payer.
func DamagePayment(n int) UpkeepPayment {
	a := game.PayAction{Kind: game.PayActionDamage, Count: n}
	return UpkeepPayment{Action: &a, Words: game.PayActionWords(a)}
}

// NoPayment is an offer whose "yes" costs nothing (ADR 0150 §3:
// Distant Memories).
func NoPayment() UpkeepPayment {
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionNothing}}
}

// ExileFromGraveyardPayment is "exile N cards from your graveyard" as a
// payment (ADR 0187 §3): Carrion Rats' "any player may exile a card
// from their graveyard", Carrion Wurm's three. Each payer exiles from
// their OWN graveyard, and a payer with fewer than n cards can't pay
// (CR 118.3).
func ExileFromGraveyardPayment(n int) UpkeepPayment {
	words := "Exile a card from your graveyard"
	if n != 1 {
		words = "Exile " + numberWord(n) + " cards from your graveyard"
	}
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionExileFromGraveyard, Count: n}, Words: words}
}

// EitherPayment is "<a> or <b>" as one payment (ADR 0187 §5): Worms of
// the Earth's "any player may sacrifice two lands of their choice or
// have this enchantment deal 5 damage to that player". It is ONE offer
// with two ways to pay, not two offers: "if a player does either" is a
// single condition, so each player is asked once and their answer names
// the branch (`pay_option`).
//
// Both arguments are the ordinary non-mana payments (SacrificePayment,
// DamagePayment, LifePayment, DiscardPayment, ExileFromGraveyardPayment).
// A mana payment or another EitherPayment has no branch shape and
// panics: a programming error in the card file, caught by the first test
// that resolves the card.
func EitherPayment(a, b UpkeepPayment) UpkeepPayment {
	branches := make([]game.PayAction, 0, 2)
	phrases := make([]string, 0, 2)
	for _, p := range []UpkeepPayment{a, b} {
		if p.Action == nil || p.Action.Kind == game.PayActionEither {
			panic("effects.EitherPayment: each branch must be a non-mana payment that is not itself an either")
		}
		branch := *p.Action
		branch.Words = eitherPhrase(p)
		branches = append(branches, branch)
		phrases = append(phrases, branch.Words)
	}
	words := phrases[0] + " or " + phrases[1]
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionEither, Either: branches},
		Words:  strings.ToUpper(words[:1]) + words[1:],
	}
}

// eitherPhrase is how a branch of an either payment reads as a verb
// phrase after "may": "sacrifice two lands", "take 5 damage", "pay 5
// life". A damage or life payment carries only its amount ("5 damage"),
// so it needs the verb; a card payment's words already are one.
func eitherPhrase(p UpkeepPayment) string {
	switch p.Action.Kind {
	case game.PayActionDamage:
		return "take " + p.Words
	case game.PayActionLife:
		return "pay " + p.Words
	}
	if p.Words == "" {
		return game.PayActionPhrase(*p.Action)
	}
	return strings.ToLower(p.Words[:1]) + p.Words[1:]
}

// HalfTheirLifeRoundedUp is a PerPayer for "pay half their life,
// rounded up" (Temporal Extortion): each player's own half, read as
// they are asked.
func HalfTheirLifeRoundedUp(g *game.Game, payer uuid.UUID) (game.PayAction, string) {
	life := 0
	if p := g.PlayerByIDForEffect(payer); p != nil {
		life = p.Life
	}
	a := game.PayAction{Kind: game.PayActionLife, Count: game.HalfLifeRoundedUp(life)}
	return a, game.PayActionWords(a)
}

func (a AnyPlayerPays) Apply(ctx *Context) error {
	if a.Then.Key() == "" {
		return fmt.Errorf("%w: AnyPlayerPays with no registered consequences (AnyPlayerConsequences)", game.ErrInvalidParam)
	}
	var guards uuid.UUID
	if a.GuardsSource {
		guards = ctx.Source()
	}
	cost, action := a.Cost, a.Pay.Action
	if action != nil {
		// "Sacrifice a land" is the upkeep prompt's words; in an
		// offer it follows "any player may" (ADR 0150 §5).
		cost = a.Pay.Words
		if cost != "" {
			cost = strings.ToLower(cost[:1]) + cost[1:]
		}
	}
	return ctx.Game.QueueAnyPlayerPaysKeyedForEffect(game.AnyPlayerPaysKeyed{
		Source:       ctx.Source(),
		Controller:   ctx.Controller(),
		Cost:         cost,
		Action:       action,
		Question:     a.Question,
		Opponents:    a.Opponents,
		Guards:       guards,
		AcceptLabel:  a.AcceptLabel,
		DeclineLabel: a.DeclineLabel,
		Then:         a.Then,
		Carry:        a.Carry,
		Asked:        game.NewAskedItem(ctx.Item),

		HelpsController: a.HelpsController,
	})
}

// counterThisSpell counters the item's source spell if it is still on
// the stack: the "counter <this spell>" of a cast trigger (Dash Hopes).
// A spell that has already left has nothing to counter (CR 608.2b).
func counterThisSpell(ctx *Context) error {
	id := ctx.Source()
	if it := ctx.Game.StackItemForEffect(id); it == nil || it.Kind != game.StackItemSpell {
		return nil
	}
	return CounterTarget{StackID: id}.Apply(ctx)
}

// anyPlayerMayCounterThisSpell is "When you cast this spell, any player
// may <pay>. If a player does, counter <this spell>." (ADR 0150 §2:
// Dash Hopes, Temporal Extortion, Brain Gorgers, Phantasmagorian). The
// trigger resolves above the spell (CR 601.2i), and each link of the
// offer guards the spell, so it can't resolve while a player is still
// deciding. `pay` is the payment, or the zero value with the per-payer
// key (anyPlayerCounterThisSpellHalfLife) for one worked out per player.
//
// The consequence is a registered key (#4264), so a table waiting on any
// player's answer is a restore point.
func anyPlayerMayCounterThisSpell(name, what string, pay UpkeepPayment, then game.AnyPlayerThen) game.TriggeredAbility {
	return WhenYouCastThisSpell(name+" — any player may "+what+" to counter it",
		func(g *game.Game, item *game.StackItem) error {
			return AnyPlayerPays{
				Pay:          pay,
				Question:     name + " — " + what + " to counter it?",
				GuardsSource: true,
				Then:         then,
			}.Apply(NewContext(g, item))
		})
}

// anyPlayerCounterThisSpell is "If a player does, counter <this spell>",
// keyed (#4264): the spell is the offer's source.
var anyPlayerCounterThisSpell = AnyPlayerConsequences("any-player/counter-this-spell",
	func(ctx *Context, _ game.AnyPlayerAnswer) error { return counterThisSpell(ctx) }, nil, nil)

// anyPlayerCounterThisSpellHalfLife is the same with Temporal Extortion's
// payment, half the asked player's life rounded up, worked out per
// player.
var anyPlayerCounterThisSpellHalfLife = AnyPlayerConsequences("any-player/counter-this-spell-half-life",
	func(ctx *Context, _ game.AnyPlayerAnswer) error { return counterThisSpell(ctx) }, nil,
	func(g *game.Game, payer uuid.UUID, _ []uuid.UUID) (game.PayAction, string) {
		return HalfTheirLifeRoundedUp(g, payer)
	})

// anyPlayerMayTakeDamageFromThis is "Any player may have <this> deal N
// damage to them. If no one does, <onNobody>." (Book Burning, Breaking
// Point, Browbeat) as a spell's OnResolve (ADR 0150 §3). The spell
// deals the damage through the ordinary path, so prevention and
// replacement apply, and taking it is always a payment. `nobody` is "if
// no one does" as a registered consequence (AnyPlayerConsequences, #4264),
// so a table waiting on any player's answer is a restore point.
func anyPlayerMayTakeDamageFromThis(name string, n int, what string, nobody game.AnyPlayerThen) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return AnyPlayerPays{
			Pay:          DamagePayment(n),
			Question:     name + " — take " + strconv.Itoa(n) + " damage to stop " + what + "?",
			AcceptLabel:  "Take " + strconv.Itoa(n) + " damage",
			DeclineLabel: "Don't",
			Then:         nobody,
		}.Apply(ctx)
	}
}

// anyPlayerMaySacrificeALandToTuckThis is "When this creature enters,
// any player may sacrifice a land of their choice. If a player does,
// put this creature on top of its owner's library." (Argothian Wurm,
// Shivan Wumpus.) A Wurm that has already left, or come back as a new
// object, is not moved (CR 400.7).
func anyPlayerMaySacrificeALandToTuckThis(name string) game.TriggeredAbility {
	return WhenThisEnters(name+" — any player may sacrifice a land to put it on top of its owner's library",
		func(g *game.Game, item *game.StackItem) error {
			return AnyPlayerPays{
				Pay:      SacrificePayment(1, "land", "lands", QueryType("land")),
				Question: name + " — sacrifice a land to put it on top of its owner's library?",
				Then:     anyPlayerTucksThis,
			}.Apply(NewContext(g, item))
		})
}

// AnyPlayerPaysForEach is "For each <object>, <effect> unless any
// player pays <N life>" (ADR 0187 §2, #3572: Cleansing). It is ONE
// offer for the whole set, not a prompt per object: every player in
// APNAP order from the active player chooses which objects they will pay
// for, knowing what the earlier players chose, then every payment is
// made, one life payment per object chosen. An object anyone paid for is
// saved; Then gets the rest (game.QueueAnyPlayerPaysForEachForEffect).
//
// Unlike AnyPlayerPays the first payer does not end it: two players may
// pay for the same object, and each pays. Bound to the resolving stack
// item the way AnyPlayerPays is, so Then reads the item's targets and X
// when it finally runs.
type AnyPlayerPaysForEach struct {
	// Objects are the battlefield objects, worked out as the effect
	// is applied (CR 608.2h).
	Objects func(ctx *Context) []uuid.UUID
	// EachLife is the life one object costs.
	EachLife int
	// Noun names one object ("land").
	Noun string
	// Question is the prompt header each player sees. Optional.
	Question string
	// Then is the effect on the objects nobody paid for, a REGISTERED
	// consequence (AnyPlayerEachConsequence, #4264), handed only objects
	// that are still the same objects; Carry is the plain data it is
	// handed. Every leg of the offer holds no closure, so a table waiting
	// on any seat's pick is a restore point. Required.
	Then  game.AnyPlayerEachThen
	Carry []uuid.UUID
}

// AnyPlayerEachConsequence registers a keyed per-object offer's
// consequence under `key`, an on-disk identity (never renamed, never
// reused, ledgered "choose <key>", the offer's choose_cards continuation):
// `onUnsaved` is "the objects nobody paid for ...", handed them in r.Unsaved
// and run against a Context bound to the asking item the offer carried.
// Call it once, from a package-level var.
func AnyPlayerEachConsequence(key string, onUnsaved func(ctx *Context, r game.AnyPlayerEachAnswer) error) game.AnyPlayerEachThen {
	return game.RegisterAnyPlayerPaysEach(key, func(g *game.Game, r game.AnyPlayerEachAnswer) error {
		if r.Item != nil {
			return onUnsaved(NewContext(g, r.Item), r)
		}
		return onUnsaved(keyedContext(g, r.Controller, r.Source), r)
	})
}

func (a AnyPlayerPaysForEach) Apply(ctx *Context) error {
	if a.Then.Key() == "" {
		return fmt.Errorf("%w: AnyPlayerPaysForEach with no registered consequence (AnyPlayerEachConsequence)", game.ErrInvalidParam)
	}
	item := ctx.Item
	var objects []uuid.UUID
	if a.Objects != nil {
		objects = a.Objects(ctx)
	}
	return ctx.Game.QueueAnyPlayerPaysForEachForEffect(game.AnyPlayerPaysForEach{
		Source:     ctx.Source(),
		Controller: ctx.Controller(),
		Objects:    objects,
		EachLife:   a.EachLife,
		Noun:       a.Noun,
		Question:   a.Question,
		Then:       a.Then,
		Carry:      a.Carry,
		Asked:      game.NewAskedItem(item),
	})
}

// anyPlayerTucksThis is "If a player does, put this creature on top of its
// owner's library", keyed (#4264): a creature that left and came back is a
// new object and is not moved (CR 400.7), which the asking item's object
// answers.
var anyPlayerTucksThis = AnyPlayerConsequences("any-player/tuck-this-on-top-of-library",
	func(ctx *Context, _ game.AnyPlayerAnswer) error {
		if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return ctx.Game.TuckToLibraryForEffect(ctx.Item.SourceCardID, false)
	}, nil, nil)
