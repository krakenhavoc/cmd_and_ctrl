package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elixir — Artifact {1} (Final Fantasy, #4462):
//
//	"This artifact enters tapped.
//	 {5}, {T}, Exile this artifact: Shuffle all nonland cards from your
//	 graveyard into your library. You gain life equal to the number of
//	 cards shuffled into your library this way."
//
// The enters-tapped clause is the shared CR 614.1c replacement. The
// ability is Feldon's Cane's cost and shuffle with a {5} added: the
// Elixir is exiled as a cost, so it is not in the graveyard when the
// ability resolves. Only nonland cards move; lands stay. The life is
// the number of cards that reached the library, counted as the shuffle
// happens (CR 608.2h), so an empty graveyard still shuffles and gains
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e978db3-77ef-4b93-8f8f-e414e27254bd",
		Name:         "Elixir",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}, Exile this artifact: Shuffle all nonland cards from your graveyard into your library. You gain life equal to the number of cards shuffled into your library this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}"), TapCost(), ExileThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return elixirShuffleAndGain(NewContext(g, item), item.Controller)
			},
		}},
	})
}

// elixirShuffleAndGain moves every nonland card in the controller's
// graveyard to their library, shuffles it, and gains them life equal to
// how many arrived. The graveyard is snapshotted into an ID slice before
// anything moves, as finaleShuffleGraveyardIntoLibrary does.
func elixirShuffleAndGain(ctx *Context, controller uuid.UUID) error {
	p := ctx.Game.PlayerByIDForEffect(controller)
	if p == nil || p.Graveyard == nil || p.Library == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if !c.IsLand() {
			ids = append(ids, c.InstanceID)
		}
	}
	before := p.Library.Size()
	for _, id := range ids {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
			return err
		}
	}
	shuffled := p.Library.Size() - before
	if err := (ShuffleLibrary{Player: controller}).Apply(ctx); err != nil {
		return err
	}
	if shuffled <= 0 {
		return nil
	}
	return GainLife{Player: controller, Amount: shuffled}.Apply(ctx)
}
