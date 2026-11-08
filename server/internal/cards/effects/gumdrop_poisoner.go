package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gumdrop Poisoner // Tempt with Treats — Creature — Human Warlock {2}{B},
// 3/2 // Instant — Adventure {B} (Wilds of Eldraine, #4471):
//
//	Gumdrop Poisoner — "Lifelink
//	 When this creature enters, up to one target creature gets -X/-X until
//	 end of turn, where X is the amount of life you gained this turn."
//	Tempt with Treats — "Create a Food token. (Then exile this card. You
//	 may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). X is read as the trigger resolves from the turn's life-gain
// tally (b15LifeGainedThisTurn, The Gaffer's count), so life gained in
// response to the trigger counts; with no life gained the target is left
// alone. The shrink lands only on a target that is still legal
// (vsPumpTheTarget, CR 608.2b). Tempt with Treats makes the catalog's
// Food token (CR 111.10b); the Adventure lifecycle is the engine's
// (CR 715.3d).
//
// No simplification.
const gumdropPoisonerOracleID = "4f5ebbb8-f49b-488f-bb85-888aaa918bdd"

func init() {
	const label = "Gumdrop Poisoner — up to one target creature gets -X/-X, X the life you gained this turn"
	Register(Spec{
		OracleID:        gumdropPoisonerOracleID,
		Name:            "Gumdrop Poisoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				x := b15LifeGainedThisTurn(g, item.Controller)
				if x <= 0 {
					return nil
				}
				return vsPumpTheTarget(NewContext(g, item), -x, -x, "Gumdrop Poisoner — -X/-X")
			}), UpToOneTargetCreature("up to one target creature")),
		},
	})
	Register(Spec{
		OracleID:     gumdropPoisonerOracleID + "#1",
		Name:         "Tempt with Treats",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve:    ResolveSteps(CreateToken{Template: FoodToken(), N: 1}),
	})
}
