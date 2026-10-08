package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// suspect.go — the card-facing half of CR 701.60, suspect (ADR 0071
// amendment 2026-10-08, #2698). The engine half is game/suspect.go:
// the Card.Suspected designation, the layer-6 menace grant and the
// can't-block restriction it carries.
//
// A card file writes one of three things:
//
//	Suspect{Target: ctx.Source()}.Apply(ctx)       // "suspect it" / "suspect this creature"
//	Unsuspect{Target: id}.Apply(ctx)                 // "it's no longer suspected"
//	UnsuspectAll{Controlled: …}.Apply(ctx)           // "all suspected creatures are no longer suspected"
//
// and, for a targeted clause, SuspectEachLegalTarget as the Effect of
// "suspect up to one target creature" (TargetCreature(…).WithCount(0, 1)).
// "If it's suspected" is the Suspected() predicate on a target clause
// or the SuspectedPermanent helper on a resolved object.

// Suspected passes for a card that is currently suspected (CR 701.60).
// Reads the designation straight off the card, which the layer pass
// never rewrites, so it is right at announce and at resolution alike.
func Suspected() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Suspected }
}

// NotSuspected is Not(Suspected()): "target creature that isn't
// suspected".
func NotSuspected() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !c.Suspected }
}

// Suspect is "suspect [a creature]" (CR 701.60a). A target that has
// left the battlefield, is not a creature, or is already suspected
// (CR 701.60d: it can't become suspected again) is skipped, never an
// error.
type Suspect struct {
	Target uuid.UUID
}

func (s Suspect) Apply(ctx *Context) error {
	if ctx.isNewSourceObject(s.Target) { // #1432
		return nil
	}
	ctx.Game.SuspectForEffect(s.Target)
	return nil
}

// Unsuspect is "[it] is no longer suspected" for one permanent.
type Unsuspect struct {
	Target uuid.UUID
}

func (u Unsuspect) Apply(ctx *Context) error {
	if ctx.isNewSourceObject(u.Target) { // #1432
		return nil
	}
	ctx.Game.UnsuspectForEffect(u.Target)
	return nil
}

// UnsuspectAll is "all suspected creatures are no longer suspected"
// (Absolving Lammasu) and its narrower forms ("creatures your
// opponents control" — Eliminate the Impossible). With no Match it
// reaches every suspected permanent on the battlefield; Match narrows
// it, and is judged against the card as it stands now.
type UnsuspectAll struct {
	Match CardPredicate
}

func (u UnsuspectAll) Apply(ctx *Context) error {
	g := ctx.Game
	// Collect first: unsuspecting rewrites the battlefield's cards in
	// place, and the predicate must see them all as they were.
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.Suspected {
			continue
		}
		if u.Match != nil && !u.Match(g, ctx.Controller(), c) {
			continue
		}
		ids = append(ids, c.InstanceID)
	}
	for _, id := range ids {
		g.UnsuspectForEffect(id)
	}
	return nil
}

// SuspectEachLegalTarget is the Effect of a clause that suspects what
// it targets — "suspect up to one target creature". Each target still
// legal at resolution is suspected (CR 608.2b); one that left, or that
// was already suspected, is skipped.
func SuspectEachLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (Suspect{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// suspectEnchantedCreature is "When this Aura enters, suspect enchanted
// creature" (Convenient Target, Incriminating Impetus). It names the
// creature through the Aura's attachment as the trigger resolves, or
// as it last was if the Aura has already gone (CR 608.2h), so
// "enchanted creature" still means the host.
func suspectEnchantedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	aura, ok := g.PermanentForEffect(ref)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	return Suspect{Target: aura.AttachedTo.ID}.Apply(ctx)
}

// SacrificeASuspectedCreature is the cost "Sacrifice a suspected
// creature" (Rune-Brand Juggler). A sacrifice cost does not target, so
// hexproof never applies; the source itself qualifies when it is a
// suspected creature, as it does for any sacrifice outlet.
func SacrificeASuspectedCreature() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a suspected creature", Creature(), Suspected())}
}

// UpToOneTargetCreature is the "up to one target creature" clause the
// suspect ETBs share.
func UpToOneTargetCreature(label string, preds ...CardPredicate) *game.TargetSpec {
	return TargetCreature(label, preds...).WithCount(0, 1)
}
