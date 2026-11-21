package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dalek Drone — Artifact Creature — Dalek {3}{B}{B}, 3/3 (Doctor Who, #5948):
//
//	"Flying, menace
//	 Exterminate! — When this creature enters, destroy target creature an
//	 opponent controls. That player loses 3 life."
//
// One enters trigger with a target. "That player" is the target's controller
// as the ability resolves (CR 608.2h, last-known information once it is
// destroyed), and they lose 3 life whether or not the destroy happened
// (regeneration or indestructible stop only the destroy), as long as the
// target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "95d018f4-7f97-4b2c-abb4-cef69031caa1",
		Name:            "Dalek Drone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Dalek Drone — Exterminate! Destroy target creature an opponent controls. That player loses 3 life", dalekDroneExterminate),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

func dalekDroneExterminate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	victim, _ := controllerOfTarget(ctx, id)
	if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	// A zero victim (no controller found) is no player: nothing is lost.
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -3)
}
