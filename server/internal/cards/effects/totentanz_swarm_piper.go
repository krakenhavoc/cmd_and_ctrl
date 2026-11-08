package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cfa46f8f-b3ec-4ee0-acdd-622f3177e02a",
		Name:         "Totentanz, Swarm Piper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ThisDied(ev, source, ch, g) || b19AnotherNontokenCreatureYouControlDied(ev, source, g)
			}, "Totentanz, Swarm Piper — create a 1/1 black Rat creature token with \"This token can't block.\"",
				Do(CreateToken{Template: CantBlockRatToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Target attacking Rat you control gains deathtouch until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch")),
			Cost:    ManaCost("{1}{B}"),
			Targets: TargetCreature("target attacking Rat you control", attackingCreature(), HasSubtype("Rat"), YouControl()),
			Effect:  jmpB7FirstLegalTargetGains("Totentanz, Swarm Piper — deathtouch until end of turn", "deathtouch"),
		}},
	})
}
