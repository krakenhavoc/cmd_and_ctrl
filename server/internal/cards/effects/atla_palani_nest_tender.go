package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atla Palani, Nest Tender — Legendary Creature — Human Shaman
// {1}{R}{G}{W}, 2/3:
//
//	"{2}, {T}: Create a 0/1 green Egg creature token with defender.
//	 Whenever an Egg you control dies, reveal cards from the top of your
//	 library until you reveal a creature card. Put that card onto the
//	 battlefield and the rest on the bottom of your library in a random
//	 order."
//
// Eggs that hatch into whatever the deck has. The death trigger is the
// shared reveal-until sentence (RevealUntilThenPutOntoBattlefield,
// #745): the run is revealed, the creature enters through the CR 614
// pipeline without being cast, and the rest go to the bottom in an
// order drawn from the game's seeded RNG.
//
// "An Egg you control" reads the Egg as it last existed on the
// battlefield, so a token Egg counts (it is in the graveyard long
// enough to be read) and so does any other Egg creature — a
// changeling included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b56cebe0-3752-4ce5-afbd-911543784015",
		Name:         "Atla Palani, Nest Tender",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Create a 0/1 green Egg creature token with defender.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: Do(CreateToken{
				Template: TokenCard("0/1 green Egg with defender"),
				N:        1,
			}),
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, anEggYouControlDied,
				"Atla Palani, Nest Tender — reveal until a creature card and put it onto the battlefield",
				Do(RevealUntilThenPutOntoBattlefield{
					Match:  game.Card.IsCreature,
					Reason: "Atla Palani — revealed until a creature card",
				})),
		},
	})
}

// anEggYouControlDied is "whenever an Egg you control dies".
func anEggYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller && dead.HasSubtype("Egg")
}
