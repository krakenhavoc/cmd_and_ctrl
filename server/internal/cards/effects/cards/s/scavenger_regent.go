package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scavenger Regent // Exude Toxin — Tarkir: Dragonstorm (#4463), an
// omen card (CR 720, #5175):
//
//	Scavenger Regent — Creature — Dragon {3}{B}, 4/4
//	  "Flying
//	   Ward—Discard a card."
//	Exude Toxin — Sorcery — Omen {X}{B}{B}
//	  "Each non-Dragon creature gets -X/-X until end of turn. (Then
//	   shuffle this card into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). Ward—Discard a card
// is Ward(WardDiscard(1)), Graveyard Trespasser's. Exude Toxin is Toxic
// Deluge's -X/-X with the Dragons left out; the affected set is fixed as
// it resolves (CR 611.2c), so a creature that becomes a Dragon later in
// the turn keeps the -X/-X.
//
// No simplification.

const ScavengerRegentOracleID = "7e97a242-0bac-4f18-9228-e98ecca36254"

func init() {
	effects.Register(effects.Spec{
		OracleID:        ScavengerRegentOracleID,
		Name:            "Scavenger Regent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardDiscard(1), "Scavenger Regent — ward, discard a card"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     ScavengerRegentOracleID + "#1",
		Name:         "Exude Toxin",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}},
		XMatters: true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return effects.BoostUntilEOT{
				Match:     effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Dragon"))),
				Power:     -x,
				Toughness: -x,
				Label:     "Exude Toxin — -X/-X",
			}.Apply(ctx)
		},
	})
}
