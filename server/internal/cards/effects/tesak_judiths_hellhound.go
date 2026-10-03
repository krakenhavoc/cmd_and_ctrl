package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tesak, Judith's Hellhound — Legendary Creature — Elemental Dog
// {3}{R}, 3/3:
//
//	"Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)
//	 Other Dogs you control have unleash.
//	 Creatures you control with counters on them have haste.
//	 Whenever Tesak attacks, add {R} for each attacking creature."
//
// Unleash is the engine's keyword (ADR 0109 §10). "Other Dogs you
// control have unleash" is a layer-6 tribal grant, so a Dog entering
// under Tesak has unleash as it would exist on the battlefield and is
// asked as it enters (CR 614.12); a Dog that prints unleash too is asked
// twice (CR 113.2c). "Creatures you control with counters on them" is
// any kind of counter, Tesak included. The mana is counted as the
// trigger resolves, every attacking creature whoever controls it, and
// empties with the pool at the end of the step (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4117e428-0c78-4dea-9019-dd1836fbb03e",
		Name:            "Tesak, Judith's Hellhound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUnleash},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Dog"}, Others: true, YoursOnly: true}, game.KeywordUnleash),
			CreaturesYouControlWithCountersHave("", false, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Tesak, Judith's Hellhound — add {R} for each attacking creature", func(g *game.Game, item *game.StackItem) error {
				n := len(b26AttackingCreatures(g))
				if n == 0 {
					return nil
				}
				return AddMana{Player: item.Controller, Produced: strings.Repeat("{R}", n)}.Apply(NewContext(g, item))
			}),
		},
	})
}
