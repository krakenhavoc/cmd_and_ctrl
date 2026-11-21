package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renegade Silent — Creature — Alien Horror {3}{U}, 3/3 (Doctor Who, #5948):
//
//	"At the beginning of your end step, goad up to one target creature you
//	 don't control and put a +1/+1 counter on this creature. This creature
//	 phases out. (Treat it and anything attached to it as though they don't
//	 exist until your next turn.)"
//
// One trigger with an optional target: the goad is applied to the chosen
// creature (CR 701.15a: it attacks each combat if able and attacks a player
// other than the goader if able, until the goader's next turn), the counter
// goes on the Silent, and then the Silent phases out (PhaseOut) and phases in
// during its controller's next untap step (CR 502.1, CR 702.26). The counter
// and the phasing happen whether or not a target was chosen. If the
// Silent has left by the time the trigger resolves, it gets no counter and
// does not phase out.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b5565a2-be5c-46b1-99df-8094e77446cc",
		Name:         "Renegade Silent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Renegade Silent — goad up to one target creature you don't control, put a +1/+1 counter on this creature, then it phases out",
					renegadeSilentEndStep),
				UpToOneTargetCreature("up to one target creature you don't control", OpponentControls())),
		},
	})
}

func renegadeSilentEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if id, ok := b16FirstLegalTargetCard(ctx); ok {
		b33Goad(g, id, item.Controller)
	}
	if !onBattlefield(g, item.SourceCardID) || !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return PhaseOut{Targets: []uuid.UUID{item.SourceCardID}}.Apply(ctx)
}
