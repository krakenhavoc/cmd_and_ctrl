package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smoke Bomb — Artifact {3} (Assassin's Creed, #5948):
//
//	"Flash
//	 All creatures have shroud. (They can't be the targets of spells or
//	 abilities.)
//	 At the beginning of your upkeep, sacrifice this artifact. When you do,
//	 target creature you control can't be blocked this turn."
//
// A layer 6 shroud grant over every creature on the battlefield, every
// player's (Absolute Grace's shape). The upkeep trigger is mandatory: the
// Bomb is sacrificed as it resolves, and only if that happened is the
// reflexive "when you do" trigger put on the stack, with its own target
// chosen then (CR 603.12). A Bomb that left and came back is a new object and
// is not sacrificed (CR 400.7). The creature, once it has the evasion, keeps
// it until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c262e015-f379-484f-894f-1a1deb3c75cc",
		Name:            "Smoke Bomb",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B29AllCreatures, "shroud"),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Smoke Bomb — sacrifice it; when you do, target creature you control can't be blocked this turn", smokeBombUpkeep),
		},
	})
}

func smokeBombUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	source := ctx.Source()
	if source == uuid.Nil || ctx.IsNewSourceObjectAsThis(source) {
		return nil
	}
	if z := g.FindCardZoneForEffect(source); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := g.SacrificePermanentForEffect(source); err != nil {
		return err
	}
	// The Bomb's shroud is gone with it: refresh the layers before the
	// reflexive trigger chooses a target.
	g.RecomputeLayersIfStaleLocked()
	return effects.ReflexiveTrigger{
		Label: "Smoke Bomb — target creature you control can't be blocked this turn",
		Body:  smokeBombUnblockableBody,
	}.Apply(ctx)
}

// smokeBombUnblockableBody is the reflexive trigger's registration.
var smokeBombUnblockableBody = game.ReflexiveBody("smoke-bomb/unblockable", effects.SimpleBody(smokeBombUnblockable),
	func(_ uuid.UUID, _ game.EffectParams) *game.TargetSpec {
		return effects.PermanentYouControl("target creature you control", effects.Creature())
	})

func smokeBombUnblockable(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return effects.RestrictUntilEOT{
		Target:       t.ID,
		Restrictions: game.CantBeBlocked,
		Label:        "Smoke Bomb — can't be blocked",
	}.Apply(ctx)
}
