package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauthi Cutthroat — Creature — Dauthi Minion {1}{B}, 1/1
// (Exodus, #5626):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 {1}{B}, {T}: Destroy target creature with shadow."
//
// Shadow is a printed keyword. The ability is Merfolk Assassin's row
// with shadow in place of islandwalk and a mana cost beside the tap:
// HasKeyword reads the creature's current abilities, so a creature
// granted shadow is a legal target and one that loses it in response is
// not (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5fdde593-54cb-4186-974b-df4383f7efe2",
		Name:            "Dauthi Cutthroat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Destroy target creature with shadow.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with shadow", effects.HasKeyword("shadow")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
