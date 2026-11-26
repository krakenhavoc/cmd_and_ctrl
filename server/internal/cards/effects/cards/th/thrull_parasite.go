package cardsth

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thrull Parasite — Creature — Thrull {B}, 1/1 (Gatecrash, S82 OLD b408,
// #5565):
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each
//	 opponent loses 1 life and you gain that much life.)
//	 {T}, Pay 2 life: Remove a counter from target nonland permanent."
//
// Extort is the keyword (Pontiff of Blight). The activation pays 2 life as a
// cost (CR 118.4) and taps the Thrull. Removing "a counter" is the controller's
// choice of kind: a permanent with one kind loses one of it, one with several
// kinds asks which, through a registered option pick that carries the target.
// A target that is gone or has no counters leaves nothing to remove.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "577eb51d-7979-4f97-a8b1-e31242720611",
		Name:         "Thrull Parasite",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Extort("Thrull Parasite")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay 2 life: Remove a counter from target nonland permanent.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.TapCost(), effects.PayLife(2)),
			Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, found := g.LookupCardForEffect(id)
				if !found {
					return nil
				}
				var kinds []string
				for kind, n := range c.Counters {
					if n > 0 {
						kinds = append(kinds, kind)
					}
				}
				sort.Strings(kinds)
				switch len(kinds) {
				case 0:
					return nil
				case 1:
					return effects.AddCounter{Target: id, Kind: kinds[0], N: -1}.Apply(ctx)
				}
				options := make([]game.ChoiceOption, 0, len(kinds))
				for _, kind := range kinds {
					options = append(options, game.ChoiceOption{Label: kind})
				}
				return effects.PickOption{
					Question: "Thrull Parasite — choose the kind of counter to remove",
					Options:  options,
					ThenKey:  thrullParasiteKind,
					Carry:    []uuid.UUID{id},
				}.Apply(ctx)
			},
		}},
	})
}

// thrullParasiteKind removes one counter of the chosen kind from the target
// carried on the prompt. An on-disk key: never renamed, never reused.
var thrullParasiteKind = effects.OptionPickThen("option-pick/s82-old-b408-thrull-parasite-counter", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	return effects.AddCounter{Target: r.Carry[0], Kind: r.Option.Label, N: -1}.Apply(ctx)
})
