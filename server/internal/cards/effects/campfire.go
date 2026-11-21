package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Campfire — Artifact {1} (Commander Masters):
//
//	"{1}, {T}: You gain 2 life.
//	 {2}, {T}, Exile this artifact: Put all commanders you own from the
//	 command zone and from your graveyard into your hand. Then shuffle
//	 your graveyard into your library."
//
// The second ability puts every commander its controller owns that sits in
// their command zone or graveyard into their hand (nothing is chosen), by the
// same route Command Beacon uses (BounceToHandForEffect), so a CR 903.9
// choice to keep one in the command zone is still the owner's. Each is judged
// as it resolves. Then, once those moves have settled, the graveyard is
// shuffled into the library.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "011ff7a5-6bce-4784-a19b-e74f7b49b8f1",
		Name:         "Campfire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: You gain 2 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Effect:  Do(GainLife{Amount: 2}),
			},
			{
				Label:   "{2}, {T}, Exile this artifact: Put all commanders you own from the command zone and from your graveyard into your hand. Then shuffle your graveyard into your library.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), TapCost(), ExileThis()),
				Effect:  Do(Instruction(campfireRecoverCommanders)),
			},
		},
	})
}

func campfireRecoverCommanders(ctx *Context) error {
	g := ctx.Game
	me := ctx.Controller()
	p := g.PlayerByIDForEffect(me)
	if p == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, z := range []*game.Zone{p.Command, p.Graveyard} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if c.IsCommander && c.Owner == me && !c.IsToken() {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return finaleShuffleGraveyardIntoLibrary(ctx, me)
	}
	return g.BounceCardsToHandKeyedForEffect(ids, ctx.Tail(campfireShuffled))
}

// campfireShuffled is "Then shuffle your graveyard into your library", run
// once every commander's move has settled (a CR 903.9 offer to keep one in
// the command zone is answered first), so a commander that is still in the
// graveyard is shuffled with it. An on-disk key: never renamed, never reused.
var campfireShuffled = CardTail("campfire/shuffle-graveyard", func(ctx *Context, _ game.TailOutcome) error {
	return finaleShuffleGraveyardIntoLibrary(ctx, ctx.Controller())
})
