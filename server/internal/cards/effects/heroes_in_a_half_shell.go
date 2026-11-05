package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heroes in a Half Shell — Legendary Creature — Mutant Ninja Turtle
// {W}{U}{B}{R}{G}, 5/5 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Vigilance, menace, trample, haste
//	 Whenever one or more Mutants, Ninjas, and/or Turtles you control
//	 deal combat damage to a player, put a +1/+1 counter on each of those
//	 creatures and draw a card."
//
// The trigger is Vulture, Feathered Fiend's shape
// (tmcB1bThoseCreaturesDealtCombatDamage) with the creature filter "a
// Mutant, Ninja or Turtle you control", read off the damage source's
// effective subtypes through this card's own words (CR 612.2), so a
// changeling counts. Per the ruling it triggers once for each player
// damaged at the same time, and each instance puts counters only on the
// Mutants, Ninjas and Turtles that dealt combat damage to that player.
// The four keywords are PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9f2f12cc-0adb-482d-bbce-05e4ce8850c4",
		Name:            "Heroes in a Half Shell",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			tmcB1bThoseCreaturesDealtCombatDamage(
				"Heroes in a Half Shell — put a +1/+1 counter on each of those creatures and draw a card",
				heroesInAHalfShellHit),
		},
	})
}

// heroesInAHalfShellHit is "a Mutant, Ninja, or Turtle you control dealt
// combat damage to a player", read as the damage is dealt.
func heroesInAHalfShellHit(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	if !ok {
		return false
	}
	for _, sub := range source.TextSubtypes("Mutant", "Ninja", "Turtle") {
		if c.HasSubtype(sub) {
			return true
		}
	}
	return false
}
