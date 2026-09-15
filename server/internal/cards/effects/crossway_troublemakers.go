package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crossway Troublemakers — Creature — Vampire {5}{B}, 5/5 (EDHREC
// rank 3630):
//
//	"Attacking Vampires you control have deathtouch and lifelink.
//	 (Any amount of damage they deal to a creature is enough to
//	 destroy it. Damage dealt by those creatures also causes their
//	 controller to gain that much life.)
//	 Whenever a Vampire you control dies, you may pay 2 life. If you
//	 do, draw a card."
//
// The Vampire tribal six-drop. The printed static is written as an
// attack trigger for the reason Blade Historian gives: the layer
// engine does not recompute on an attack declaration, so a static
// gated on "is attacking" stays cached from before combat. One
// trigger per attacking Vampire the controller controls — the
// Troublemakers themselves included — granting both keywords until
// end of turn through the turn-scoped registry
// (b34AttackingVampiresHaveDeathtouchAndLifelink). The dies trigger
// fires for any Vampire creature the controller controlled, the
// Troublemakers' own death included; its "you may pay 2 life" is
// the ordinary trigger prompt, and a controller who cannot pay when
// it resolves neither pays nor draws.
//
// Two declared simplifications, both weaker than printed:
//
//   - Blade Historian's: the grant lasts the turn rather than the
//     attack, and lands only on Vampires DECLARED as attackers. A
//     Vampire put onto the battlefield attacking is not covered; a
//     Vampire removed from combat keeps the keywords until end of
//     turn, which changes nothing outside combat.
//   - The pay-2-life decision is taken when the trigger fires
//     rather than as it resolves, so opponents see the answer before
//     it resolves.
func init() {
	Register(Spec{
		OracleID:     "1a362e4d-6c02-4b67-ab63-c6622e505195",
		Name:         "Crossway Troublemakers",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Deathtouch and lifelink are granted as each of your Vampires is declared as an attacker and last until end of turn, so a Vampire that enters the battlefield already attacking doesn't get them.",
			"Whether to pay 2 life for the card is decided when the trigger goes on the stack, not as it resolves.",
		},
		Triggered: []game.TriggeredAbility{
			b34AttackingVampiresHaveDeathtouchAndLifelink("Crossway Troublemakers"),
			Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b34VampireYouControlDied(ev, source, g)
			}, "Crossway Troublemakers — pay 2 life, draw a card", b34PayLifeToDraw(2)), "Crossway Troublemakers — a Vampire died. Pay 2 life to draw a card?"),
		},
	})
}
