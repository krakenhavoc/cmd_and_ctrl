package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scavenge.go — #1221: scavenge (CR 702.96).
//
// "Scavenge {5}{G}" is an ACTIVATED ABILITY that functions only from
// its owner's GRAVEYARD, and whose cost exiles the card that has it
// (CR 702.96a):
//
//	"{5}{G}, Exile this card from your graveyard: Put a number of
//	 +1/+1 counters equal to this card's power on target creature.
//	 Activate only as a sorcery."
//
// It is the first card in the catalog to pay AbilityCost.ExileSelf,
// and the reason that component exists (exile_cost.go). The order it
// forces is worth stating once: the COST moves the card to exile
// before the ability is even on the stack, so "this card's power" is
// read at RESOLUTION off a card sitting in exile, by instance ID.
// That is the same number the graveyard held — printed power on a
// card outside the battlefield, which no layer touches (CR 613 runs
// on permanents) — so reading it late is reading it right, and no
// snapshot on the stack item is needed to get there.

// Scavenge is "Scavenge <cost>" — CR 702.96a. `cost` is the printed
// mana component, "{5}{G}"; the exile, the graveyard zone, the target
// clause and the sorcery-speed restriction come with the keyword.
//
//	Activated: []ActivatedAbility{Scavenge("{5}{G}")},
func Scavenge(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label: "Scavenge " + cost + " (" + cost + ", Exile this card from your graveyard: " +
			"Put a number of +1/+1 counters equal to this card's power on target creature. " +
			"Activate only as a sorcery.)",
		Cost:         Plus(ManaCost(cost), ExileThis()),
		Zones:        []game.ZoneKind{game.ZoneGraveyard},
		SorcerySpeed: true,
		Targets:      TargetCreature("target creature"),
		Effect:       scavengeCounters,
	}
}

// scavengeCounters is the ability body: read the exiled card's power,
// put that many +1/+1 counters on the target.
//
// Package-level and capture-free, like every other ability body that
// has to survive an undo's clone.
func scavengeCounters(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	// The source is in EXILE by now — its own activation cost put it
	// there — and LookupCardForEffect finds a card in whatever zone
	// holds it. A card that has since moved again (an effect that
	// shuffles exile into a library) answers false and the ability
	// does nothing, CR 608.2a.
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	// CR 107.1b: a negative power puts no counters on, and neither
	// does a zero. Not an error — "a number of counters equal to 0"
	// is a legal instruction that does nothing.
	n := src.Power
	if n <= 0 {
		return nil
	}
	return AddCounter{
		Target: item.Targets[0].ID,
		Kind:   game.CounterPlusOne,
		N:      n,
	}.Apply(NewContext(g, item))
}
