package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saddle.go — Saddle N (CR 702.171), Outlaws of Thunder Junction's Mount
// keyword, and the vocabulary of the cards printed around it (#2695,
// ADR 0071 amendment 2026-10-08).
//
// A Mount is written as:
//
//	Activated: []ActivatedAbility{Saddle(2)},
//	Triggered: []game.TriggeredAbility{
//	    AttacksWhileSaddled("Gilded Ghoda — create a Treasure", effect),
//	},
//
// and that is the whole card, plus whatever else it prints.
//
//	Printed                                     Written
//	Saddle N                                    Saddle(n)
//	Whenever this attacks while saddled, …      AttacksWhileSaddled(label, effect)
//	Whenever this becomes saddled, …            WhenBecomesSaddled(label, effect)
//	As long as this is saddled, it has …        SaddledKeywords(kws...)
//	[target Mount] becomes saddled              SaddleTarget(g, id) / Saddled{...}.Apply
//	a creature that saddled it this turn        SaddlersOf(ctx, mount)
//
// The designation itself is game.Card.Saddled: until end of turn, gone
// when the Mount leaves. See game.SaddleForEffect.

// Saddle is "Saddle N" (CR 702.171a): "Tap any number of other untapped
// creatures you control with total power N or more: This permanent
// becomes saddled until end of turn. Saddle only as a sorcery."
//
// Crew's cost over OTHER creatures (game.AbilityCost.Saddle), at sorcery
// speed (CR 702.171a), with an effect that sets the designation and
// remembers who paid. Deliberately NOT composed with TapCost(): the
// creatures that tap are not the Mount, which stays untapped and can
// attack. Summoning sickness does not apply to the creatures tapped (the
// cost is not a {T} cost, CR 702.122b's reasoning), and the Mount itself
// needs no haste to be saddled.
//
// The label is the printed line without its reminder text, "Saddle 2",
// which is what TestAbilitiesMatchOracleText compares.
func Saddle(n int) ActivatedAbility {
	return ActivatedAbility{
		Label:        fmt.Sprintf("Saddle %d", n),
		Cost:         game.AbilityCost{Saddle: n},
		SorcerySpeed: true,
		Effect:       saddleEffect,
	}
}

// saddleEffect is the resolution of a saddle ability. #1432: a Mount
// that left and came back is a new object; the ability of the old one
// does not saddle the new one. The creatures that paid are on the
// item's payment record, as the objects they were.
func saddleEffect(g *game.Game, item *game.StackItem) error {
	if sourceIsNewObject(g, item) {
		return nil
	}
	var by []game.ObjectRef
	for _, t := range item.Paid.TappedOthers {
		by = append(by, game.ObjectRef{ID: t.ID, Epoch: t.Epoch})
	}
	g.SaddleForEffect(item.SourceCardID, by)
	return nil
}

// Saddled is the CR 702.171 gate: this ability exists while the Mount
// is saddled — "As long as this Mount is saddled, …".
func Saddled() game.Designation { return game.SaddledGate() }

// isSaddledSource is "while saddled" for a trigger's AppliesTo: the
// source is a battlefield permanent that is saddled right now.
func isSaddledSource(source *game.Card) bool { return source != nil && source.Saddled }

// AttacksWhileSaddled is "Whenever this creature attacks while saddled,
// …" (every Mount that cares). The condition is read when the attack is
// declared, off the live permanent: a Mount saddled this turn, before
// combat, as the rule requires (saddle is sorcery speed, so a Mount can
// only be saddled in a main phase or by a spell or ability).
//
// Set Targets on the result for a targeted trigger, exactly as on any
// other trigger.
func AttacksWhileSaddled(label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return attackDeclared(ev, source) && isSaddledSource(source)
		},
		Effect: effect,
	}
}

// WhenBecomesSaddled is "Whenever this Mount becomes saddled, …". The
// event fires only when the Mount was not already saddled, so "for the
// first time each turn" (Stubborn Burrowfiend) is this constructor.
func WhenBecomesSaddled(label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecameSaddled},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return source != nil && ev.CardID == source.InstanceID
		},
		Effect: effect,
	}
}

// SaddledKeywords is "As long as this Mount is saddled, it has
// [keywords]" — a layer-6 self-grant behind the saddled gate.
func SaddledKeywords(keywords ...string) game.StaticAbility {
	kws := append([]string(nil), keywords...)
	return game.StaticAbility{
		Layer:      game.Layer6Ability,
		ActiveWhen: Saddled(),
		AppliesTo:  selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			appendKeywordsTo(c, kws)
		},
	}
}

// BecomeSaddled is the primitive for "[target Mount] becomes saddled
// until end of turn" (Guidelight Matrix, Alacrian Armory, Kolodin): the
// designation with no saddlers, so "creatures that saddled it" is empty.
// A permanent that is not a Mount is left alone, which is how "becomes
// saddled if it's a Mount" (Alacrian Armory) is expressed.
type BecomeSaddled struct {
	// Target is the permanent. Zero means "the source of the ability".
	Target uuid.UUID
}

func (b BecomeSaddled) Apply(ctx *Context) error {
	target := b.Target
	if target == uuid.Nil {
		target = ctx.Source()
		if ctx.isNewSourceObjectAsThis(target) {
			return nil
		}
	} else if ctx.isNewSourceObject(target) {
		return nil
	}
	ctx.Game.SaddleForEffect(target, nil)
	return nil
}

// SaddlersOf is "the creatures that saddled [this Mount] this turn"
// (CR 702.171c): those tapped to pay for the saddle abilities that made
// it saddled, still on the battlefield as the same objects.
func SaddlersOf(ctx *Context, mount uuid.UUID) []uuid.UUID {
	return ctx.Game.SaddlersOf(mount)
}
