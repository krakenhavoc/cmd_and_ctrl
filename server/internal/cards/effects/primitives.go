package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// primitives.go declares the 15 composable effect primitives. Each
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

// ExileTarget moves a card from whichever zone it's in to the
// shared exile zone. Works for battlefield permanents and hand
// / graveyard cards alike.
type ExileTarget struct {
	Target uuid.UUID
}

func (e ExileTarget) Apply(ctx *Context) error {
	return ctx.Game.ExileCardForEffect(e.Target)
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
}

func (s SearchLibrary) Apply(ctx *Context) error {
	return ctx.Game.SearchLibraryForEffect(s.Player, s.Predicate, s.Dest, s.Limit, s.Reveal, s.Shuffle)
}
