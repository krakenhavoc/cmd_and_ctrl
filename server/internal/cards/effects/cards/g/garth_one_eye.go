package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garth One-Eye — Legendary Creature — Human Wizard {W}{U}{B}{R}{G}, 5/5:
//
//	"{T}: Choose a card name that hasn't been chosen from among Disenchant,
//	 Braingeyser, Terror, Shivan Dragon, Regrowth, and Black Lotus. Create a
//	 copy of the card with the chosen name. You may cast the copy. (You
//	 still pay its costs.)"
//
// The names already chosen are noted on Garth (Card.Notes, #6097), so
// "hasn't been chosen" spans its activations and turns, and a Garth that
// leaves and returns is a new object with none chosen (CR 400.7). With all
// six chosen there is nothing to choose and nothing happens.
//
// The copy is CR 707.13's: built from the named card's Oracle reference
// (its printed record), created outside the game, and cast from there
// while the ability resolves (CR 707.12, 608.2g), at its own costs
// (game.CopyNamedCardAndOfferCastForEffect, #6155). Casting it is optional;
// a copy that is not cast ceases to exist (CR 707.10a). A permanent spell
// copy becomes a token as it resolves (CR 707.10f, 608.3f).
//
// The name is noted as it is chosen, so a Garth that has left the
// battlefield by then (its ability still resolves, CR 113.7a) notes
// nothing, which no later activation of that object can read anyway.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8c4312e-03ae-43f3-b92c-4adb6875534c",
		Name:         "Garth One-Eye",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Choose a card name that hasn't been chosen from among Disenchant, Braingeyser, Terror, Shivan Dragon, Regrowth, and Black Lotus. Create a copy of the card with the chosen name. You may cast the copy.",
			Cost:  effects.TapCost(),
			// The union over the six copies (ADR 0142): Disenchant and
			// Terror remove, a Shivan Dragon copy is a creature made at
			// instant speed.
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				left := garthUnchosen(g, ctx)
				if len(left) == 0 {
					return nil
				}
				options := make([]game.ChoiceOption, len(left))
				for i, n := range left {
					options[i] = game.ChoiceOption{Label: n}
				}
				return effects.PickOption{
					Question: "Garth One-Eye — choose a card name that hasn't been chosen",
					Options:  options,
					ThenKey:  garthChosen,
					WithItem: true,
					Words:    left,
				}.Apply(ctx)
			},
		}},
	})
}

// garthNames is the fixed list, in printed order.
var garthNames = effects.CheckNamedCardCopies("Garth One-Eye",
	[]string{"Disenchant", "Braingeyser", "Terror", "Shivan Dragon", "Regrowth", "Black Lotus"})

// garthUnchosen is the names not yet noted for this Garth, in printed
// order.
func garthUnchosen(g *game.Game, ctx *effects.Context) []string {
	notes := g.NotesForEffect(ctx.Source())
	var out []string
	for _, n := range garthNames {
		if !game.NotedIn(notes, n) {
			out = append(out, n)
		}
	}
	return out
}

// garthChosen notes the chosen name and makes and offers the copy. An
// on-disk key.
var garthChosen = effects.OptionPickThen("option-pick/garth-one-eye-name", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	name := r.Option.Label
	if r.Index >= 0 && r.Index < len(r.Words) {
		name = r.Words[r.Index]
	}
	if ref, ok := ctx.SourceRef(); ok {
		if info, ok := ctx.Game.PermanentForEffect(ref); ok && !info.Left {
			ctx.Game.AddNoteForEffect(ref.ID, name)
		}
	}
	_, err := effects.CopyNamedCardAndOfferCast(ctx, "Garth One-Eye", name)
	return err
})
