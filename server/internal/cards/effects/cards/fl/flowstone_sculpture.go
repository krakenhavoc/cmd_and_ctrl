package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flowstone Sculpture — Artifact Creature — Shapeshifter {6}, 4/4
// (Tempest, S82 OLD batch 759, #5628):
//
//	"{2}, Discard a card: Put a +1/+1 counter on this creature or this
//	 creature gains flying, first strike, or trample. (This effect lasts
//	 indefinitely.)"
//
// The choice among the four is made as the ability resolves (Ballroom
// Brawlers' keyword pick, with the counter first as the branch that always
// works). A keyword is granted to the Sculpture itself for no stated
// duration (CR 611.2a), pinned to the object (CR 400.7); the counter and the
// grant happen only if it is still the permanent that activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42b01697-8663-4b53-86d5-699063bd0778",
		Name:         "Flowstone Sculpture",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Discard a card: Put a +1/+1 counter on this creature or this creature gains flying, first strike, or trample.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.DiscardN(1, "a card")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.PickOption{
					Question: "Flowstone Sculpture — put a +1/+1 counter on it, or it gains…",
					Options:  []game.ChoiceOption{{Label: "A +1/+1 counter"}, {Label: "Flying"}, {Label: "First strike"}, {Label: "Trample"}},
					ThenKey:  flowstoneSculptureChoice,
					Carry:    []uuid.UUID{item.SourceCardID},
					WithItem: true,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// flowstoneSculptureChoice is the pick, keyed (an on-disk identity: never
// renamed, never reused).
var flowstoneSculptureChoice = effects.OptionPickThen("option-pick/s82-old-b759-flowstone-sculpture", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 || r.Index > 3 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	id := r.Carry[0]
	if r.Index == 0 {
		return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	keyword := []string{"", "flying", "first strike", "trample"}[r.Index]
	return effects.ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddKeywordsMod(keyword)},
		Duration: game.IndefiniteDuration(),
		Label:    "Flowstone Sculpture — " + keyword,
	}.Apply(ctx)
})
