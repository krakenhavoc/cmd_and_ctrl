package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Idyllic Grange — Land — Plains (EDHREC rank 4286):
//
//	"({T}: Add {W}.)
//	 This land enters tapped unless you control three or more other
//	 Plains.
//	 When this land enters untapped, put a +1/+1 counter on target
//	 creature you control."
//
// The white member of the M19 "Grange" cycle, and Gingerbread Cabin's
// older cousin: a basic-type land that pays a small bonus once your
// mana base is mostly one colour. It IS a Plains — the {W} is in
// reminder text because the land type supplies it, and a fetchland can
// find it — so the ability is declared explicitly for one click in the
// client and the intrinsic one the type would give is covered by it.
//
// The condition counts OTHER Plains, which costs nothing extra: the
// Grange is not on the battlefield while its own entry is being
// replaced, so the walk only ever sees the others. It names the LAND
// TYPE, not the basic land: a Hallowed Fountain, a Snow-Covered Plains
// and another Grange all count, and so does any land Urborg's cousin
// turned into a Plains.
//
// "When this land enters UNTAPPED" reads the tapped flag as the
// harvester sees the ETB — the enters-tapped replacement has already
// run by then, so the flag is the answer. A Grange that entered tapped
// puts no counter anywhere, and one that entered untapped off a
// "lands you control enter untapped" effect does.
//
// The counter clause targets, so it goes through Triggered rather than
// the entry hook: the creature is chosen when the trigger goes on the
// stack. A controller with no creatures never puts the trigger on the
// stack at all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23d349a0-e441-40b8-b634-13e61440a7c8",
		Name:         "Idyllic Grange",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			SelfEntersTappedUnless(func(g *game.Game, controller uuid.UUID) bool {
				return b08LandsWithSubtypeControlled(g, controller, "Plains") >= 3
			}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b27SelfEnteredUntapped(ev, source)
				}, "Idyllic Grange — put a +1/+1 counter on target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
						}
						return nil
					}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
