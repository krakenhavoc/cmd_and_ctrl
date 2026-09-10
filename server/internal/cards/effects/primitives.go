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
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), g.Player, g.Amount)
}

// DrawCards draws N cards for `Player`. Non-positive N is a no-op.
// Drawing from an empty library flags the player as LosesAtNextSBA
// via the standard drawCardLocked path.
type DrawCards struct {
	Player uuid.UUID
	N      int
}

func (d DrawCards) Apply(ctx *Context) error {
	if d.N <= 0 {
		return nil
	}
	return ctx.Game.DrawNForEffect(d.Player, d.N)
}

// DiscardCards removes N cards from `Player`'s hand. Sandbox
// simplification: random selection regardless of whether the real
// card says "at random" vs. "the player chooses" — the target-
// picker UI lands in S22. Cards land in the graveyard known to
// every seated player (public-zone rule).
type DiscardCards struct {
	Player uuid.UUID
	N      int
}

func (d DiscardCards) Apply(ctx *Context) error {
	if d.N <= 0 {
		return nil
	}
	return ctx.Game.DiscardRandomForEffect(d.Player, d.N)
}

// MillCards mills the top N cards of `Player`'s library to their
// graveyard. Empty library mid-mill sets LosesAtNextSBA.
type MillCards struct {
	Player uuid.UUID
	N      int
}

func (m MillCards) Apply(ctx *Context) error {
	if m.N <= 0 {
		return nil
	}
	return ctx.Game.MillNForEffect(m.Player, m.N)
}

// DestroyTarget routes a battlefield permanent to its owner's
// graveyard. Wrath-style iteration composes this inside a
// `for _, id := range ctx.CreatureIDs()` loop in the card's
// OnResolve.
type DestroyTarget struct {
	Target uuid.UUID
}

func (d DestroyTarget) Apply(ctx *Context) error {
	return ctx.Game.DestroyPermanentForEffect(d.Target)
}

// SacrificePermanent sacrifices a battlefield permanent on behalf
// of its controller (CR 701.17). Distinct from DestroyTarget:
// sacrifice ignores indestructible / regeneration and fires
// EventSacrifice as well as the ordinary dies-trigger, which is
// what aristocrats payoffs watch. Added in S21 sub-PR 1.
type SacrificePermanent struct {
	Target uuid.UUID
}

func (s SacrificePermanent) Apply(ctx *Context) error {
	return ctx.Game.SacrificePermanentForEffect(s.Target)
}

// ExileTarget moves a card from whichever zone it's in to the
// shared exile zone. Works for battlefield permanents and hand
// / graveyard cards alike.
type ExileTarget struct {
	Target uuid.UUID
}

func (e ExileTarget) Apply(ctx *Context) error {
	return ctx.Game.ExileCardForEffect(e.Target)
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
type Flicker struct {
	Target     uuid.UUID
	Controller uuid.UUID
	Tapped     bool
}

func (f Flicker) Apply(ctx *Context) error {
	if err := (ExileTarget{Target: f.Target}).Apply(ctx); err != nil {
		return err
	}
	return ReturnFromExile{Target: f.Target, Controller: f.Controller, Tapped: f.Tapped}.Apply(ctx)
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
		Controller:   controller,
		SourceCardID: ctx.Source(),
		Label:        s.Label,
		At:           at,
		Cards:        s.Cards,
		Effect:       s.Effect,
	})
	return nil
}

// BounceToHand returns a card to its owner's hand. Used by
// Unsummon-style effects.
type BounceToHand struct {
	Target uuid.UUID
}

func (b BounceToHand) Apply(ctx *Context) error {
	return ctx.Game.BounceToHandForEffect(b.Target)
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
	return ctx.Game.CreateTokenForEffect(c.Controller, c.Template, c.N)
}

// ReturnFromGraveyard moves a card from a graveyard to `Dest`
// (typically ZoneHand, occasionally ZoneBattlefield for
// reanimation). Target must currently be in a graveyard.
type ReturnFromGraveyard struct {
	Target uuid.UUID
	Dest   game.ZoneKind
}

func (r ReturnFromGraveyard) Apply(ctx *Context) error {
	return ctx.Game.ReturnFromGraveyardForEffect(r.Target, r.Dest)
}

// SearchLibrary looks through `Player`'s library for up to `Limit`
// cards matching `Predicate`, moves them to `Dest`, optionally
// reveals them to all seated players (via KnownBy), and optionally
// shuffles the library afterwards. Sandbox simplification: picks
// the first match in library order (no "you choose" UI — deferred
// to S22). Effects that need a specific card pass a tight
// predicate (e.g. "basic land named Forest").
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
	// Only meaningful when Dest == ZoneBattlefield. Added in S17
	// sub-PR 4 to close the S14 "enters untapped" deferrals for
	// fetched-land cards.
	TappedOnEntry bool
}

func (s SearchLibrary) Apply(ctx *Context) error {
	return ctx.Game.SearchLibraryForEffectWithOptions(
		s.Player, s.Predicate, s.Dest, s.Limit, s.Reveal, s.Shuffle, s.TappedOnEntry,
	)
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

func (p PayUnless) Apply(ctx *Context) error {
	item := ctx.Item
	decline := p.OnDecline
	return ctx.Game.QueuePayUnlessForEffect(p.Chooser, ctx.Source(), p.Cost, p.Question,
		func(g *game.Game) error {
			if decline == nil {
				return nil
			}
			return decline(NewContext(g, item))
		})
}

// EachPlayerSacrifices is "each player sacrifices a creature" (Fleshbag
// Marauder), "each other player sacrifices a creature" (Grave Pact) or
// "each opponent sacrifices a creature" (Butcher of Malakir) — CR
// 701.17a.
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
	ctx.Game.EachPlayerSacrificesForEffect(ctx.Source(), except, spec, "Sacrifice "+label)
	return nil
}

// Scry is "scry N" (CR 701.18) — look at the top N cards of your
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
