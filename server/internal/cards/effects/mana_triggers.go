package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_triggers.go — the constructors for Spec.ManaTriggers, the
// CR 605.1b TRIGGERED MANA abilities (#763, ADR 0074).
//
// A triggered mana ability fires when a permanent is TAPPED FOR MANA,
// adds mana, and therefore does not use the stack at all (CR 605.4a).
// It resolves the instant the mana ability that triggered it has
// finished, with no priority window for anybody — which is the whole
// point: Wild Growth's extra {G} has to be in the pool before the
// spell it is paying for is cast.
//
// Write one of these, never a Spec.Triggered entry. An "add mana"
// trigger that fires on a CAST or an ATTACK (Electro, Fire Nation
// Palace) is an ordinary stack trigger, CR 605.5a, and stays there.
//
// Every callback here runs under the write lock the mana production
// path holds. READ-ONLY: *ForEffect accessors only.

// WheneverAttachedTapsForMana is the Aura family's condition:
// "Whenever enchanted land is tapped for mana, its controller adds
// …" — Wild Growth, Overgrowth, Fertile Ground, Utopia Sprawl.
//
// The host predicate is the attachment itself, re-read on every
// production, so moving the Aura moves the trigger with no
// bookkeeping. The mana goes to the player whose pool the land's mana
// landed in, which is "its controller" for every printed card in the
// family.
//
// `produced` is in the ParseProducedMana grammar — "{G}", "{G}{G}",
// "{W|U|B|R|G}" for "one mana of any color".
func WheneverAttachedTapsForMana(label, produced string) game.ManaTrigger {
	return WheneverAttachedTapsForManaFunc(label, func(game.ManaProduced, *game.Card, *game.Game) string {
		return produced
	})
}

// WheneverAttachedTapsForManaFunc is WheneverAttachedTapsForMana for a
// clause whose output is computed — Utopia Sprawl's chosen colour,
// "one mana of any type that land produced" on an Aura.
func WheneverAttachedTapsForManaFunc(
	label string,
	produced func(prod game.ManaProduced, source *game.Card, g *game.Game) string,
) game.ManaTrigger {
	return game.ManaTrigger{
		Label: label,
		AppliesTo: func(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
			return source.IsAttachedTo(prod.SourceID())
		},
		Produced: produced,
	}
}

// WheneverAPlayerTapsALandForMana is the symmetric global: "Whenever a
// player taps a land for mana, that player adds …" — Mana Flare,
// Heartbeat of Spring, Dictate of Karametra. It fires for EVERY
// player's land, opponents included, and the mana goes to whoever
// tapped it.
func WheneverAPlayerTapsALandForMana(
	label string,
	produced func(prod game.ManaProduced, source *game.Card, g *game.Game) string,
) game.ManaTrigger {
	return game.ManaTrigger{
		Label: label,
		AppliesTo: func(prod game.ManaProduced, _ *game.Card, _ *game.Game) bool {
			return prod.Source.IsLand()
		},
		Produced: produced,
	}
}

// WheneverYouTapALandForMana is the one-sided version: "Whenever you
// tap a land for mana, add …" — Mirari's Wake, Zendikar Resurgent. The
// land has to be one the trigger's controller controls, which is what
// "you tap" means.
func WheneverYouTapALandForMana(
	label string,
	produced func(prod game.ManaProduced, source *game.Card, g *game.Game) string,
) game.ManaTrigger {
	return game.ManaTrigger{
		Label: label,
		AppliesTo: func(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
			return prod.Source.IsLand() && prod.Controller == source.Controller
		},
		Produced: produced,
	}
}

// AddsFixedMana is the produced callback for a clause that names its
// mana outright — "{G}", "{G}{G}", "{B}". The overwhelming majority.
func AddsFixedMana(produced string) func(game.ManaProduced, *game.Card, *game.Game) string {
	return func(game.ManaProduced, *game.Card, *game.Game) string { return produced }
}

// AddsOneManaOfAnyTypeProduced is "one mana of any type that land
// produced" (Mana Flare, Mirari's Wake, Zendikar Resurgent, Heartbeat
// of Spring). One produced type is no choice at all and drops straight
// into the pool; a land that produced two is an ordinary mana pick —
// or, inside the auto-tapper, a greedy pick against what the cast
// still owes.
func AddsOneManaOfAnyTypeProduced() func(game.ManaProduced, *game.Card, *game.Game) string {
	return func(prod game.ManaProduced, _ *game.Card, _ *game.Game) string {
		return prod.ProducedColorPipe()
	}
}

// AddsOneManaOfTheChosenColor is "one mana of the chosen color" on a
// permanent that stored one with ChooseColorAsEnters (Utopia Sprawl).
// Before the colour is chosen it adds nothing — the same weaker
// reading ProducedChosenColor takes for a mana ability, and never "any
// colour".
func AddsOneManaOfTheChosenColor() func(game.ManaProduced, *game.Card, *game.Game) string {
	return func(_ game.ManaProduced, source *game.Card, g *game.Game) string {
		if c := g.ChosenColorOf(source.InstanceID); c != "" {
			return "{" + c + "}"
		}
		return ""
	}
}
