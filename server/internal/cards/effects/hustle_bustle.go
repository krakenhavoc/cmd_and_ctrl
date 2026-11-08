package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hustle // Bustle — split card (Murders at Karlov Manor, #4469):
//
//	Hustle — Instant {U/R}: "Target creature attacks or blocks this turn
//	         if able."
//	Bustle — Sorcery {4}{R/G}{R/G}: "Creatures you control get +2/+2 and
//	         gain trample until end of turn. You may turn a creature you
//	         control face up."
//
// Each half is an ADR 0034 face (ADR 0103), Hustle under the bare oracle
// ID and Bustle under "#1"; only one half can be cast (CR 709.3).
//
// Hustle is Into the Fray's attack requirement and Academic Dispute's
// block requirement on one creature, both until end of turn, in one
// effect. A creature can only attack on its controller's turn and only
// block on another player's, so the two never meet; a tapped creature,
// or one with a cost to attack or block, is not forced (the ruling).
//
// Bustle is Overrun's two layers at +2/+2 over the creatures you control
// as it resolves (CR 611.2c), then Zimone, Mystery Unraveler's pick of
// up to one face-down creature you control that an effect can turn face
// up, turned for free (CR 701.40b). The pump comes first, so the
// creature turned up keeps it: it is the same object (CR 708.8).
//
// No simplification.
const hustleBustleOracleID = "e294d85f-caf2-4939-99d3-7ca5bba2b8d1"

func init() {
	Register(Spec{
		OracleID:     hustleBustleOracleID,
		Name:         "Hustle",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ScopedEffectFor{
				Target: id,
				Mods: []game.Mod{
					game.AddAttackRequirementMod(uuid.Nil),
					game.AddBlockRequirementMod(game.BlockRequirementBlocks),
				},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Hustle — attacks or blocks this turn if able",
			}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     hustleBustleOracleID + "#1",
		Name:         "Bustle",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				yours := And(Creature(), YouControl())
				if err := (BoostUntilEOT{Match: yours, Power: 2, Toughness: 2, Label: "Bustle — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"}, Label: "Bustle — trample"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return ChoosePermanents{
					Question: "Bustle — you may turn a creature you control face up",
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						var ids []uuid.UUID
						for _, id := range faceDownPermanentsThatCanTurnUp(g, of) {
							if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
								ids = append(ids, id)
							}
						}
						if len(ids) == 0 {
							return nil, 0, 0
						}
						return ids, 0, 1
					},
					Keyed: bustleTurnedFaceUp,
				}.Apply(ctx)
			}),
		),
	})
}

// bustleTurnedFaceUp turns the chosen creature face up, keyed (#4290).
var bustleTurnedFaceUp = RegisterPermanentsPick("prompt-run/hustle-bustle-turn-face-up", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		return TurnFaceUp{Targets: picked.Cards()}.Apply(ctx)
	},
})
