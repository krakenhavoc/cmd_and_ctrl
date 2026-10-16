package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
// With two commanders (partners, a Background) in the command zone the
// player chooses which one comes back (CR 608.2c); with one there is no
// question to ask.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e8c2a18-e404-40ff-a9e0-ec3eeb6d576e",
		Name:         "Command Beacon",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Put your commander into your hand from the command zone",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true},
			Effect:  commandBeaconToHand,
		}},
	})
}

// commandBeaconToHand moves the activator's commander out of the
// command zone and into their hand. An empty command zone — the
// commander is on the battlefield, or in a graveyard — does nothing,
// which is what the printed card does too.
func commandBeaconToHand(g *game.Game, item *game.StackItem) error {
	return putCommanderIntoHand(g, item, "Command Beacon", false)
}

// putCommanderIntoHand is "Put your commander into your hand from the
// command zone" — Command Beacon, and Road of Return's second bullet.
// With two (partners, a Background) the player chooses which (CR
// 608.2c); with none nothing happens. `name` heads the question.
// `commandersOnly` narrows the candidates to the commanders the
// controller owns (Road of Return); Command Beacon keeps every card in
// its activator's command zone, as it always has.
func putCommanderIntoHand(g *game.Game, item *game.StackItem, name string, commandersOnly bool) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Command == nil {
		return nil
	}
	candidates := make([]uuid.UUID, 0, len(p.Command.Cards))
	for _, c := range p.Command.Cards {
		if commandersOnly && (!c.IsCommander || c.Owner != item.Controller) {
			continue
		}
		candidates = append(candidates, c.InstanceID)
	}
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return g.BounceToHandForEffect(candidates[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   name + " — put which commander into your hand?",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		Zone:       game.ZoneCommand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if err := g.BounceToHandForEffect(id); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
}
