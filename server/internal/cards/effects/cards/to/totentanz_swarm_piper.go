package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Totentanz, Swarm Piper — Legendary Creature — Human Warlock Bard
// {1}{B}{R}, 2/3 (Wilds of Eldraine, #4471):
//
//	"Whenever Totentanz or another nontoken creature you control dies,
//	 create a 1/1 black Rat creature token with "This token can't block."
//	 {1}{B}: Target attacking Rat you control gains deathtouch until end
//	 of turn."
//
// The death trigger is Avacyn, Angel of Horror's: Totentanz itself, or
// another nontoken creature that was yours as it died (CR 603.10a). A Rat
// token dying makes nothing, so the Rats don't feed themselves. The Rat is
// Twisted Sewer-Witch's. The activated row is Alabaster Mage's grant on a
// target that is attacking, a Rat ("Rat" is a word of the card's text,
// read through it by the targeting walk) and yours; a Rat that has left
// combat in response is no longer a legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfa46f8f-b3ec-4ee0-acdd-622f3177e02a",
		Name:         "Totentanz, Swarm Piper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.ThisDied(ev, source, ch, g) || effects.B19AnotherNontokenCreatureYouControlDied(ev, source, g)
			}, "Totentanz, Swarm Piper — create a 1/1 black Rat creature token with \"This token can't block.\"",
				effects.Do(effects.CreateToken{Template: effects.CantBlockRatToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Target attacking Rat you control gains deathtouch until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")),
			Cost:    effects.ManaCost("{1}{B}"),
			Targets: effects.TargetCreature("target attacking Rat you control", effects.IsAttackingCreature(), effects.HasSubtype("Rat"), effects.YouControl()),
			Effect:  effects.JmpB7FirstLegalTargetGains("Totentanz, Swarm Piper — deathtouch until end of turn", "deathtouch"),
		}},
	})
}
