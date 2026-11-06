package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// renew.go — Renew, the Tarkir: Dragonstorm ability word (#4463):
//
//	"Renew — {B}, Exile this card from your graveyard: Put a +1/+1
//	 counter on target creature. Activate only as a sorcery."
//
// An ability word has no rules meaning (CR 207.2c): each renew ability is
// an ordinary activated ability that functions only from its owner's
// graveyard and exiles the card as part of its cost, Scavenge's shape
// (scavenge.go). The cost moves the card to exile before the ability is on
// the stack, so it can be activated once.
//
//	Activated: []ActivatedAbility{Renew("{B}", "Put a +1/+1 counter on target creature.",
//	    TargetCreature("target creature"), renewCountersOnTarget(game.CounterPlusOne, 1))},
//
// Append-only: a new renew card calls Renew.

// Renew is "Renew — <cost>, Exile this card from your graveyard: <text>
// Activate only as a sorcery." `text` is the printed effect sentence,
// `targets` its clause (nil for none) and `effect` its body.
func Renew(cost, text string, targets *game.TargetSpec, effect func(g *game.Game, item *game.StackItem) error) ActivatedAbility {
	return ActivatedAbility{
		Label:        "Renew — " + cost + ", Exile this card from your graveyard: " + text + " Activate only as a sorcery.",
		Cost:         Plus(ManaCost(cost), ExileThis()),
		Zones:        []game.ZoneKind{game.ZoneGraveyard},
		SorcerySpeed: true,
		Targets:      targets,
		Effect:       effect,
	}
}

// renewCountersOnTarget is "put N <kind> counters on target creature": the
// renew body of Adorned Crocodile, Agent of Kotis and Alchemist's
// Assistant. A target that has left by resolution is illegal and the
// ability does nothing (CR 608.2b).
func renewCountersOnTarget(kind string, n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		return AddCounter{Target: item.Targets[0].ID, Kind: kind, N: n}.Apply(NewContext(g, item))
	}
}
