package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rebellious Captives — Creature — Human Peasant Ally {1}{G}, 2/2:
//
//	"Exhaust — {6}: Put two +1/+1 counters on this creature, then
//	 earthbend 2. (Target land you control becomes a 0/0 creature with
//	 haste that's still a land. Put two +1/+1 counters on it. When it
//	 dies or is exiled, return it to the battlefield tapped. Activate
//	 each exhaust ability only once.)"
//
// The exhaust is the engine's once-per-object marker (CR 702.177a), so a
// Captives that left and came back may activate it again (the ruling).
// The ability's only target is the earthbend's land: if that land is
// illegal as the ability resolves, the ability doesn't resolve and this
// creature gets no counters either (CR 608.2b, the ruling). Otherwise
// the counters go on this creature first, then the land is earthbent
// (Bitter Work's keyword action, earthbend.go).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "70401db7-edbc-402f-98ef-be74bd316273",
		Name:         "Rebellious Captives",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {6}: Put two +1/+1 counters on this creature, then earthbend 2.",
			Exhaust: true,
			Cost:    ManaCost("{6}"),
			Targets: EarthbendTargets(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target := FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
				return Earthbend{Target: target, N: 2}.Apply(ctx)
			},
		}},
	})
}
