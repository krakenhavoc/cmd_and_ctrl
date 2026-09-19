package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Command Beacon — Land:
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Put your commander into your hand from
//	 the command zone."
//
// The answer to commander tax, and the batch-01 triage filed it under
// "player / game-rule statics" — which it is not. It is an ordinary
// activated ability with two cost components the engine has carried
// since S21 ({T} and sacrifice-this), and one move: the command zone
// is a zone like any other, findable by the shared zone lookup, so
// the commander goes to hand through the same routed exit every
// bounce uses.
//
// Two cost components and the {T} among them is why the land can only
// do this once, and why a Command Beacon that is already tapped for
// mana cannot also fetch.
//
// A commander moving to a hand is a CR 903.9 destination, so the
// route offers its owner the command zone instead — an offer they
// will decline, since going to hand is the whole point, but it is the
// same window every other commander move opens and this card is not
// special-cased out of it.
//
// DECLARED SIMPLIFICATION, weaker than printed: with two commanders
// (partners, a Background) the card takes the FIRST one in the
// command zone rather than asking which. The printed card lets the
// player choose. One commander — every other deck — is unaffected.
func init() {
	Register(Spec{
		OracleID:     "7e8c2a18-e404-40ff-a9e0-ec3eeb6d576e",
		Name:         "Command Beacon",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"With two commanders in the command zone it takes the first one rather than asking you which.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice this land: Put your commander into your hand from the command zone",
			Cost:   game.AbilityCost{Tap: true, SacrificeSelf: true},
			Effect: commandBeaconToHand,
		}},
	})
}

// commandBeaconToHand moves the activator's commander out of the
// command zone and into their hand. An empty command zone — the
// commander is on the battlefield, or in a graveyard — does nothing,
// which is what the printed card does too.
func commandBeaconToHand(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Command == nil || len(p.Command.Cards) == 0 {
		return nil
	}
	return g.BounceToHandForEffect(p.Command.Cards[0].InstanceID)
}
