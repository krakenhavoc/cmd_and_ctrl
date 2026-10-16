package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// convert.go — More Than Meets the Eye (CR 702.162), convert (CR
// 701.28) and living metal (CR 702.161): the Transformers mechanics,
// ADR 0152.
//
//	MoreThanMeetsTheEye("{W}{U}")   the front face's offer: cast the card
//	                                converted, back face up, for that
//	                                price, from wherever it can be cast.
//	Convert / ConvertThis           the permanent turns over IN PLACE,
//	                                exactly as Transform / TransformThis
//	                                do (CR 701.28a), and the log says
//	                                "converted".
//	LivingMetal()                   the back face's static: an artifact
//	                                creature during its controller's turn.
//
// "Return it to the battlefield converted" is NOT ConvertThis: that is a
// new object entering back face up (CR 712.14a), which is
// ReturnFromGraveyard{Transformed: true}. A convert is the same object
// turning over (CR 712.18): counters, damage and attachments stay and
// nothing enters.

// MoreThanMeetsTheEye is "More Than Meets the Eye [cost] (You may cast
// this card converted for [cost].)" — CR 702.162a, ADR 0152 §1.
//
// Disturb's offer without the zone: CastsFace 1 puts the card on the
// stack back face up (CR 712.11a) with its front face's mana value
// (CR 712.8c), and the permanent enters back face up (CR 712.13).
// FromZone is left empty, so the offer is claimable wherever the card
// may be cast — the hand, the command zone, exile under a permission
// that does not insist on its own price — which is CR 702.162a's
// "functions in any zone from which the spell may be cast". It is an
// alternative cost (CR 601.2b, 601.2f–h): it can't be combined with
// another one, and additional costs and the commander tax still apply.
//
// Declared on the FRONT face's entry (the bare oracle ID):
//
//	AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{W}{U}")},
func MoreThanMeetsTheEye(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:       "more_than_meets_the_eye",
		Label:     "More Than Meets the Eye " + cost,
		ManaCost:  cost,
		CastsFace: 1,
	}
}

// Convert converts the named permanent (CR 701.28a).
//
// A permanent that cannot — not a double-faced permanent (CR 701.28c),
// one whose other face is an instant or sorcery (CR 701.28d), one that
// can't transform (CR 701.28f) — does NOTHING, and that is not an
// error. No printed card converts another permanent (ADR 0152, out of
// scope 2); this exists so ConvertThis has the shape TransformThis has.
type Convert struct {
	Target uuid.UUID
}

func (c Convert) Apply(ctx *Context) error {
	if c.Target == uuid.Nil || ctx.isNewSourceObject(c.Target) { // #1432
		return nil
	}
	// CR 701.28a converts a permanent. A card that has left the
	// battlefield has nothing to turn over, so the instruction does
	// nothing (CR 608.2b): Ratchet, Field Medic's lifelink trigger
	// resolving after Ratchet died in the same combat, or Rescue Racer's
	// own trigger after it died beside another artifact.
	if z := ctx.Game.FindCardZoneForEffect(c.Target); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return ctx.Game.ConvertPermanentForEffect(c.Target)
}

// ConvertThis converts the SOURCE permanent — "convert Goldbug", the
// clause on every Transformers card.
//
// CR 701.28e: an ability of a permanent that tries to convert it does
// so only if the permanent hasn't converted or transformed since the
// ability was put onto the stack (for a delayed trigger, since it was
// created); "if the permanent has already transformed or converted, an
// instruction to do either is ignored". Two "convert it" triggers from
// one double-strike hit turn it over once (ADR 0152 §3).
type ConvertThis struct{}

func (ConvertThis) Apply(ctx *Context) error {
	if sourceTurnedOverSince(ctx) {
		return nil
	}
	return Convert{Target: ctx.Source()}.Apply(ctx)
}

// sourceTurnedOverSince is CR 701.27f / 701.28e for TransformThis and
// ConvertThis: the resolving ability's source has transformed or
// converted since the ability went on the stack (ADR 0152 §3).
func sourceTurnedOverSince(ctx *Context) bool {
	return ctx != nil && ctx.Game != nil && ctx.Game.SourceTurnedOverSinceForEffect(ctx.Item)
}

// LivingMetal is "Living metal (During your turn, this Vehicle is also
// a creature.)" — CR 702.161a, ADR 0152 §4.
//
// A layer-4 static on the face that prints it (the "<oracle_id>#1"
// entry): while its controller is the active player, the permanent is
// an artifact creature in addition to its other types. It sets no power
// or toughness: a Vehicle's printed power and toughness are its own and
// apply whenever it is a creature (CR 301.7a–b). The turn change bumps
// the layer version, so it is re-read on every turn with no state.
//
// Summoning sickness is about continuous control since the turn began
// (CR 302.6), so a Vehicle cast this turn can't attack without haste,
// and one already there that converted can.
func LivingMetal() game.StaticAbility {
	return game.StaticAbility{
		Label: "Living metal (During your turn, this Vehicle is also a creature.)",
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, t := range []string{"Artifact", "Creature"} {
				has := false
				for _, have := range c.Types {
					if have == t {
						has = true
						break
					}
				}
				if !has {
					c.Types = append(c.Types, t)
				}
			}
		},
	}
}
