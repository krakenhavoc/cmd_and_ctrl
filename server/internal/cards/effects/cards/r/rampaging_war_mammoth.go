package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampaging War Mammoth — Creature — Elephant {5}{R}{R}, 9/7 (Tales of
// Middle-earth Commander, S82 GAP2 batch 04, #5948):
//
//	"Trample
//	 Cycling {X}{2}{R} ({X}{2}{R}, Discard this card: Draw a card.)
//	 When you cycle this card, destroy up to X target artifacts."
//
// Webstrike Elite's shape: the cycling ability draws, and a separate
// trigger fired from the graveyard (CR 702.29c) destroys, so countering
// either leaves the other. EventCycle carries the X the cycling cost was
// announced with, and the trigger's clause is sized by it as the trigger
// is put on the stack (#6085, UpToXTargetsFromTheEventAmount): up to X
// target artifacts, and none at all when X is 0. The artifacts still
// legal as it resolves are destroyed together.
//
// No simplification.
func init() {
	cycling := effects.Cycling("{2}{R}")
	cycling.Label = "Cycling {X}{2}{R} ({X}{2}{R}, Discard this card: Draw a card.)"
	cycling.Cost = effects.Plus(effects.ManaCost("{X}{2}{R}"), effects.DiscardThis())
	effects.Register(effects.Spec{
		OracleID:        "05e910aa-cf84-468c-8a48-256588cbbf08",
		Name:            "Rampaging War Mammoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{cycling},
		Triggered: []game.TriggeredAbility{
			effects.UpToXTargetsFromTheEventAmount(effects.Targeting(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Rampaging War Mammoth — cycled: destroy up to X target artifacts",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
					})),
				effects.TargetPermanent("up to X target artifacts", effects.Artifact()))),
		},
	})
}
