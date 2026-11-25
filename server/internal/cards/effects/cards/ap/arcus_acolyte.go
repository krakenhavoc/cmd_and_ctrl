package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcus Acolyte — Creature — Human Cleric Archer {G}{W}, 2/2 (Modern
// Horizons 2, #5501):
//
//	"Reach, lifelink
//	 Outlast {G/W} ({G/W}, {T}: Put a +1/+1 counter on this creature. Outlast
//	 only as a sorcery.)
//	 Each other creature you control without a +1/+1 counter on it has
//	 outlast {G/W}."
//
// Outlast is the shared sorcery-speed activation (CR 702.107a). The grant is
// a layer 6 ability grant over your other creatures with no +1/+1 counter,
// re-read on every recompute, so it ends the moment a creature gets one (its
// activation already on the stack still resolves). The Acolyte has outlast
// itself, printed, whether or not it has a counter.
//
// No simplification.
func init() {
	const grant = "arcus-acolyte/outlast"
	effects.Register(effects.Spec{
		OracleID:        "628f2deb-519e-4810-a7e5-e31316e9a4e0",
		Name:            "Arcus Acolyte",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "lifelink"},
		Activated:       []effects.ActivatedAbility{effects.RprB04Outlast("{G/W}")},
		Grants: []effects.AbilityGrant{{
			Key:       grant,
			Text:      "Outlast {G/W}",
			Activated: []effects.ActivatedAbility{effects.RprB04Outlast("{G/W}")},
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.IsCreature() &&
					target.Controller == source.Controller && target.Counters[game.CounterPlusOne] <= 0
			}, grant),
		},
	})
}
