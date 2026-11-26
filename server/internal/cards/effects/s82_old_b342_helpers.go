package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b342_helpers.go — S82 OLD batch 342 (#5553).

// SacrificeAPegasusOrThisAnswer is Sacred Mesa's upkeep tax answered,
// keyed: index 0 asks the controller for a Pegasus to sacrifice, any other
// index sacrifices the Mesa. The predicate reads the Pegasus subtype
// through the card, so a changeling counts.
var SacrificeAPegasusOrThisAnswer = sacrificeOneOrThisAnswer("option-pick/sacrifice-this-unless-you-sacrifice-a-pegasus", "a Pegasus", Subtype("Pegasus"))

// DivideGreatestDamageDealtThisTurn is Impact Resonance's "X damage divided
// as you choose among any number of target creatures, where X is the
// greatest amount of damage dealt by a source to a permanent or player this
// turn": the largest single damage event of the turn log, read once as the
// spell is cast (CR 601.2d), so damage dealt in response does not change
// what is divided. Damage that was fully prevented dealt none and is not an
// event with an amount.
var DivideGreatestDamageDealtThisTurn = game.RegisterDivideAmount("greatest-damage-dealt-this-turn",
	func(g *game.Game, _ game.DivideAmountArgs) game.DivideSpec {
		most := 0
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind == game.EventDealDamage && ev.Amount > most {
				most = ev.Amount
			}
		}
		return game.DivideSpec{Total: most}
	})
