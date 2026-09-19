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
//
// The pay-2-life decision used to be taken when the trigger fired
// (the CR 603.5 "you may" prompt the harvester asks before the
// ability goes on the stack), because that was the only yes/no the
// engine had. #796 gave a resolving effect its own, so the question
// is now asked where the card prints it — as the ability RESOLVES,
// after the response window, which is when a player actually knows
// whether the 2 life is affordable.
func init() {
	Register(Spec{
		OracleID:     "1a362e4d-6c02-4b67-ab63-c6622e505195",
		Name:         "Crossway Troublemakers",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Deathtouch and lifelink are granted as each of your Vampires is declared as an attacker and last until end of turn, so a Vampire that enters the battlefield already attacking doesn't get them.",
		},
		Triggered: []game.TriggeredAbility{
			b34AttackingVampiresHaveDeathtouchAndLifelink("Crossway Troublemakers"),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b34VampireYouControlDied(ev, source, g)
			}, "Crossway Troublemakers — pay 2 life, draw a card", crosswayMayPayTwoLifeToDraw),
		},
	})
}

// crosswayMayPayTwoLifeToDraw is the printed body: "you may pay 2
// life. If you do, draw a card."
//
// LifeCost is declared so the move list prices it (#547). Without it
// a bot at 2 life answers "pay" and dies — the branch re-checks
// affordability anyway (life moves between the question and the
// answer), but a policy that cannot see the price never gets that far.
//
// Caller holds g.mu.
func crosswayMayPayTwoLifeToDraw(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Crossway Troublemakers — a Vampire died. Pay 2 life to draw a card?",
		YesLabel: "Pay 2 life",
		NoLabel:  "Decline",
		LifeCost: crosswayLifePayment,
		OnYes:    crosswayPayTwoLifeThenDraw,
	}.Apply(NewContext(g, item))
}

// crosswayPayTwoLifeThenDraw is the "if you do" branch: a CR 118.3
// cost, then the linked draw. The body is the shared one in
// helpers.go — Erebos, Bleak-Hearted prints the same sentence.
//
// Caller holds g.mu.
func crosswayPayTwoLifeThenDraw(ctx *Context) error {
	return payLifeThenDraw(ctx, crosswayLifePayment, 1)
}

// crosswayLifePayment is the printed 2.
const crosswayLifePayment = 2
