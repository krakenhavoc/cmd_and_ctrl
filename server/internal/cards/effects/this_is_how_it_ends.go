package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// This Is How It Ends — Instant {3}{B}:
//
//	"Target creature's owner shuffles it into their library, then
//	 faces a villainous choice — They lose 5 life, or they shuffle
//	 another creature they own into their library."
//
// A villainous choice (CR 701.55, #3140), faced by the creature's OWNER
// (not its controller, not the caster) once the shuffle has finished.
// The shuffle is the ordinary tuck-then-shuffle (Chaos Warp's): a
// commander's owner is offered the command zone first (CR 903.9), and
// the choice waits for that answer, so it is the tuck's continuation
// rather than the next line. Whether the commander went to the command
// zone or into the library, "then" is not an "if you do": the owner
// faces the choice either way.
//
// "Another creature they own" is a creature on the battlefield they own
// (a creature card in a graveyard or hand is not a creature), whoever
// controls it, other than the one just shuffled in. With none the owner
// may still pick it and shuffle nothing (CR 701.55b), so it is priced as
// free. The pick is the ordinary choose-cards prompt; the creature is
// re-read from the battlefield when it is answered (the prompt's Zone).
//
// No simplifications.
var thisIsHowItEndsChoice = NewVillainousChoice("this-is-how-it-ends",
	"This Is How It Ends — lose 5 life, or shuffle another creature you own into your library?",
	VillainousOption{
		Label:    "Lose 5 life",
		LifeCost: 5,
		Body: func(ctx *Context, f Facing) error {
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), f.Player, -5)
		},
	},
	VillainousOption{
		Label:    "Shuffle another creature you own into your library",
		Harm:     2,
		Possible: func(g *game.Game, facing uuid.UUID) bool { return len(thisIsHowItEndsOwnedCreatures(g, facing)) > 0 },
		Body: func(ctx *Context, f Facing) error {
			owned := thisIsHowItEndsOwnedCreatures(ctx.Game, f.Player)
			if len(owned) == 0 {
				return nil
			}
			owner := f.Player
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    owner,
				FromPlayer: owner,
				Source:     ctx.Source(),
				Question:   "This Is How It Ends — shuffle a creature you own into your library",
				Cards:      owned,
				Min:        1,
				Max:        1,
				LiveZone:   game.ZoneBattlefield,
				ThenKey:    thisIsHowItEndsShuffled,
			})
			return nil
		},
	})

// thisIsHowItEndsShuffled is the creature shuffled into its owner's library,
// keyed (#4253): the chooser owns it and is the player who shuffles. An
// on-disk key: never renamed, never reused.
var thisIsHowItEndsShuffled = game.RegisterChooseCardsThen("choose-cards/this-is-how-it-ends-shuffle", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{r.Picked[0]}, game.TuckOptions{}, game.KeyedTail{Ref: thenShuffle, Player: r.Chooser})
})

// thisIsHowItEndsOwnedCreatures lists the creatures on the battlefield
// that `owner` owns.
func thisIsHowItEndsOwnedCreatures(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Owner == owner && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func init() {
	Register(Spec{
		OracleID:     "53467cf4-5db2-4b9c-af2b-03b044c558cd",
		Name:         "This Is How It Ends",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{},
				withCarry(ctx.ItemTail(thisIsHowItEndsTucked), c.Owner))
		},
	})
}

// thisIsHowItEndsTucked is "its owner shuffles, then faces a villainous
// choice", after the tuck, keyed (#4253 slice 10a): the owner is Carry[0].
var thisIsHowItEndsTucked = CardTail("then/this-is-how-it-ends-shuffle-choice", func(ctx *Context, o game.TailOutcome) error {
	owner := tailCarry0(o)
	if owner == uuid.Nil {
		return nil
	}
	if err := ctx.Game.ShuffleLibraryForEffect(owner); err != nil {
		return err
	}
	return FaceVillainousChoice{Players: []uuid.UUID{owner}, Choice: thisIsHowItEndsChoice}.Apply(ctx)
})
