package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soldevi Steam Beast — Artifact Creature — Beast {5}, 4/2 (Alliances,
// #5634):
//
//	"Whenever this creature becomes tapped, target opponent gains 2 life.
//	 {2}: Regenerate this creature."
//
// "Becomes tapped" watches both a tap event and an attack declaration,
// since attacking taps a creature without one (CR 508.1f); a vigilance
// attacker did not become tapped (ThisBecameTapped). The life goes to the
// opponent targeted as the trigger is put on the stack (CR 603.3d). The
// second ability is a regeneration shield on itself (CR 701.19a), built
// by regenerateThisRow (regenerate_this.go). Regenerating taps the Beast
// (CR 701.19a), so an untapped Beast that regenerates also triggers the
// first ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "935728dd-ad97-47a9-97ca-068af7960d07",
		Name:         "Soldevi Steam Beast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Soldevi Steam Beast — target opponent gains 2 life", targetPlayersGainTwoLife),
				TargetPlayer("target opponent", Opponent())),
		},
		Activated: []ActivatedAbility{regenerateThisRow("{2}")},
	})
}

// targetPlayersGainTwoLife gives each still-legal target player 2 life.
func targetPlayersGainTwoLife(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (GainLife{Player: t.ID, Amount: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
