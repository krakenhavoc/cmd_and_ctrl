package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// monstrosity.go — the card-facing half of CR 701.37, monstrosity
// (ADR 0071 amendment 2026-09-28, #1700). The engine half is
// game.MonstrosityForEffect and the DesignationMonstrous gate.
//
// Every monstrosity card is some subset of three printed lines, and
// each has one constructor here:
//
//	Activated: []ActivatedAbility{Monstrosity(ManaCost("{5}{R}{R}"), 3)},
//	Triggered: []game.TriggeredAbility{WhenBecomesMonstrous("…", effect)},
//	Static:    []game.StaticAbility{MonstrousKeywords("hexproof", "indestructible")},
//
// "Monstrosity X" is MonstrosityX with a cost carrying {X}; a
// becomes-monstrous trigger that says "X" reads MonstrosityXOf(item).

// Monstrous is the CR 701.37b gate: this ability exists while the
// permanent is monstrous — "As long as this creature is monstrous,
// …". Once monstrous a permanent stays monstrous until it leaves the
// battlefield, so nothing that reads this has to worry about it going
// back off.
func Monstrous() game.Designation { return game.Monstrous() }

// Monstrosity is the "[cost]: Monstrosity N." activated ability
// (CR 701.37a) for a fixed N.
//
// No Condition and no ActiveWhen, for Harness's reason: "If this
// creature isn't monstrous, …" is an IF inside the effect, not an
// activation restriction. The ability can be activated again once the
// creature is monstrous — and, more to the point, activated a second
// time in response to the first — and the second resolution does
// nothing. game.MonstrosityForEffect makes that check as the ability
// resolves (CR 701.37b's "stays monstrous" plus CR 608.2), which a
// Condition checked at activation would get wrong in exactly the
// in-response case.
//
// Not sorcery speed: monstrosity is activated any time its controller
// has priority, which is how a Stormbreath Dragon grows at the end of
// an opponent's turn.
//
// The label is the printed line without its reminder text,
// "{5}{R}{R}: Monstrosity 3.", which is what
// TestAbilitiesMatchOracleText compares.
func Monstrosity(cost game.AbilityCost, n int) ActivatedAbility {
	return monstrosityAbility(cost, strconv.Itoa(n), func(*game.StackItem) int { return n })
}

// MonstrosityX is "[cost]: Monstrosity X." — N is the X announced when
// the ability was activated (CR 107.3 / 602.2b), so `cost` must carry
// {X} in its mana (Polukranos's "{X}{X}{G}", Domesticated Hydra's
// "{X}{G}{G}{G}"). X of zero is legal: the creature gets no counters
// and still becomes monstrous.
func MonstrosityX(cost game.AbilityCost) ActivatedAbility {
	return monstrosityAbility(cost, "X", func(item *game.StackItem) int {
		if item.XValue < 0 {
			return 0
		}
		return item.XValue
	})
}

func monstrosityAbility(cost game.AbilityCost, n string, amount func(*game.StackItem) int) ActivatedAbility {
	label := "Monstrosity " + n + "."
	if cost.Mana != "" {
		label = cost.Mana + ": " + label
	}
	return ActivatedAbility{
		Label: label,
		Cost:  cost,
		// ADR 0142: +1/+1 counters, and the creature becomes monstrous.
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Effect: func(g *game.Game, item *game.StackItem) error {
			// #1432: a creature that left and came back is a new
			// object; the monstrosity ability of the old one does
			// not make the new one monstrous.
			if sourceIsNewObject(g, item) {
				return nil
			}
			return g.MonstrosityForEffect(item.SourceCardID, amount(item))
		},
	}
}

// WhenBecomesMonstrous is "When this creature becomes monstrous, …"
// (CR 701.37c). It watches the one event MonstrosityForEffect emits,
// which fires only on the resolution that actually made the creature
// monstrous — never for a second activation, never for a creature
// that was already monstrous.
//
// Set Targets (or TargetsFrom, for a clause sized by X) on the result
// for a targeted trigger, exactly as on any other trigger.
func WhenBecomesMonstrous(label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecameMonstrous},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return source != nil && ev.CardID == source.InstanceID
		},
		Effect: effect,
	}
}

// MonstrosityXOf is the X a becomes-monstrous trigger reads: the N of
// the "Monstrosity N" instruction that made the creature monstrous
// (CR 701.37c), carried on the triggering event. The ANNOUNCED N, not
// the counters that landed — a Doubling Season doubles Polukranos's
// counters and leaves its X alone.
func MonstrosityXOf(item *game.StackItem) int {
	if item == nil {
		return 0
	}
	return item.Trigger.Event.Amount
}

// MonstrousKeywords is "As long as this creature is monstrous, it has
// [keywords]" — a layer-6 self-grant behind the Monstrous gate, the
// monstrosity twin of station's ThresholdKeywords.
func MonstrousKeywords(keywords ...string) game.StaticAbility {
	kws := append([]string(nil), keywords...)
	return game.StaticAbility{
		Layer:      game.Layer6Ability,
		ActiveWhen: Monstrous(),
		AppliesTo:  selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			appendKeywordsTo(c, kws)
		},
	}
}
