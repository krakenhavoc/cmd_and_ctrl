package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stegron the Dinosaur Man — Legendary Creature — Dinosaur Villain {4}{R},
// 5/4 (Marvel Super Heroes Commander, #4455):
//
//	"Menace
//	 Dinosaur Formula — {1}{R}, Discard this card: Until end of turn,
//	 target creature you control gets +3/+1 and becomes a Dinosaur in
//	 addition to its other types."
//
// A hand ability (Born to Drive's channel shape): the discard is the cost,
// and the target is chosen from the stack, so Stegron himself is already in
// the graveyard. The pump and the extra type are one until-end-of-turn
// effect locked in at resolution (CR 611.2c, layer 4 and layer 7c), and a
// target that is gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5c94b616-8edf-465c-b316-4599df95c52a",
		Name:            "Stegron the Dinosaur Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "Dinosaur Formula — {1}{R}, Discard this card: Until end of turn, target creature you control gets +3/+1 and becomes a Dinosaur in addition to its other types.",
			Cost:    game.AbilityCost{Mana: "{1}{R}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: ForTargets(BenefitUntilEOT(0, 3, 1)),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return untilEndOfTurn(ctx, id, nil, "Stegron the Dinosaur Man — +3/+1 and a Dinosaur",
					game.ModifyPTMod(3, 1), game.AddSubtypesMod(ctx.TextSubtype("Dinosaur")))
			},
		}},
	})
}
