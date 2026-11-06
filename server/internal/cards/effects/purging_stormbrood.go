package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Purging Stormbrood // Absorb Essence — Tarkir: Dragonstorm (#4463), an
// omen card (CR 720, #5175):
//
//	Purging Stormbrood — Creature — Dragon {4}{B}, 4/4
//	  "Flying
//	   Ward—Pay 2 life.
//	   When this creature enters, remove all counters from up to one
//	   target creature."
//	Absorb Essence — Instant — Omen {1}{W}
//	  "Target creature gets +2/+2 and gains lifelink and hexproof until
//	   end of turn. (Then shuffle this card into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). Ward—Pay 2 life is
// Hexing Squelcher's WardLife(2); "remove all counters" is Enchanted
// River's Grasp's removeEveryCounterFrom, every kind it carries.
//
// No simplification.

const purgingStormbroodOracleID = "5e8086d5-ea48-4555-b65a-a33d08ba4afd"

func init() {
	Register(Spec{
		OracleID:        purgingStormbroodOracleID,
		Name:            "Purging Stormbrood",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(2), "Purging Stormbrood — ward, pay 2 life"),
			Targeting(WhenThisEnters("Purging Stormbrood — remove all counters from up to one target creature", purgingStormbroodRemoveCounters),
				TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
	Register(Spec{
		OracleID:     purgingStormbroodOracleID + "#1",
		Name:         "Absorb Essence",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "lifelink", "hexproof")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 2, "Absorb Essence — +2/+2, lifelink and hexproof", "lifelink", "hexproof")(ctx.Game, item)
		},
	})
}

// purgingStormbroodRemoveCounters removes every counter from the
// trigger's target, if it is still legal as the trigger resolves
// (CR 608.2b).
func purgingStormbroodRemoveCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCards(item, g) {
		if err := removeEveryCounterFrom(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
