package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellspark Elemental — Creature — Elemental {1}{R}, 3/1 (Conflux,
// #5587):
//
//	"Trample, haste
//	 At the beginning of the end step, sacrifice this creature.
//	 Unearth {1}{R} ({1}{R}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the next
//	 end step or if it would leave the battlefield. Unearth only as a
//	 sorcery.)"
//
// Lightning Serpent's end-step sacrifice, which triggers at the
// beginning of every end step, not only its controller's, as printed.
// Unearth is the shared graveyard ability (unearth.go, CR 702.84a): an
// unearthed Elemental that is sacrificed is exiled instead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a7ba88c6-81e7-4ab7-9c9b-308831e56833",
		Name:            "Hellspark Elemental",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, "Hellspark Elemental — sacrifice it", func(g *game.Game, item *game.StackItem) error {
				return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{Unearth("{1}{R}")},
	})
}
