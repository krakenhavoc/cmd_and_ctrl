package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// speed.go — the card-side vocabulary for Aetherdrift's speed
// (CR 702.178, 702.179). ADR 0138, #2122.
//
// The engine owns the mechanic (game/speed.go): the start-your-engines
// state-based action, the inherent once-per-turn trigger, the cap at 4.
// A card declares two things and nothing else:
//
//	PrintedKeywords: []string{StartYourEngines},       // "Start your engines!"
//	Static: []game.StaticAbility{MaxSpeedSelfPump(1, 1)}, // "Max speed — this creature gets +1/+1"
//
// "Max speed — [ability]" means "as long as your speed is 4, this
// object has [ability]" (CR 702.178a). It is NOT an ADR 0071
// designation gate: that gate reads the object only, and max speed is
// the PLAYER's (ADR 0138 §5). So each constructor below puts one
// predicate, YouHaveMaxSpeed, in the ability's own game-aware hook —
// AppliesTo for a static, a trigger, a replacement or a cost modifier,
// Condition for an activated or mana ability. Wrap the ability exactly
// as it would be written without "Max speed —"; never write the check
// by hand, so every max-speed ability reads the same number the same
// way.
//
// Append new shapes here; never change what one that exists means.

// StartYourEngines is the keyword token for "Start your engines!"
// (CR 702.179a), for Spec.PrintedKeywords.
const StartYourEngines = game.KeywordStartYourEngines

// YourSpeed is player's speed, 0 while they have none (CR 702.179f) — "where X is
// your speed" (The Speed Demon, Point the Way, Momentum Breaker).
func YourSpeed(g *game.Game, player uuid.UUID) int {
	if g == nil {
		return 0
	}
	return g.SpeedOf(player)
}

// YouHaveMaxSpeed reports whether player has max speed (CR 702.179e).
func YouHaveMaxSpeed(g *game.Game, player uuid.UUID) bool {
	return g != nil && g.HasMaxSpeed(player)
}

// speedPlayerOf is the "you" of an ability printed on c: its controller
// on the battlefield, its owner anywhere else (CR 108.4a; the glossary's
// "Max Speed": "that permanent's controller (or that card's owner, if
// it isn't on the battlefield)").
func speedPlayerOf(c *game.Card) uuid.UUID {
	if c == nil {
		return uuid.Nil
	}
	if c.Controller != uuid.Nil {
		return c.Controller
	}
	return c.Owner
}

// MaxSpeedCondition is the bare ActivationCondition: "activate only
// while you have max speed", for an ability built by hand. The
// constructors below are the usual way in.
func MaxSpeedCondition() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return YouHaveMaxSpeed(g, controller)
	}
}

// MaxSpeedActivated is "Max speed — [activated ability]": the ability
// can be activated only while its activator has max speed. The check
// is ANDed with any condition the ability already carries, so a
// printed "Activate only as a sorcery" or "only once each turn" still
// holds. While the activator is short the view greys the row
// (condition_unmet), the engine refuses it and the enumerator does not
// offer it (ADR 0138 §5).
func MaxSpeedActivated(a ActivatedAbility) ActivatedAbility {
	a.Condition = AllConditions(MaxSpeedCondition(), a.Condition)
	return a
}

// MaxSpeedMana is MaxSpeedActivated for a mana ability: Muraganda
// Raceway's "Max speed — {T}: Add {C}{C}". The auto-tapper reads
// Condition, so it never plans one its controller can't activate.
func MaxSpeedMana(m ManaAbility) ManaAbility {
	m.Condition = AllConditions(MaxSpeedCondition(), m.Condition)
	return m
}

// MaxSpeedStatic is "Max speed — [static ability]": the static applies
// only while its source's controller has max speed. The speed is a
// layer input (game.EventSpeedChanged bumps the layer version), so the
// effective characteristics follow it the moment it changes.
func MaxSpeedStatic(s game.StaticAbility) game.StaticAbility {
	applies := s.AppliesTo
	s.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		if !YouHaveMaxSpeed(g, speedPlayerOf(source)) {
			return false
		}
		return applies == nil || applies(target, g, source)
	}
	return s
}

// MaxSpeedSelfPump is "Max speed — this creature gets +p/+t" (Gastal
// Raider, Nesting Bot, Walking Sarcophagus).
func MaxSpeedSelfPump(power, toughness int) game.StaticAbility {
	return MaxSpeedStatic(game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: thisPermanent,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	})
}

// MaxSpeedSelfKeywords is "Max speed — this creature has <keywords>"
// (Gastal Raider's menace, Burnout Bashtronaut's double strike): one
// layer-6 grant per keyword, each through KeywordGrant's dedupe.
func MaxSpeedSelfKeywords(keywords ...string) []game.StaticAbility {
	out := make([]game.StaticAbility, 0, len(keywords))
	for _, kw := range keywords {
		out = append(out, MaxSpeedStatic(KeywordGrant(thisPermanent, kw)))
	}
	return out
}

// thisPermanent is the "this creature" scope of a self static.
func thisPermanent(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target != nil && source != nil && target.InstanceID == source.InstanceID
}

// MaxSpeedTrigger is "Max speed — [triggered ability]": the ability
// triggers only while its source's controller has max speed. Checked
// as the event is read, before anything is asked, so a short player is
// never prompted for a target or a "you may".
func MaxSpeedTrigger(t game.TriggeredAbility) game.TriggeredAbility {
	applies := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if !YouHaveMaxSpeed(g, speedPlayerOf(source)) {
			return false
		}
		return applies == nil || applies(ev, source, lki, g)
	}
	return t
}

// MaxSpeedReplacement is "Max speed — [replacement effect]" (Vnwxt,
// Verbose Host's "if you would draw a card, draw two cards instead").
func MaxSpeedReplacement(r game.ReplacementEffect) game.ReplacementEffect {
	applies := r.AppliesTo
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !YouHaveMaxSpeed(g, speedPlayerOf(src)) {
			return false
		}
		return applies == nil || applies(ev, g, src)
	}
	return r
}

// MaxSpeedCostModifier is "Max speed — [cost modifier]" (Racers'
// Scoreboard's "spells you cast cost {1} less to cast"). The source's
// controller is the "you" of the modifier.
func MaxSpeedCostModifier(m game.CostModifier) game.CostModifier {
	applies := m.AppliesTo
	m.AppliesTo = func(q game.CostQuery) bool {
		src := q.Source
		if !YouHaveMaxSpeed(q.Game, speedPlayerOf(&src)) {
			return false
		}
		return applies == nil || applies(q)
	}
	return m
}
