package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// AnyPlayerPays is "<effect> unless any player pays <Cost>" and "Any
// player may pay <Cost>. If no one does, <effect>" (ADR 0150, #1903):
// the payment is offered to every player still in the game, in APNAP
// order from the active player, the controller included, until one
// pays (game.QueueAnyPlayerPaysForEffect). The offer blocks the step
// it was made in and takes no standing answer.
//
// Both branches are bound to the resolving stack item the way
// PayUnless's are, so OnNobody reads the item's targets and X when it
// finally runs (Wild Might's creature, Soul Strings' cards).
type AnyPlayerPays struct {
	// Cost is the payment ("{2}"), with any X already worked out.
	Cost string
	// Question is the prompt header each player sees.
	Question string
	// OnNobody is the effect that happens when nobody pays.
	OnNobody func(ctx *Context) error
	// OnPaid is "If a player does, …", handed the payer. Optional.
	OnPaid func(ctx *Context, payer uuid.UUID) error
	// HelpsController marks a payment the controller wants made (Icy
	// Prison's upkeep): game.AnyPlayerPays.HelpsController.
	HelpsController bool

	// The fields below are ADR 0150 PR 2's (§3): payments other than
	// mana, and the "any opponent" and "counter it" shapes.

	// Pay is a payment that is not mana: LifePayment, DamagePayment,
	// SacrificePayment, DiscardPayment or NoPayment. Cost is unused
	// when it is set; the offer's words are the payment's, lower-cased
	// ("sacrifice a land", "5 life").
	Pay UpkeepPayment
	// PerPayer works the payment out for each player as they are
	// asked (Temporal Extortion's half their life, rounded up). It
	// replaces Pay and Cost for that player.
	PerPayer func(g *game.Game, payer uuid.UUID) (game.PayAction, string)
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
	item := ctx.Item
	var onPaid func(*game.Game, uuid.UUID) error
	if a.OnPaid != nil {
		paid := a.OnPaid
		onPaid = func(g *game.Game, payer uuid.UUID) error {
			return paid(NewContext(g, item), payer)
		}
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
	return ctx.Game.QueueAnyPlayerPaysForEffect(game.AnyPlayerPays{
		Source:       ctx.Source(),
		Controller:   ctx.Controller(),
		Cost:         cost,
		Action:       action,
		PerPayer:     a.PerPayer,
		Question:     a.Question,
		Opponents:    a.Opponents,
		OnPaid:       onPaid,
		OnNobody:     declineAgainstTheSameItem(item, a.OnNobody),
		Guards:       guards,
		AcceptLabel:  a.AcceptLabel,
		DeclineLabel: a.DeclineLabel,

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
// deciding. `pay` is the payment, or the zero value with `perPayer`
// set for one worked out per player.
func anyPlayerMayCounterThisSpell(name, what string, pay UpkeepPayment, perPayer func(*game.Game, uuid.UUID) (game.PayAction, string)) game.TriggeredAbility {
	return WhenYouCastThisSpell(name+" — any player may "+what+" to counter it",
		func(g *game.Game, item *game.StackItem) error {
			return AnyPlayerPays{
				Pay:          pay,
				PerPayer:     perPayer,
				Question:     name + " — " + what + " to counter it?",
				GuardsSource: true,
				OnPaid: func(ctx *Context, _ uuid.UUID) error {
					return counterThisSpell(ctx)
				},
			}.Apply(NewContext(g, item))
		})
}

// anyPlayerMayTakeDamageFromThis is "Any player may have <this> deal N
// damage to them. If no one does, <onNobody>." (Book Burning, Breaking
// Point, Browbeat) as a spell's OnResolve (ADR 0150 §3). The spell
// deals the damage through the ordinary path, so prevention and
// replacement apply, and taking it is always a payment.
func anyPlayerMayTakeDamageFromThis(name string, n int, what string, onNobody func(ctx *Context) error) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return AnyPlayerPays{
			Pay:          DamagePayment(n),
			Question:     name + " — take " + strconv.Itoa(n) + " damage to stop " + what + "?",
			AcceptLabel:  "Take " + strconv.Itoa(n) + " damage",
			DeclineLabel: "Don't",
			OnNobody:     onNobody,
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
				OnPaid: func(ctx *Context, _ uuid.UUID) error {
					if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return ctx.Game.TuckToLibraryForEffect(ctx.Item.SourceCardID, false)
				},
			}.Apply(NewContext(g, item))
		})
}
