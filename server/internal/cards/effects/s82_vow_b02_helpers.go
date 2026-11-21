package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82VowB02ControlsAnother reports whether `controller` controls a
// permanent other than `self` with the creature subtype, read through the
// layers (Resistance Squad's "if you control another Human").
func S82VowB02ControlsAnother(g *game.Game, controller, self uuid.UUID, subtype string) bool {
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.InstanceID != self && c.IsCreature() && c.HasSubtype(subtype) {
			return true
		}
	}
	return false
}

// s82VowB02ControlsAnotherHuman is Resistance Squad's intervening "if",
// re-read as its trigger resolves.
var S82VowB02ControlsAnotherHuman = CondBoard{
	Name: "you control another Human",
	Test: func(ctx *Context) bool {
		return S82VowB02ControlsAnother(ctx.Game, ctx.Controller(), ctx.Source(), "Human")
	},
}

// s82VowB02ArmSlot is one of Arm the Cathars' three picks: +N/+N and
// vigilance until end of turn for the slot's target if it is still legal.
func S82VowB02ArmSlot(ctx *Context, slot, n int) error {
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if err := (BoostUntilEOT{Target: t.ID, Power: n, Toughness: n, Label: "Arm the Cathars — pump until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance"}, Label: "Arm the Cathars — vigilance"}.Apply(ctx)
}
