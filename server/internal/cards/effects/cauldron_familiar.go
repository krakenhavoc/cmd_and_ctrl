package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cauldron Familiar — Creature — Cat {B}, 1/1 (EDHREC rank 2974):
//
//	"When this creature enters, each opponent loses 1 life and you
//	 gain 1 life.
//	 Sacrifice a Food: Return this card from your graveyard to the
//	 battlefield."
//
// The Cat half of Cat-Oven. The enters trigger drains every opponent
// for 1 and gains the controller 1 — one life, not one per opponent,
// as printed (eachOpponentLosesLife then GainLife).
//
// The second ability is a CR 602 activation from the GRAVEYARD
// (CR 113.6), reaching it through ActivatedAbility.Zones exactly as
// Reassembling Skeleton's does — a sacrifice cost rather than a mana
// one, but the same activation path, the same cost validation, the
// same stack item (#1381). The Cat returns to the battlefield as a
// NEW object, under its owner's control (no "under your control"
// clause is printed), and its ETB trigger above fires again — which
// is the whole Cat-Oven loop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7dc2e62-1c50-4ed7-b71f-2d782a447a5e",
		Name:         "Cauldron Familiar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cauldron Familiar — each opponent loses 1 life and you gain 1 life", drainEachOpponent),
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice a Food: Return this card from your graveyard to the battlefield.",
			Cost:  game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))},
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnThisFromGraveyardToBattlefield(g, item)
			},
		}},
	})
}

// returnThisFromGraveyardToBattlefield is returnThisFromGraveyardTapped's
// untapped sibling (reassembling_skeleton.go) — "<cost>: Return this
// card from your graveyard to the battlefield," no "tapped" clause.
// Cauldron Familiar is the first card to need it (#1381); package-level
// for the same undo-safety reason every other Effect closure is.
func returnThisFromGraveyardToBattlefield(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	// CR 602.5 / 608.2a: the card may have left the graveyard between
	// activation and resolution.
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return (ReturnFromGraveyard{
		Target:     id,
		Dest:       game.ZoneBattlefield,
		Controller: item.Controller,
	}).Apply(NewContext(g, item))
}
