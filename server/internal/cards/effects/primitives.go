package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// primitives.go declares the composable effect primitives. Each
// is a plain struct carrying the parameters the effect needs, with
// an Apply(ctx *Context) error method that delegates to the game's
// locked-context API (effect_api.go).
//
// Composition is explicit in per-card Spec files: a card that does
// "deal 3 damage then gain 3 life" sequences two primitives in its
// OnResolve body. The primitives themselves are single-purpose.
//
// Error handling: any Apply that can't complete (target no longer
// exists, player left, etc.) returns an error that the caller
// logs via EventEffectError. The resolution path never crashes on
// a primitive error — partial application is the S14 sandbox posture.

// DealDamage deals `Amount` damage from `Source` to `Target`.
// Target can be either a seated player or a battlefield card; the
// primitive routes to the right path via PlayerByID vs. zone scan.
// Non-positive Amount is a no-op. A Target that no longer resolves
// (player eliminated, card moved off the battlefield) silently
// emits EventEffectError and returns nil so the rest of a composed
// effect can still run.
type DealDamage struct {
	Source uuid.UUID
	Target uuid.UUID
	Amount int
}

func (d DealDamage) Apply(ctx *Context) error {
	if d.Amount <= 0 {
		return nil
	}
	if p := ctx.Game.PlayerByIDForEffect(d.Target); p != nil {
		return ctx.Game.DealDamageToPlayerForEffect(d.Source, d.Target, d.Amount)
	}
	if z := ctx.Game.FindCardZoneForEffect(d.Target); z != nil && z.Kind == game.ZoneBattlefield {
		return ctx.Game.DealDamageToCreatureForEffect(d.Source, d.Target, d.Amount)
	}
	return nil
}

// GainLife gives `Player` `Amount` life. Shorthand for
// ChangePlayerLife(+Amount) — present as a named primitive so
// cards read declaratively ("deal 3, gain 3 life" for Helix).
type GainLife struct {
	Player uuid.UUID
	Amount int
}

func (g GainLife) Apply(ctx *Context) error {
	if g.Amount == 0 {
		return nil
	}
	player := g.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, g.Amount)
}

// DrawCards draws N cards for `Player`. Non-positive N is a no-op.
// Drawing from an empty library sets the player's AttemptedEmptyDraw
// via the standard drawCardLocked path (CR 704.5b).
type DrawCards struct {
	Player uuid.UUID
	N      int
}

func (d DrawCards) Apply(ctx *Context) error {
	if d.N <= 0 {
		return nil
	}
	player := d.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.DrawNForEffect(player, d.N)
}

// DiscardCards removes N cards from `Player`'s hand AT RANDOM
// (CR 701.8b). Cards land in the graveyard known to every seated
// player (public-zone rule).
//
// It is for cards that print "at random" — Burning Inquiry — and for
// nothing else (#651). A discard the player CHOOSES is
// game.QueueDiscardChoiceForEffect, which opens a real prompt over
// their hand and holds the table until it is answered; using this
// primitive for one is a different card, not a simplification of it.
type DiscardCards struct {
	Player uuid.UUID
	N      int
}

func (d DiscardCards) Apply(ctx *Context) error {
	if d.N <= 0 {
		return nil
	}
	player := d.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.DiscardRandomForEffect(player, d.N)
}

// MillCards mills the top N cards of `Player`'s library to their
// graveyard. A library holding fewer mills what it has and nobody
// loses for it (CR 701.17b) — only a draw from an empty library does.
type MillCards struct {
	Player uuid.UUID
	N      int
}

func (m MillCards) Apply(ctx *Context) error {
	if m.N <= 0 {
		return nil
	}
	player := m.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.MillNForEffect(player, m.N)
}

// DestroyTarget routes a battlefield permanent to its owner's
// graveyard. Wrath-style iteration composes this inside a
// `for _, id := range ctx.CreatureIDs()` loop in the card's
// OnResolve.
type DestroyTarget struct {
	Target uuid.UUID

	// CantBeRegenerated is the clause printed on Mortify, Putrefy,
	// Pongify, Terminate, Snuff Out and the rest: this destruction
	// ignores regeneration shields (CR 701.19c). The shields are not
	// spent — CR 701.19d leaves an ignored one on the permanent.
	//
	// It was cosmetic on every card that printed it until #667 gave
	// the engine a shield to ignore. Set it wherever the oracle text
	// says it; leaving it off a card that prints it is a real bug now.
	CantBeRegenerated bool
}

func (d DestroyTarget) Apply(ctx *Context) error {
	return ctx.Game.DestroyPermanentForEffect(d.Target,
		game.DestroyOptions{CantBeRegenerated: d.CantBeRegenerated})
}

// Regenerate creates one regeneration shield for a permanent
// (CR 701.19a) — "Regenerate target creature", "regenerate it".
//
// The shield replaces the NEXT destruction of that permanent this
// turn: instead of being destroyed it is tapped, all damage is
// removed from it, and it is removed from combat. It is used up doing
// so, it expires at the cleanup step if it is not, and a second
// Regenerate stacks a second shield. Everything about how that works
// is in game/regeneration.go; a card just says this.
//
// What it does NOT save the permanent from: a sacrifice (CR 701.21a),
// zero toughness (CR 704.5f), the legend rule, exile, a bounce, or a
// destruction whose effect says it can't be regenerated
// (CR 701.19c — DestroyTarget.CantBeRegenerated above).
//
// A target that is no longer on the battlefield is a no-op rather
// than an error: CR 701.19b regenerates nothing.
type Regenerate struct {
	Target uuid.UUID
}

func (r Regenerate) Apply(ctx *Context) error {
	if r.Target == uuid.Nil {
		return nil
	}
	if err := ctx.Game.RegenerateForEffect(r.Target); err != nil {
		// The permanent left before the ability resolved. CR 701.19b:
		// regenerating a permanent that is not on the battlefield does
		// nothing, which is not an error the card should report.
		return nil
	}
	return nil
}

// SacrificePermanent sacrifices a battlefield permanent on behalf
// of its controller (CR 701.21). Distinct from DestroyTarget:
// sacrifice ignores indestructible / regeneration and fires
// EventSacrifice as well as the ordinary dies-trigger, which is
// what aristocrats payoffs watch. Added in S21 sub-PR 1.
type SacrificePermanent struct {
	Target uuid.UUID

	// Then is the "if you do" / "for each permanent sacrificed this
	// way" clause for ONE permanent, and `sacrificed` is whether it
	// really left the battlefield. Optional; leave it nil for a plain
	// sacrifice with nothing hanging off it.
	//
	// #993, and the same shape as ExileTarget.Then (#870). A sacrifice
	// is not itself replaceable (CR 701.17a), but the MOVE it makes is
	// an ordinary zone change, so a sacrificed commander opens the
	// CR 903.9 window and the permanent is still on the battlefield
	// while the question is open. A clause written on the next line
	// therefore reads "still here, so it was not sacrificed" for a leg
	// that is merely PAUSED and pays out nothing for a sacrifice that
	// does land a beat later.
	//
	// `sacrificed` is sacrificedThisWayLocked's answer: true whenever
	// the permanent left the battlefield, including a commander that
	// took the command zone and a card an "exile it instead"
	// replacement took — only where it went was replaced. Write the
	// clause as something that acts on what it is told, not as the next
	// line of the card.
	Then func(ctx *Context, sacrificed bool) error
}

func (s SacrificePermanent) Apply(ctx *Context) error {
	if s.Then == nil {
		return ctx.Game.SacrificePermanentForEffect(s.Target)
	}
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract massEffect.apply explains: an undo restores
	// this game's fields in place, so a captured *Game would be the
	// wrong one.
	//
	// The source is uuid.Nil for the reason the fire-and-forget form
	// passes none: SacrificePermanentForEffect stamps no source on
	// EventSacrifice, and adding a Then must change the sequencing and
	// nothing else.
	item := ctx.Item
	return ctx.Game.SacrificeThenForEffect(uuid.Nil, s.Target, func(g *game.Game, sacrificed bool) error {
		return s.Then(NewContext(g, item), sacrificed)
	})
}

// ExileTarget moves a card from whichever zone it's in to the
// shared exile zone. Works for battlefield permanents and hand
// / graveyard cards alike.
type ExileTarget struct {
	Target uuid.UUID

	// Then is the "if you do" / "for each card exiled this way"
	// clause for ONE card, and `exiled` is whether the card actually
	// reached exile. Optional; leave it nil for a plain exile with
	// nothing hanging off it.
	//
	// #870: it runs from a CONTINUATION for the reason
	// ExileAllMatching.Then does — the exile opens the CR 614 window,
	// so a commander stops to answer CR 903.9 and whether it was
	// exiled is not knowable on the next line. `exiled` is false when
	// the window cancelled the move, when a replacement sent the card
	// somewhere else, and when a commander took the command zone: it
	// left, but not to exile (CR 400.7). Write the clause as
	// something that acts on what it is told, not as the next line of
	// the card.
	Then func(ctx *Context, exiled bool) error
}

func (e ExileTarget) Apply(ctx *Context) error {
	if e.Then == nil {
		return ctx.Game.ExileCardForEffect(e.Target)
	}
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract massEffect.apply explains: an undo restores
	// this game's fields in place, so a captured *Game would be the
	// wrong one.
	item := ctx.Item
	return ctx.Game.ExileCardThenForEffect(e.Target, func(g *game.Game, exiled bool) error {
		return e.Then(NewContext(g, item), exiled)
	})
}

// ExileThenIfItWas is "Exile target card from a graveyard. If it was
// a creature card, <clause>" — Cling to Dust, Scavenging Ooze and
// Deluge of the Dead, and the ONE body all three now share (#911).
//
// # Two facts, two moments
//
// The clause reads WHAT THE CARD WAS, which is a question about the
// past: after the move the card is in exile with none of its
// battlefield-era layers, so `Was` is answered BEFORE anything moves
// (CR 608.2h / last-known information). That much the three cards
// already did by hand.
//
// What they did NOT do is wait for the move. `ExileTarget` without a
// `Then` is fire-and-forget: it returns nil when the CR 614 window
// CANCELLED the exile ("cards in graveyards can't be exiled"), when a
// replacement sent the card somewhere else, and when the move merely
// PAUSED on a commander card's CR 903.9 prompt. All three paid out
// anyway — life, a +1/+1 counter, a Zombie — for a card that was still
// sitting in its graveyard. So the clause hangs off `ExileTarget.Then`
// and is gated on `exiled`, CR 400.7's reading: the card that ARRIVED
// in exile is the one the effect exiled.
//
// # Why no exile means no clause at all
//
// ADR 0013 §5m left this as a rules question and §5t answers it. "If
// it WAS a creature card" has no referent when nothing was exiled:
// "it" is the card the first sentence moved, and CR 614.10 says an
// event replaced with nothing never happened. So neither branch runs —
// Cling to Dust's `Otherwise` ("you draw a card") is the other half of
// the same conditional, not a separate sentence, and a Cling to Dust
// whose exile was cancelled draws nothing.
//
// That is the line between this primitive and the exile-then-an-
// unconditional-clause family (Swords to Plowshares, Solitude, Path to
// Exile's search), which §5m declared ungated and which stays ungated:
// there the second sentence is about a player, makes no claim about the
// card, and happens either way.
//
// A commander card that takes CR 903.9's offer left the graveyard but
// did not reach exile, so it pays out nothing either — the same answer
// the batch gives "for each card exiled this way".
type ExileThenIfItWas struct {
	Target uuid.UUID

	// Was is the question the clause asks about the card, answered
	// against the card as it was BEFORE the exile. Nil means "any
	// card", which turns this into a plain "exile it; if you do, …".
	// WasCreatureCard is the printed phrase all three cards use.
	Was func(c game.Card) bool

	// Then is the clause. It runs only when the card actually reached
	// exile AND Was said yes.
	Then func(ctx *Context) error

	// Otherwise is Cling to Dust's "Otherwise, you draw a card": the
	// same conditional's other branch, so it runs only when the card
	// reached exile and Was said no. Optional.
	Otherwise func(ctx *Context) error
}

// WasCreatureCard is the predicate behind the printed phrase "if it
// was a creature card". Named so the three cards read like their own
// oracle text and so a fourth does not re-derive it.
func WasCreatureCard(c game.Card) bool { return c.IsCreature() }

func (e ExileThenIfItWas) Apply(ctx *Context) error {
	was := false
	if c, ok := ctx.Game.LookupCardForEffect(e.Target); ok {
		was = e.Was == nil || e.Was(c)
	}
	return ExileTarget{
		Target: e.Target,
		Then: func(ctx *Context, exiled bool) error {
			if !exiled {
				// Nothing was exiled, so there is no "it" for the
				// clause to be about. Neither branch.
				return nil
			}
			clause := e.Then
			if !was {
				clause = e.Otherwise
			}
			if clause == nil {
				return nil
			}
			return clause(ctx)
		},
	}.Apply(ctx)
}

// ReturnFromExile puts a card that is currently in exile back onto
// the battlefield — the other half of a flicker, and the primitive
// the catalog was missing entirely before S22 (ExileTarget could
// send a permanent to exile; nothing brought one back).
//
// Controller is who it returns under; leave it zero for "under its
// owner's control", which is what nearly every card says. Set Tapped
// for "return that card to the battlefield tapped".
//
// The returned permanent is a NEW OBJECT with a fresh InstanceID
// (CR 400.7) and re-triggers ETB — both the `Triggered` /
// `EventETB` path and the direct `OnETB` hook. A target that is no
// longer in exile is silently skipped: a delayed return whose card
// moved on in the meantime does nothing rather than erroring.
type ReturnFromExile struct {
	Target     uuid.UUID
	Controller uuid.UUID
	Tapped     bool
}

func (r ReturnFromExile) Apply(ctx *Context) error {
	z := ctx.Game.FindCardZoneForEffect(r.Target)
	if z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	_, err := ctx.Game.ReturnFromExileToBattlefieldForEffect(r.Target, r.Controller, r.Tapped)
	return err
}

// Flicker is "exile it, then return it to the battlefield" resolved
// in one go — Y'shtola Rhul, Thassa, Restoration Angel, Ephemerate.
// The permanent is gone for no observable window at all, but it
// still comes back as a new object, so counters and damage fall off
// and every ETB fires again.
//
// The delayed variant ("exile it. At the beginning of the next end
// step, return it") is NOT this: it exiles now and schedules a
// separate delayed trigger for the return — see
// ScheduleDelayedTrigger.
//
// #894: the return is the exile's CONTINUATION, because the exile can
// pause. A commander flickered by the old two-line form was asked about
// the command zone, the return ran with that question still open and
// found nothing in exile to bring back, and the commander landed in
// exile for good a moment later. The return now waits, and it happens
// only if the card really reached exile — a commander that takes the
// command zone stays there, which is the printed outcome rather than a
// stranded card.
type Flicker struct {
	Target     uuid.UUID
	Controller uuid.UUID
	Tapped     bool
}

func (f Flicker) Apply(ctx *Context) error {
	return ExileTarget{
		Target: f.Target,
		Then: func(ctx *Context, exiled bool) error {
			if !exiled {
				return nil
			}
			return ReturnFromExile{Target: f.Target, Controller: f.Controller, Tapped: f.Tapped}.Apply(ctx)
		},
	}.Apply(ctx)
}

// ScheduleDelayedTrigger registers a CR 603.7 delayed triggered
// ability: "at the beginning of the next end step, <do X>". The
// instruction sits on the Game rather than on any card — the spell
// that created it is usually in a graveyard by the time it fires —
// and goes on the stack when the named step begins, so every player
// gets a response window before it resolves.
//
// "Next" needs no bookkeeping: the queue is drained on step ENTRY,
// so an ability scheduled during an end step waits for the following
// one. That is observable — a blink cast in an opponent's end step
// returns the permanent a whole turn later.
//
// Cards is the payload, stamped onto the fired stack item's Targets;
// Effect reads it back from item.Targets rather than closing over it,
// which is what keeps the trigger correct across a Clone / undo.
type ScheduleDelayedTrigger struct {
	// At is the step whose beginning fires the trigger. Zero means
	// game.StepEnd — "the next end step" is the overwhelmingly
	// common case.
	At game.Step

	// Label is the stack-overlay copy, phrased like a trigger's:
	// "Waterbender's Restoration — return the exiled creatures".
	Label string

	// Controller is who controls the delayed ability. Zero means the
	// controller of the effect scheduling it (CR 603.7d).
	Controller uuid.UUID

	// ControllerTurnOnly is "at the beginning of YOUR next <step>"
	// (Mana Drain's main phase) as opposed to "the next turn's
	// <step>" (Arcane Denial's upkeep, which the very next player's
	// upkeep satisfies). Leave it false unless the printed text says
	// "your".
	ControllerTurnOnly bool

	// Cards is the instance IDs the effect acts on.
	Cards []uuid.UUID

	// Effect runs when the trigger's stack item resolves. Same
	// contract as a triggered ability's: read everything off `item`
	// and the `g` handed in, capture neither a *Game nor a pointer
	// into a zone slice.
	Effect func(g *game.Game, item *game.StackItem) error
}

func (s ScheduleDelayedTrigger) Apply(ctx *Context) error {
	controller := s.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	at := s.At
	if at == "" {
		at = game.StepEnd
	}
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:         controller,
		SourceCardID:       ctx.Source(),
		Label:              s.Label,
		At:                 at,
		ControllerTurnOnly: s.ControllerTurnOnly,
		Cards:              s.Cards,
		Effect:             s.Effect,
	})
	return nil
}

// BounceToHand returns a card to its owner's hand. Used by
// Unsummon-style effects.
type BounceToHand struct {
	Target uuid.UUID

	// Then is the "then …" / "if you do" clause for ONE card, and
	// `bounced` is whether the card actually reached a hand. Optional;
	// leave it nil for a plain bounce with nothing hanging off it.
	//
	// #993, and the same shape as ExileTarget.Then (#870). A hand is a
	// CR 903.9 destination, so every bounce can pause: a commander
	// returned to its owner's hand stops to ask them about the command
	// zone, and a clause written on the next line runs with the
	// permanent still on the battlefield and the question still open.
	// Chain of Vapor asked its controller to sacrifice a land while
	// they were already being asked something else.
	//
	// `bounced` is CR 400.7's reading: false when the window cancelled
	// the move, when a replacement sent the card somewhere else, and
	// when a commander took the command zone — it left, but not to a
	// hand. A clause that is NOT gated on the move (the common case for
	// a bounce: "then that permanent's controller may …" is a sentence
	// about a player) simply ignores the argument and gets the ordering
	// for free.
	Then func(ctx *Context, bounced bool) error
}

func (b BounceToHand) Apply(ctx *Context) error {
	if b.Then == nil {
		return ctx.Game.BounceToHandForEffect(b.Target)
	}
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract massEffect.apply explains: an undo restores
	// this game's fields in place, so a captured *Game would be the
	// wrong one.
	item := ctx.Item
	return ctx.Game.BounceToHandThenForEffect(b.Target, func(g *game.Game, bounced bool) error {
		return b.Then(NewContext(g, item), bounced)
	})
}

// TapTarget taps a battlefield card. No effect if the card is not
// on the battlefield.
type TapTarget struct {
	Target uuid.UUID
}

func (t TapTarget) Apply(ctx *Context) error {
	return ctx.Game.TapTargetForEffect(t.Target)
}

// UntapTarget untaps a battlefield card.
type UntapTarget struct {
	Target uuid.UUID
}

func (u UntapTarget) Apply(ctx *Context) error {
	return ctx.Game.UntapTargetForEffect(u.Target)
}

// CounterTarget counters a spell or ability on the stack. Routes
// spell items to their owner's graveyard by default; ability items
// cease to exist. `StackID` is the StackItem.ID (equal to the
// card's InstanceID for spells, a synthetic UUID for abilities).
type CounterTarget struct {
	StackID uuid.UUID
}

func (c CounterTarget) Apply(ctx *Context) error {
	return ctx.Game.CounterTargetForEffect(c.StackID)
}

// CounterAllMatching counters every spell on the stack matching
// Match — the "change 'target' in its text to 'each'" an overloaded
// counterspell applies to its own printed "counter target spell"
// (CR 702.96, Counterflux).
//
// The set is snapshotted once, before anything is countered — CR
// 608.2's "objects as they existed when the spell began resolving" —
// so a spell that leaves the stack as a side effect of an earlier
// counter in the same sweep is neither double-counted nor able to
// dodge by leaving mid-resolution, and nothing that arrives after the
// sweep began is swept up with it.
type CounterAllMatching struct {
	Match CardPredicate
}

func (c CounterAllMatching) Apply(ctx *Context) error {
	if c.Match == nil || ctx.Game.Stack == nil {
		return nil
	}
	caster := ctx.Controller()
	var ids []uuid.UUID
	for _, card := range ctx.Game.Stack.Cards {
		if c.Match(ctx.Game, caster, card) {
			ids = append(ids, card.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (CounterTarget{StackID: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// AddCounter places (or removes, via negative N) N `Kind` counters
// on `Target`. Target can be in any zone — the primitive does not
// gate on card type. Used by The Wandering Emperor's OnETB hook
// to stamp starting loyalty.
type AddCounter struct {
	Target uuid.UUID
	Kind   string
	N      int
}

func (a AddCounter) Apply(ctx *Context) error {
	return ctx.Game.AddCounterForEffect(a.Target, a.Kind, a.N)
}

// CreateToken puts N copies of `Template` on the battlefield under
// `Controller`'s control. Each token gets a fresh InstanceID and
// is publicly known to every seated player. The template's Owner /
// Controller / InstanceID / Counters / KnownBy are overwritten —
// callers only populate the "printed" fields (Name, TypeLine,
// Power, Toughness).
type CreateToken struct {
	Controller uuid.UUID
	Template   game.Card
	N          int
}

func (c CreateToken) Apply(ctx *Context) error {
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	return ctx.Game.CreateTokenForEffect(controller, c.Template, c.N)
}

// ReturnFromGraveyard moves a card from a graveyard to `Dest`
// (typically ZoneHand, occasionally ZoneBattlefield for
// reanimation). Target must currently be in a graveyard.
type ReturnFromGraveyard struct {
	Target uuid.UUID
	Dest   game.ZoneKind

	// Controller is who the card lands under the control of, and is
	// meaningful only for Dest == ZoneBattlefield. Zero value means
	// "its owner", which is right for "return target creature card
	// FROM YOUR GRAVEYARD" (Zombify) and wrong for "put target
	// creature card from A GRAVEYARD onto the battlefield UNDER YOUR
	// CONTROL" (Reanimate) — set it to ctx.Controller() for the
	// second. Reanimating an opponent's creature and handing it back
	// to the opponent is the failure mode this field exists to stop.
	Controller uuid.UUID
}

func (r ReturnFromGraveyard) Apply(ctx *Context) error {
	return ctx.Game.ReturnFromGraveyardUnderControlForEffect(r.Target, r.Dest, r.Controller)
}

// SearchLibrary looks through `Player`'s library for up to `Limit`
// cards matching `Predicate`, moves them to `Dest`, optionally
// reveals them to all seated players (via KnownBy), and optionally
// shuffles the library afterwards.
//
// S22: the SEARCHER chooses. When the library holds more matches
// than Limit — or the clause is a "you may search", or a Validate
// constraint is in play — the engine queues a search prompt and this
// primitive returns immediately; the cards move when the searcher
// answers. When there is nothing to decide it stays synchronous.
//
// That asynchrony is why `Then` exists: anything the card does AFTER
// the search ("then shuffle" aside) has to run in the continuation,
// not on the line below the Apply call. Fabled Passage's "untap that
// land" and Gamble's random discard are the two shapes.
type SearchLibrary struct {
	Player    uuid.UUID
	Predicate func(game.Card) bool
	Dest      game.ZoneKind
	Limit     int
	Reveal    bool
	Shuffle   bool
	// TappedOnEntry stamps Card.Tapped = true on fetched permanents
	// headed for the battlefield. Matches the literal card text on
	// Cultivate / Path to Exile / Solemn Simulacrum — land enters
	// tapped even when not otherwise specified by the destination.
	// Only meaningful when Dest == ZoneBattlefield.
	//
	// It is the FETCHING effect's tapped clause, not the fetched
	// card's. Since S22 the fetched card's own enters-tapped
	// replacement runs too (#263), and the two are OR-ed.
	TappedOnEntry bool
	// Optional is "you MAY search" (CR 701.23b) — Assassin's Trophy,
	// Path to Exile, Solemn Simulacrum. Forces the prompt so the
	// searcher can decline the card AND the shuffle.
	Optional bool
	// Reason is the prompt banner. Name the card.
	Reason string
	// Validate is a legality check on the picked SET, for clauses no
	// per-card predicate can express — Myriad Landscape's "two basic
	// land cards that share a land type".
	Validate func([]game.Card) bool
	// Then is the rest of the effect, receiving the cards actually
	// found. Runs inline when no prompt was needed and from the
	// resolve path when one was.
	Then func(g *game.Game, found []uuid.UUID) error
	// Source is the card that caused the search — prompt context.
	Source uuid.UUID
	// ToTop is the "... then shuffle and put that card ON TOP"
	// clause the one-mana tutors print (Enlightened, Worldly,
	// Mystical, Vampiric, Imperial Seal). Set it with
	// Dest: game.ZoneLibrary. The card never leaves the library; it
	// is placed after the shuffle, which is what makes the clause
	// mean anything.
	ToTop bool
}

func (s SearchLibrary) Apply(ctx *Context) error {
	source := s.Source
	if source == uuid.Nil && ctx != nil {
		source = ctx.Source()
	}
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:        s.Player,
		Source:        source,
		Pred:          s.Predicate,
		Dest:          s.Dest,
		Limit:         s.Limit,
		Reveal:        s.Reveal,
		Shuffle:       s.Shuffle,
		TappedOnEntry: s.TappedOnEntry,
		Optional:      s.Optional,
		Reason:        s.Reason,
		Validate:      s.Validate,
		Then:          s.Then,
		ToTop:         s.ToTop,
	})
}

// PayUnless queues the CR 118.12 "unless that player pays <Cost>"
// prompt for Chooser. The calling effect has already resolved;
// what remains is the payer's decision, which arrives later via
// resolve_choice. On "no" — or "yes" without the mana in pool +
// untapped sources — OnDecline runs against a fresh Context bound
// to the same stack item, so it can read Controller / Source the
// way the outer effect did. Used by Rhystic Study, Smothering
// Tithe, Esper Sentinel. Added in S19 sub-PR 6.
type PayUnless struct {
	Chooser   uuid.UUID
	Cost      string
	Question  string
	OnDecline func(ctx *Context) error
}

// USE UpkeepPayUnless FOR "AT THE BEGINNING OF YOUR UPKEEP, PAY OR
// ELSE". It is the same CR 118.12 prompt, and the difference is the
// whole of #997: a pay-unless the ACTIVE player owes during their OWN
// upkeep must stop the table while it is unanswered, because the rest
// of the turn is what hangs on the answer. A card that writes that
// clause out of PayUnless gets the Rhystic Study latitude instead and
// the table can take the turn with the question still open. Stasis
// and Pact of Negation did exactly that.

func (p PayUnless) Apply(ctx *Context) error {
	return ctx.Game.QueuePayUnlessForEffect(p.Chooser, ctx.Source(), p.Cost, p.Question,
		declineAgainstTheSameItem(ctx.Item, p.OnDecline))
}

// declineAgainstTheSameItem adapts a card-side "or else" (a *Context
// branch) to the engine-side one (a *game.Game branch) for the two
// pay-unless primitives that take one.
//
// The Context is built FRESH when the branch runs, against whichever
// *Game it is handed and bound to the same stack item: an undo
// restores a clone, so the game the answer arrives at is not the one
// that asked (StackItem.Effect's contract, which every resume frame
// in the engine keeps). `item` is a detached pointer and the closure
// captures nothing else.
func declineAgainstTheSameItem(item *game.StackItem, decline func(ctx *Context) error) func(*game.Game) error {
	return func(g *game.Game) error {
		if decline == nil {
			return nil
		}
		return decline(NewContext(g, item))
	}
}

// UpkeepPayUnless is "at the beginning of your upkeep, pay <Cost> or
// <consequence>" — Stasis's "sacrifice Stasis unless you pay {U}",
// Pact of Negation's "pay {3}{U}{U}. If you don't, you lose the
// game", and every cumulative upkeep (CR 702.24).
//
// It is PayUnless with the halt the shape needs and cannot be trusted
// to ask for. The prompt is addressed to the player whose upkeep it
// is, about their own permanent or their own survival, so the table
// must not leave the step while it is unanswered (CR 117.3, CR
// 500.4). The engine derives that from the cursor rather than from a
// flag on the card (Game.QueueUpkeepPayUnlessForEffect,
// game/upkeep_pay_unless.go): #567 shipped the flag, and the two
// cards written afterwards with the same sentence printed on them
// both missed it.
//
// The step is read off the cursor when the prompt is queued, so the
// same primitive is right for a beginning-of-end-step pay-or-else; it
// is named for the family every printed card of it belongs to.
type UpkeepPayUnless struct {
	// Chooser is the player asked to pay — the permanent's
	// controller, which is who the printed clause always means.
	Chooser uuid.UUID

	// Cost is the printed payment ("{U}", "{3}{U}{U}").
	Cost string

	// Question is the prompt header.
	Question string

	// OnDecline is the "or else": sacrifice the permanent, lose the
	// game. It runs on "no" and on a "yes" the chooser cannot fund,
	// exactly as PayUnless's does.
	OnDecline func(ctx *Context) error
}

func (p UpkeepPayUnless) Apply(ctx *Context) error {
	return ctx.Game.QueueUpkeepPayUnlessForEffect(game.UpkeepPayUnlessPrompt{
		Chooser:   p.Chooser,
		Source:    ctx.Source(),
		Cost:      p.Cost,
		Question:  p.Question,
		OnDecline: declineAgainstTheSameItem(ctx.Item, p.OnDecline),
	})
}

// CounterUnlessPaid is "counter <StackID> unless its controller pays
// <Cost>" — Daze, Dazzling Denial, Izzet Charm's first mode, Mystic
// Confluence's first mode, Spell Stutter, and ward's mana leg.
//
// USE THIS RATHER THAN PayUnless WITH A CounterTarget DECLINE. It is
// the same CR 118.12 prompt, and the difference is the whole of #951:
// a pay-unless whose decline counters an object on the stack must
// stop the table while it is unanswered, because resolving that
// object answers the question by doing it. The engine derives the
// halt from the guarded object (Game.QueueCounterUnlessPaidForEffect,
// game/counter_unless_paid.go); a card that hand-rolls the shape out
// of PayUnless gets the Rhystic Study latitude instead and the spell
// resolves for free. Six cards did exactly that before this existed.
//
// It also absorbs the two checks every one of those six spelled out:
// an object that has already left the stack raises no prompt at all
// (there is nothing to counter and so nothing to charge for), and the
// decline re-checks before countering.
type CounterUnlessPaid struct {
	// StackID is the object to counter — the "that spell".
	StackID uuid.UUID

	// Cost is the printed payment ("{1}", "{2}").
	Cost string

	// Question is the prompt header.
	Question string

	// Chooser overrides "its controller", which is what the engine
	// reads off the guarded object when this is left zero. Ward sets
	// it: CR 702.21a asks the player who cast the spell that targeted
	// the warded permanent, and the trigger captured that player when
	// the targeting event fired.
	Chooser uuid.UUID
}

func (c CounterUnlessPaid) Apply(ctx *Context) error {
	return ctx.Game.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
		StackItem: c.StackID,
		Chooser:   c.Chooser,
		Source:    ctx.Source(),
		Cost:      c.Cost,
		Question:  c.Question,
	})
}

// EachPlayerSacrifices is "each player sacrifices a creature" (Fleshbag
// Marauder), "each other player sacrifices a creature" (Grave Pact) or
// "each opponent sacrifices a creature" (Butcher of Malakir) — CR
// 701.21a.
//
// Every affected player chooses their own, so this fans out one prompt
// per player rather than picking for them; that is the whole rules
// content of the card, and it is why the effect is not targeted. A
// creature with hexproof or protection is still a legal choice, and
// the effect resolves whether or not anyone has a creature.
//
// Set ExceptController for "each OTHER player" / "each opponent".
// Leave Match nil for "a permanent"; pass Creature() for the usual
// "a creature".
type EachPlayerSacrifices struct {
	// ExceptController skips the effect's controller — the
	// difference between Grave Pact ("each other player") and
	// Fleshbag Marauder ("each player", you included).
	ExceptController bool

	// Match narrows what may be chosen. Nil means any permanent.
	Match CardPredicate

	// Label is the picker's banner copy: "a creature".
	Label string

	// Then is the clause printed after the edict — "if you sacrificed
	// a creature this way, …", "then you draw a card". Optional; leave
	// it nil for an edict with nothing hanging off it, which is most
	// of them.
	//
	// #1019, and the same shape as SacrificePermanent.Then (#993). The
	// prompts are a QUESTION per seat and return how many seats were
	// asked, so a clause written on the next line pays out before
	// anybody has chosen anything. This one runs once every asked seat
	// has answered AND the permanents they named have finished moving
	// — so a sacrificed commander's CR 903.9 prompt holds it too.
	//
	// `sacrificed` carries one entry per seat that was ASKED, in APNAP
	// ask order, with the permanents that really left the battlefield
	// (game.PromptedSacrifices). A commander that took the command
	// zone is in it; a leg the CR 614 window cancelled is not. Write
	// the clause as something that acts on what it is told.
	Then func(ctx *Context, sacrificed game.PromptedSacrifices) error
}

func (e EachPlayerSacrifices) Apply(ctx *Context) error {
	except := uuid.Nil
	if e.ExceptController {
		except = ctx.Controller()
	}
	var spec *game.TargetSpec
	if e.Match != nil {
		spec = sacrificeSpec(e.Label, e.Match)
	}
	label := e.Label
	if label == "" {
		label = "a permanent"
	}
	if e.Then == nil {
		ctx.Game.EachPlayerSacrificesForEffect(ctx.Source(), except, spec, "Sacrifice "+label)
		return nil
	}
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract massEffect.apply explains: an undo restores
	// this game's fields in place, so a captured *Game would be the
	// wrong one.
	item := ctx.Item
	return ctx.Game.EachPlayerSacrificesThenForEffect(ctx.Source(), except, spec, "Sacrifice "+label,
		func(g *game.Game, sacrificed game.PromptedSacrifices) error {
			return e.Then(NewContext(g, item), sacrificed)
		})
}

// Scry is "scry N" (CR 701.22) — look at the top N cards of your
// library, then put any number on the bottom and the rest back on top
// in any order.
//
// The whole effect is a choice, so this queues a prompt rather than
// doing anything to the library: nothing moves until the player
// answers. A scry with an empty library is not an error and queues
// nothing.
//
// Scry is "look at", not "reveal" — only the scrying player sees the
// cards. The engine handles that; a card's effect never needs to.
type Scry struct {
	Player uuid.UUID
	N      int

	// Then is the rest of the effect, for a card whose text says
	// "Scry N, THEN ..." (Preordain: "Scry 2, then draw a card"). It
	// runs once the player has put the cards back, so the card left on
	// top is the card drawn.
	//
	// Anything after "then" MUST go here rather than after this
	// primitive returns. Apply only queues the prompt, so a draw
	// written as the next statement happens BEFORE the player has
	// chosen — a different card, and it also leaves the prompt
	// unanswerable, because the drawn card is no longer in the library
	// for the reorder to put back.
	Then func(g *game.Game) error
}

func (s Scry) Apply(ctx *Context) error {
	player := s.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ctx.Game.ScryThenForEffect(player, ctx.Source(), s.N, s.Then)
	return nil
}

// Surveil is "surveil N" (CR 701.25) — look at the top N cards of
// your library, then put any number of them into your graveyard and
// the rest back on top in any order.
//
// Scry with the bottom-of-library leg replaced by the graveyard, and
// that replacement is the entire reason the keyword exists: a card
// binned by a surveil is somewhere you can reanimate, delve or escape
// it from, not buried under the rest of your deck.
//
// Like Scry, the whole effect is a choice, so Apply queues a prompt
// and nothing moves until the player answers. A surveil with an empty
// library is not an error and queues nothing.
//
// Surveil is "look at", not "reveal" — only the surveilling player
// sees the cards. The engine handles that; a card's effect never
// needs to.
type Surveil struct {
	Player uuid.UUID
	N      int

	// Then is the rest of the effect, for a card whose text says
	// "Surveil N, THEN ...". It runs once the player has put the
	// cards back, so it sees the library and graveyard the player
	// chose.
	//
	// The same warning as Scry.Then applies, for the same reason:
	// anything after "then" MUST go here rather than after this
	// primitive returns. Apply only queues the prompt, so a draw
	// written as the next statement happens BEFORE the player has
	// chosen — and also leaves the prompt unanswerable, because the
	// drawn card is no longer in the library to put back.
	Then func(g *game.Game) error
}

func (s Surveil) Apply(ctx *Context) error {
	player := s.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ctx.Game.SurveilThenForEffect(player, ctx.Source(), s.N, s.Then)
	return nil
}

// LookAtTop is "look at the top N cards of your library, then put
// them back in any order" — Ponder, Sensei's Divining Top,
// Soothsaying.
//
// The scry family's third member and the one with no away lane: every
// card goes back on top, and the only decision is the order. That
// still has to be a real prompt, because a card whose text is "put
// them back in any order" and whose implementation puts them back in
// the order they were is a blank.
//
// Like Scry and Surveil, Apply only queues the prompt — nothing moves
// until the player answers, and an empty library queues nothing.
type LookAtTop struct {
	Player uuid.UUID
	N      int

	// Then is the rest of the effect, for "... then draw a card"
	// (Ponder). Same warning as Scry.Then: it MUST go here, not on
	// the line after Apply, or the draw happens before the player has
	// decided which card is on top — which is the entire point of
	// the card.
	Then func(g *game.Game) error
}

func (l LookAtTop) Apply(ctx *Context) error {
	player := l.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ctx.Game.LookAtTopThenForEffect(player, ctx.Source(), l.N, l.Then)
	return nil
}

// ShuffleLibrary is "shuffle your library" as an instruction in its
// own right — Soothsaying's {3}{U}{U}, Blood Moon-era library
// resets — rather than the "then shuffle" tail of a search, which
// rides SearchLibrary.Shuffle.
//
// Not a no-op even when nothing was searched for: shuffling is how a
// player answers an opponent's Sensei's Divining Top or their own
// bad scry, and the engine clears every KnownBy in the zone, so the
// knowledge really does dissolve.
type ShuffleLibrary struct {
	// Player is whose library is shuffled. Zero means the
	// controller of the effect.
	Player uuid.UUID
}

func (s ShuffleLibrary) Apply(ctx *Context) error {
	player := s.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.ShuffleLibraryForEffect(player)
}

// MillToZone is MillCards generalised: it moves cards off the top of
// a library into a destination zone and hands the caller back what
// moved.
//
// Three things it can express that MillCards cannot:
//
//   - "Exile the top N cards of your library" — To: game.ZoneExile.
//     That is not a mill, and the engine emits an ordinary zone move
//     rather than EventMill for it, so mill payoffs stay out of it.
//   - "... until a creature card is put into their graveyard" — Until
//     ends the run after the first card it accepts, and the card that
//     ends it still moves.
//   - "... then do something with the cards milled this way" — Then
//     receives them in the order they came off the library. Diffing
//     the graveyard afterwards would be wrong the moment anything
//     else put a card there during the same resolution.
//
// The zero value of To is ZoneGraveyard, so MillToZone{Player: p, N:
// 3} is exactly MillCards.
type MillToZone struct {
	Player uuid.UUID
	N      int

	// To is ZoneGraveyard (the default, an ordinary mill) or
	// ZoneExile. Anything else is rejected rather than guessed at.
	To game.ZoneKind

	// Until, when set, ends the run after the first LANDED list it
	// returns true for. With Until set, N <= 0 means "no limit but the
	// library", which is how an unbounded mill is written.
	//
	// `landed` is the cards that have actually reached To so far, top
	// of the library first, and it is never empty when the clause is
	// asked. #1159: a card the CR 614 window sent somewhere else — a
	// commander taking the command zone (CR 903.9), "if a card would
	// be put into a graveyard from anywhere, exile it instead" — was
	// never put into To, so it is not in the list and does not end the
	// run. Same reading as Then's `milled`, because it is the same
	// rule (CR 400.7).
	//
	// Write it as a question about the WHOLE list, not as an
	// accumulator over successive calls: the engine may ask it again
	// for the same prefix when an undo replays the answer to a CR
	// 903.9 prompt, and a closure counting as it goes would be wrong
	// the second time. UntilCard and UntilTotalManaValue below are the
	// two shapes the catalog needs; reach for them first.
	Until func(landed []game.Card) bool

	// Then is the "for each card milled this way" clause, and `milled`
	// holds the instance IDs that actually reached To, in library
	// order (top first). Optional; leave it nil for a mill with
	// nothing hanging off it.
	//
	// #893: it runs from a CONTINUATION, for the reason
	// ExileTarget.Then and DestroyAllMatching.Then do — a mill opens
	// the CR 614 window per card, so a commander coming off the top
	// stops to answer CR 903.9 and what was milled is not knowable on
	// the next line. `milled` is CR 400.7's reading: a card a
	// replacement sent somewhere else (the command zone, or exile
	// under "if a card would be put into a graveyard from anywhere,
	// exile it instead") is not in it, however thoroughly it left the
	// library. Write the clause as something that acts on what it is
	// told, not as the next line of the card.
	Then func(ctx *Context, milled []uuid.UUID) error
}

func (m MillToZone) Apply(ctx *Context) error {
	if m.N <= 0 && m.Until == nil {
		if m.Then == nil {
			return nil
		}
		// A mill of nothing is still an answer. A caller sequencing
		// several mills through the continuation has to be told, or it
		// waits forever — the rule runRouteTailLocked follows for every
		// terminal outcome of a routed move.
		return m.Then(ctx, nil)
	}
	dest := m.To
	if dest == "" {
		dest = game.ZoneGraveyard
	}
	player := m.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	if m.Then == nil {
		// Nothing is waiting on the list, so the mill stays
		// fire-and-forget: every card is routed on this line and a
		// commander's CR 903.9 prompt lands its own card later without
		// holding the rest of the mill up.
		_, err := ctx.Game.MillToZoneForEffect(player, m.N, dest, m.Until)
		return err
	}
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract massEffect.apply explains: an undo restores
	// this game's fields in place, so a captured *Game would be the
	// wrong one.
	item := ctx.Item
	return ctx.Game.MillToZoneThenForEffect(player, m.N, dest, m.Until,
		func(g *game.Game, milled []uuid.UUID) error {
			return m.Then(NewContext(g, item), milled)
		})
}

// UntilCard is MillToZone.Until for the common clause: the run ends on
// the card that arrived, judged on its own — "until a creature card is
// put into their graveyard", "until they reveal a land card".
//
// Pure by construction: it reads only the last entry of the list it is
// handed, so replaying it over the same prefix gives the same answer.
// See MillToZone.Until for why that matters.
func UntilCard(pred func(c game.Card) bool) func([]game.Card) bool {
	return func(landed []game.Card) bool {
		return len(landed) > 0 && pred(landed[len(landed)-1])
	}
}

// UntilTotalManaValue is MillToZone.Until for the running-total
// clause: the run ends once the cards that LANDED total `threshold`
// mana value or more — "until you exile cards with total mana value 4
// or greater" (Improvisation Capstone, Echocasting Symposium).
//
// The total is recomputed from the whole landed list every time rather
// than accumulated across calls, which is what makes it pure and what
// makes a card the CR 614 window diverted contribute nothing: it never
// arrived, so it is not in the list (#1159).
func UntilTotalManaValue(threshold int) func([]game.Card) bool {
	return func(landed []game.Card) bool {
		total := 0
		for _, c := range landed {
			total += c.ManaValue()
		}
		return total >= threshold
	}
}

// ExileTopFaceDown is "exile the top N cards of your library face
// down" (CR 406.3) — Necropotence.
//
// A sibling of MillToZone{To: game.ZoneExile} and deliberately not a
// flag on it, because the two differ in the thing that matters about
// an exile: who can read the card. An ordinary exile is public, so
// the engine marks every seat a knower and the wire ships the name.
// A face-down exile is readable by nobody, so the card's knowledge
// set is cleared and game.Card.FaceDown is set — the client draws a
// card back, and the controller finds out what they bought when the
// card reaches their hand, exactly as in paper.
//
// Exiling from an empty library moves nothing. It is NOT a draw, so
// it does not set up the CR 704.5b loss the way running the library
// out to a draw does.
type ExileTopFaceDown struct {
	// Player is whose library is exiled from. Zero means the
	// controller of the effect.
	Player uuid.UUID

	// N is how many cards come off the top.
	N int

	// Exiled, when non-nil, is filled with the instance IDs that
	// moved, top card first. Necropotence needs them to hand to the
	// delayed trigger that puts them into hand later; diffing the
	// exile zone afterwards would be wrong the moment anything else
	// exiled a card during the same resolution.
	Exiled *[]uuid.UUID
}

func (e ExileTopFaceDown) Apply(ctx *Context) error {
	if e.N <= 0 {
		return nil
	}
	player := e.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	moved, err := ctx.Game.ExileTopFaceDownForEffect(player, e.N)
	if e.Exiled != nil {
		*e.Exiled = moved
	}
	return err
}

// RevealCards is "reveal" (CR 701.20): show the named cards to every
// player at the table, and let them all remember it.
//
// The counterpart to Scry / Surveil / LookAtTop, and the difference
// is the whole reason it is a separate primitive rather than a flag.
// Those three are LOOK AT — the engine marks the chooser alone a
// knower and the wire redacts the cards for every other seat. This
// one marks every seat, and additionally announces the reveal on
// GameView.Reveals, so the other players are TOLD it happened rather
// than left to notice a card had quietly become readable.
//
// Nothing moves. A reveal is not a zone change, so "reveal the top
// card of your library and put it into your hand" is this primitive
// followed by BounceToHand, in that order — which is also what makes
// the table see the card in the zone it was revealed from.
//
// Cards that are no longer findable are skipped rather than erroring:
// a reveal is a look, and a look at something that has left is
// nothing.
type RevealCards struct {
	// Player is whose cards are shown. Zero means the controller of
	// the effect.
	Player uuid.UUID

	// Cards are the instance IDs to reveal, in the order the table
	// should see them.
	Cards []uuid.UUID

	// Reason is the one-line label the client banner shows, written
	// the way the card is written — "Fact or Fiction — reveal the top
	// five cards of your library".
	Reason string
}

func (r RevealCards) Apply(ctx *Context) error {
	if len(r.Cards) == 0 {
		return nil
	}
	player := r.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ctx.Game.RevealForEffect(game.RevealSpec{
		Player: player,
		Source: ctx.Source(),
		Reason: r.Reason,
		Cards:  r.Cards,
	})
	return nil
}

// RevealTopOfLibrary is "reveal the top N cards of your library" —
// Dark Confidant's upkeep flip, and the half of Fact or Fiction that
// happens before anybody has to make a decision.
//
// The cards stay on the library; Revealed hands back what they were
// so the rest of the card's text can act on them. Re-reading the
// library afterwards to find out would be wrong the moment anything
// else touched it during the same resolution, and once the card has
// moved there is no other way to name it.
//
// A short library reveals what it has, an empty one reveals nothing,
// and neither is an error. Revealing is not drawing: running the
// library out this way does not set up the CR 704.5b loss.
type RevealTopOfLibrary struct {
	// Player is whose library is revealed from. Zero means the
	// controller of the effect.
	Player uuid.UUID

	// N is how many cards come off the top. Non-positive is a no-op.
	N int

	// Reason is the banner label — see RevealCards.Reason.
	Reason string

	// Revealed, when non-nil, is filled with the instance IDs that
	// were revealed, top card first.
	Revealed *[]uuid.UUID
}

func (r RevealTopOfLibrary) Apply(ctx *Context) error {
	if r.N <= 0 {
		return nil
	}
	player := r.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ids := ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), r.N, r.Reason)
	if r.Revealed != nil {
		*r.Revealed = ids
	}
	return nil
}

// SacrificeThisIfStillOnBattlefield is the trigger body for the
// commonest half-sentence in the format: "…, sacrifice it."
//
// It is a function rather than a primitive struct because it takes
// nothing — the permanent doing the sacrificing is the trigger's own
// source, read off the resolving item.
//
// The battlefield check is load-bearing rather than defensive. A
// trigger resolves after everything that was put on the stack above
// it, so between "when this becomes the target" and this body the
// permanent can have been bounced, exiled or destroyed; CR 701.21a
// says a sacrifice does nothing at all for a permanent its controller
// no longer controls, and a resolution that cannot do its job must
// not wedge the stack.
func SacrificeThisIfStillOnBattlefield(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
