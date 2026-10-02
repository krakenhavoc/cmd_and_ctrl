package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gríma Wormtongue — Legendary Creature — Human Advisor {2}{B}, 1/4:
//
//	"Your opponents can't gain life.
//	 {T}, Sacrifice another creature: Target player loses 1 life. If
//	 the sacrificed creature was legendary, amass Orcs 2."
//
// "Your opponents can't gain life" is ADR 0107 §5's battlefield static
// (CR 119.7, #1880).
//
// The sacrifice is a COST paid at announce. "The sacrificed creature"
// is read back off the log (b17PermanentSacrificedToPay, Birthing
// Pod's reader) and judged by its last-known information as it last
// existed on the battlefield (CR 608.2h): a creature that was made
// legendary by an effect counts, and one that had lost the supertype
// does not. The amass is the shared primitive (CR 701.47).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f98806e-c6b7-44af-b436-88acf36d25ec",
		Name:         "Gríma Wormtongue",
		Completeness: CompletenessFull,
		CantGainLife: OpponentsCantGainLife(),
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice another creature: Target player loses 1 life. If the sacrificed creature was legendary, amass Orcs 2.",
			Cost:    Plus(TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Targets: TargetPlayer("target player"),
			Effect:  grimaWormtongueEffect,
		}},
	})
}

// grimaWormtongueEffect is the activation's body: the life loss if the
// target is still legal (CR 608.2b), then the amass if the sacrificed
// creature was legendary. The amass is not part of the targeted
// instruction, so it happens even when the player target has become
// illegal — but a fully illegal ability does not resolve at all, and
// with one target that is the same case.
func grimaWormtongueEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -1); err != nil {
			return err
		}
	}
	sacrificed, ok := b17PermanentSacrificedToPay(g, item)
	if !ok {
		return nil
	}
	info, ok := g.LastKnownPermanentForEffect(sacrificed)
	if !ok || !slices.Contains(info.Characteristic.Supertypes, "Legendary") {
		return nil
	}
	return Amass{Subtype: "Orc", N: 2}.Apply(ctx)
}
