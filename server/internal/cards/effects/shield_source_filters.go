package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shield_source_filters.go — the card-facing half of #2026 (ADR 0108 §7,
// amendment of 2026-10-05): shields that name their sources by what they
// aren't, by power, by combat status, by counters, by being colourless
// or by who controls them. The engine half is game/shield_source_filter.go.
//
// Append-only, mechanic-named. Writing a card:
//
//	combatShieldAgainstCreatures(game.DamageSourceFilter{PowerBounded: true, PowerAtMost: 3})
//	  — Fog of War's "combat damage … by creatures with power 3 or less"
//	combatShieldAgainstCreatures(game.DamageSourceFilter{Except: []game.PermanentQuery{{Subtypes: []string{"Spider"}}}})
//	  — Arachnogenesis' "by non-Spider creatures"
//	PreventDamageFromSource{Protect: ShieldYou, Queries: creatureSources(), Filter: game.DamageSourceFilter{Combat: game.SourceCombatAttacking}}
//	  — Heavy Fog's "all damage … to you … by attacking creatures"
//
// Every filter is read as the damage would be dealt (CR 609.7b), never
// as the shield resolves.

// creatureSources is the query "a creature": the positive half of every
// "creatures …" shield.
func creatureSources() []game.PermanentQuery {
	return []game.PermanentQuery{QueryTypes("creature")}
}

// combatShieldAgainstCreatures is "Prevent all combat damage that would
// be dealt this turn by creatures <filter>": combat damage only, to
// anything, from a creature passing the filter.
func combatShieldAgainstCreatures(f game.DamageSourceFilter) PreventDamageFromSource {
	return PreventDamageFromSource{CombatOnly: true, Protect: ShieldAnything, Queries: creatureSources(), Filter: f}
}

// exceptSubtypes is "non-<subtype>" or "other than <subtypes>": a source
// with any of the subtypes is left alone. A changeling has every creature
// type, so it is never "non-Spider" (CR 702.73a).
func exceptSubtypes(subtypes ...string) game.DamageSourceFilter {
	return game.DamageSourceFilter{Except: []game.PermanentQuery{{Subtypes: subtypes}}}
}

// creaturesAttackingPlayer counts the creatures attacking `player`
// itself — not their planeswalkers or battles: Arachnogenesis' "the
// number of creatures attacking you", read as it resolves.
func creaturesAttackingPlayer(g *game.Game, player uuid.UUID) int {
	if player == uuid.Nil {
		return 0
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget == player {
			n++
		}
	}
	return n
}

// attackedThisStepLabel is Deep Wood's and Heavy Fog's cast restriction
// as printed.
const attackedThisStepLabel = "Cast this spell only during the declare attackers step and only if you've been attacked this step."

// youWereAttackedThisStep is that restriction: the declare attackers
// step, in which a creature was declared attacking you — you, not a
// planeswalker you control (Heavy Fog's ruling). A creature removed
// from combat since was still declared, so you have still been attacked;
// the record is the turn's attack declarations in this combat phase
// (TurnTally.Attacks, CR 508.1).
func youWereAttackedThisStep(g *game.Game, controller uuid.UUID, _ game.Card) bool {
	if g.Turn.Step != game.StepDeclareAttackers || controller == uuid.Nil {
		return false
	}
	for _, a := range g.TurnTally.Attacks {
		if a.PhaseID == g.Turn.PhaseID && a.Defender == controller {
			return true
		}
	}
	return false
}

// damageToYouFromAttackers is "Prevent all damage that would be dealt to
// you this turn by attacking creatures" (Heavy Fog, Deep Wood).
func damageToYouFromAttackers() PreventDamageFromSource {
	return PreventDamageFromSource{Protect: ShieldYou, Queries: creatureSources(),
		Filter: game.DamageSourceFilter{Combat: game.SourceCombatAttacking}}
}

// --- follow-ups that ask what the prevented source was (#2026) --------
//
// The follow-up item's trigger event carries the source's card types as
// it would have dealt the damage (Event.LastKnownTypes), and with
// PreventDamageFromSource.ThenPerSource there is one application per
// source, so "that creature" names one creature.

var (
	// "If damage from a creature source is prevented this way, <this>
	// deals that much damage to that creature. If damage from a
	// noncreature source is prevented this way, <this> deals that much
	// damage to the source's controller" (Comeuppance).
	preventedBackAtTheSourceBody = game.DelayedBody("prevention/back-at-the-creature-or-its-controller", preventedBackAtTheSource)

	// "Whenever damage from a creature is prevented this way, each
	// commander creature you control deals damage equal to its power to
	// that creature" (Judgment of Alexander): a triggered ability, so it
	// goes on the stack (Samite Ministration's shape), with the creature
	// as its payload.
	preventedCommandersStrikeTriggerBody = game.DelayedBody("prevention/commanders-strike-that-creature-trigger", preventedCommandersStrikeTrigger)
	commandersStrikeThatCreatureBody     = game.DelayedBody("prevention/commanders-strike-that-creature", commandersStrikeThatCreature)
)

// preventedFromACreature reports whether the prevented damage came from a
// creature, as it was when it would have dealt it.
func preventedFromACreature(item *game.StackItem) bool {
	if item == nil || item.Trigger == nil {
		return false
	}
	for _, t := range item.Trigger.Event.LastKnownTypes {
		if strings.EqualFold(t, "creature") {
			return true
		}
	}
	return false
}

func preventedBackAtTheSource(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || item == nil || item.Trigger == nil {
		return nil
	}
	if !preventedFromACreature(item) {
		return preventedDamageSourceController(g, item, p)
	}
	// "That creature": the source, while it is still on the battlefield.
	src := item.Trigger.Event.Source
	if !onBattlefield(g, src) {
		return nil
	}
	return g.DealDamageToCreatureForEffect(item.SourceCardID, src, p.Amount)
}

func preventedCommandersStrikeTrigger(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || !preventedFromACreature(item) {
		return nil
	}
	ctx := NewContext(g, item)
	t := ReflexiveTrigger{
		Label: shieldSourceName(ctx) + " — each commander creature you control deals damage equal to its power to that creature",
		Body:  commandersStrikeThatCreatureBody,
		Cards: []uuid.UUID{item.Trigger.Event.Source},
	}
	return t.Apply(ctx)
}

// commandersStrikeThatCreature is the trigger's resolution: each
// commander creature its controller controls now deals damage equal to
// its power now to the creature, all at once (one damage instance). A
// creature that has left the battlefield is not there to be dealt
// damage.
func commandersStrikeThatCreature(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	ctx := NewContext(g, item)
	targets := ctx.PayloadCards()
	if len(targets) == 0 {
		return nil
	}
	victim := targets[0]
	if !onBattlefield(g, victim) {
		return nil
	}
	type hit struct {
		source uuid.UUID
		power  int
	}
	var hits []hit
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCommander && c.IsCreature() && c.Controller == item.Controller {
			if pw := c.CurrentPower(); pw > 0 {
				hits = append(hits, hit{c.InstanceID, pw})
			}
		}
	}
	return g.DamageInstanceForEffect(func() error {
		for _, h := range hits {
			if err := g.DealDamageToCreatureForEffect(h.source, victim, h.power); err != nil {
				return err
			}
		}
		return nil
	})
}
