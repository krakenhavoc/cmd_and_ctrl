package cardsbu

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burakos, Party Leader — Legendary Creature — Orc {3}{B}, 2/4 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Burakos is also a Cleric, Rogue, Warrior, and Wizard.
//	 Whenever Burakos attacks, defending player loses X life and you create X
//	 Treasure tokens, where X is the number of creatures in your party.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// "Is also" ADDS the four party types in layer 4 (CR 205.1b), so Burakos stays
// an Orc and is a party member for each of them (CR 700.8). X is read as the
// trigger resolves, from the party the layers show then (CR 700.8a-b: one
// creature per type, assigned to maximise the party), Burakos included. The
// defending player is the player he attacks (the controller of the planeswalker
// or battle, when that is what he attacks). "Choose a Background" is a deck
// construction rule with nothing to resolve in a game, as on Wilson, Refined
// Grizzly.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "883ad210-917b-40e6-8682-31d6eb3ed12a",
		Name:         "Burakos, Party Leader",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{{
			Label: "Burakos is also a Cleric, Rogue, Warrior, and Wizard.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, t := range effects.FinB9PartyTypes {
					if !slices.Contains(c.Subtypes, t) {
						c.Subtypes = append(c.Subtypes, t)
					}
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Burakos — defending player loses X life and you create X Treasure tokens, X being your party size", burakosAttack),
		},
	})
}

func burakosAttack(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	x := effects.FinB9PartySize(g, item.Controller)
	if x <= 0 {
		return nil
	}
	if defender := effects.B17DefendingPlayer(g, item.Trigger.Event); defender != uuid.Nil {
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -x); err != nil {
			return err
		}
	}
	return effects.CreateToken{Template: effects.TreasureToken(), N: x}.Apply(effects.NewContext(g, item))
}
