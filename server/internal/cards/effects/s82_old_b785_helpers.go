package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b785_helpers.go — shared shapes of the Visions cards of S82 OLD
// batch 785 (#5631).

// RegenerateThisWithRider is the body of "Regenerate this creature. When it
// regenerates this way, <effect>" (Matopi Golem): one regeneration shield on
// the ability's source, marked as the one that carries the rider, so the
// EventRegenerated that spends it names it (Event.RegenerationRider) and a
// regeneration from any other source does not. A source that has left the
// battlefield gets nothing (CR 701.19b).
func RegenerateThisWithRider(g *game.Game, item *game.StackItem) error {
	if item.SourceCardID == uuid.Nil || NewContext(g, item).IsNewSourceObject(item.SourceCardID) {
		return nil
	}
	if err := g.RegenerateWithRiderForEffect(item.SourceCardID); err != nil {
		return nil
	}
	return nil
}

// WhenThisRegeneratesThisWay is "When it regenerates this way, <effect>": a
// trigger on the EventRegenerated of this permanent whose shield came from
// an ability that said so.
func WhenThisRegeneratesThisWay(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventRegenerated, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source != nil && ev.CardID == source.InstanceID && ev.RegenerationRider
	}, label, effect)
}

// DamageToCountersInstead is the replacement "If damage would be dealt to
// this creature, put that many <kind> counters on it instead" (Lichenthrope):
// a replacement, not a prevention (CR 614.1a), so it applies to damage that
// can't be prevented, and the damage is not dealt (CR 614.6). The amount is
// the damage as it would be dealt, and the counters go on through the counter
// window.
func DamageToCountersInstead(label, kind string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: false,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageTarget == src.InstanceID && ev.DamageAmount > 0
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			n := ev.DamageAmount
			ev.Cancel()
			_, err := g.AddCounterMustSettleNowForEffect(src.InstanceID, kind, n)
			return err
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// SacrificedSourceCounterDamage is "It deals damage equal to the number of
// <kind> counters on it to <the target>" for a source sacrificed to pay the
// cost (Magma Mine): the counters it had as it last existed (CR 608.2h), the
// damage dealt by it.
func SacrificedSourceCounterDamage(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		info, ok := ctx.SourcePermanent()
		if !ok {
			return nil
		}
		n := info.Counters[kind]
		ts := ctx.LegalTargets()
		if n <= 0 || len(ts) == 0 {
			return nil
		}
		ref, _ := ctx.SourceRef()
		return DealDamage{SourceObject: &ref, Target: ts[0].ID, Amount: n}.Apply(ctx)
	}
}

// ChimeraSacrificeAbility is the shared ability of the Visions Chimeras:
// "Sacrifice this creature: Put a +2/+2 counter on target Chimera creature.
// It gains <keyword>. (This effect lasts indefinitely.)" The keyword is
// granted with no stated duration (CR 611.2a), pinned to the object it was
// given to (CR 400.7).
func ChimeraSacrificeAbility(keyword string) ActivatedAbility {
	return ActivatedAbility{
		Label:   "Sacrifice this creature: Put a +2/+2 counter on target Chimera creature. It gains " + keyword + ". (This effect lasts indefinitely.)",
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    SacrificeThis(),
		Targets: TargetCreature("target Chimera creature", HasSubtype("Chimera")),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			id := ts[0].ID
			if err := (AddCounter{Target: id, Kind: "+2/+2", N: 1}).Apply(ctx); err != nil {
				return err
			}
			return ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddKeywordsMod(keyword)},
				Duration: game.IndefiniteDuration(),
				Label:    "Chimera — gains " + keyword + " indefinitely",
			}.Apply(ctx)
		},
	}
}
