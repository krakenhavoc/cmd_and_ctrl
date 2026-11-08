package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashroot Animist — Creature — Lizard Druid {2}{R}{G}, 4/4 (Foundations,
// #4465):
//
//	"Trample
//	 Whenever this creature attacks, another target creature you control
//	 gains trample and gets +X/+X until end of turn, where X is this
//	 creature's power."
//
// X is read as the trigger resolves (CR 608.2h), Picard, Leading by
// Example's reading: the Animist's current power while it is on the
// battlefield, its last-known power once it has left. "Another" is object
// identity (Another), so the Animist can't pick itself. The trample is a
// layer-6 grant and the pump a layer-7c one; a target gone by resolution
// gets neither (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3f008b57-6471-4785-9571-3b90193da7e2",
		Name:            "Ashroot Animist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Ashroot Animist — another target creature you control gains trample and gets +X/+X",
				Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"},
						Label: "Ashroot Animist — trample"}).Apply(ctx); err != nil {
						return err
					}
					x := sourcePowerNow(ctx)
					return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Ashroot Animist — +X/+X"}.Apply(ctx)
				}))),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
