package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wizard Class — Enchantment — Class, {U}:
//
//	(Gain the next level as a sorcery to add its ability.)
//	You have no maximum hand size.
//	{2}{U}: Level 2
//	When this Class becomes level 2, draw two cards.
//	{4}{U}: Level 3
//	Whenever you draw a card, put a +1/+1 counter on target creature
//	you control.
//
// The first Class in the catalog, and the one the CR 716 machinery
// was written against (ADR 0071). All three lines are here and none
// of them needed anything Class-specific:
//
//   - Level 1 is the printed NoMaxHandSize slot, ungated. CR 716.2b
//     makes every Class level 1 the moment it enters, so the level-1
//     line has nothing to wait for and carries no ActiveWhen. (It
//     could not carry one anyway — NoMaxHandSize is a bool on the
//     Spec, not a StaticAbility, for the reason its own comment
//     gives.)
//   - Level 2 is an ordinary triggered ability watching the level
//     event, gated at level 2. The gate is what makes "when this
//     becomes level 2" fire exactly once: the event is emitted as the
//     level is set, the trigger exists from that moment, and it never
//     exists again at a lower level.
//   - Level 3 is an ordinary targeted trigger, gated at level 3. At
//     level 2 it is not a trigger the harvester can see, so the two
//     cards drawn by the level-2 ability do NOT put counters on
//     anything — which is the card, and which is the whole reason the
//     gate is a gate rather than a predicate inside AppliesTo.
//
// The level-up abilities are LevelUp, the one constructor: sorcery
// speed and "activate only if this Class is level N-1" come with it
// (CR 716.2d-e) and are not written out here.
func init() {
	Register(Spec{
		OracleID:      "36f68aa3-9955-46f1-bc87-497f16ef5222",
		Name:          "Wizard Class",
		Completeness:  CompletenessFull,
		NoMaxHandSize: true,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{2}{U}")),
			LevelUp(3, ManaCost("{4}{U}")),
		},
		Triggered: []game.TriggeredAbility{
			BecomesLevel(2, "Wizard Class — level 2: draw two cards",
				Do(DrawCards{N: 2})),
			AtLevel(3, Targeting(
				WheneverYouDraw("Wizard Class — +1/+1 counter on target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						target, ok := NewContext(g, item).ClauseTarget(0)
						if !ok {
							return nil
						}
						return AddCounter{Target: target.ID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
					}),
				TargetCreature("target creature you control", YouControl()),
			)),
		},
	})
}
