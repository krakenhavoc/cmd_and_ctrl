package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carrion Cruiser — Artifact — Vehicle {2}{B}, 3/2 (Aetherdrift, #4464):
//
//	"When this Vehicle enters, mill two cards. Then return a creature or
//	 Vehicle card from your graveyard to your hand. (To mill two cards,
//	 put the top two cards of your library into your graveyard.)
//	 Crew 1 (Tap any number of creatures you control with total power 1
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// Overlord of the Balemurk's shape: the mill finishes first (its
// continuation), and the pick ranges over the whole graveyard, the two
// cards just milled included (CR 608.2c). The return is not targeted and
// not optional, so with a creature or Vehicle card there one of them comes
// back: a lone candidate returns without a question, and two or more ask
// for a pick with a floor of one (Ghostly Dancers' return). With none
// there, nothing happens. Crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98a61659-9882-417c-9c16-16ebb26f9037",
		Name:         "Carrion Cruiser",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Carrion Cruiser — mill two cards, then return a creature or Vehicle card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 2, ThenTail: ctx.ItemTail(carrionCruiserMilled)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Carrion Cruiser"),
		}},
	})
}

// carrionCruiserMilled is the trigger's "then": once the mill has
// settled, a creature or Vehicle card from the controller's graveyard
// goes to their hand. An on-disk key: never renamed, never reused.
var carrionCruiserMilled = CardTail("carrion-cruiser/after-mill", func(ctx *Context, _ game.TailOutcome) error {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	match := Or(Creature(), HasSubtype("Vehicle"))
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if match(ctx.Game, ctx.Controller(), c) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 1 {
		return ReturnFromGraveyard{Target: candidates[0], Dest: game.ZoneHand}.Apply(ctx)
	}
	return takeOneIntoYourHand(ctx, candidates,
		"Carrion Cruiser — return a creature or Vehicle card from your graveyard to your hand", 1)
})
