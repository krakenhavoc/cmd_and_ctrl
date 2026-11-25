package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b157_helpers.go — shared bodies for Core Set 2020, batch 157 (#5518).

// OldB157ExileTargetAndSameName is "exile target creature … and all other
// creatures that player controls with the same name as that creature": the
// target and every creature its controller has with its post-layer name, as one
// exile (CR 400.7), the set read as the effect resolves (CR 611.2c). `then` is
// the clause after the exile, told the exiled cards. A face-down target names
// no one, and nothing face-down shares its name. It returns false when no
// target is left to exile (CR 608.2b).
func OldB157ExileTargetAndSameName(ctx *Context, then game.KeyedTail) (bool, error) {
	target, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return false, nil
	}
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return false, nil
	}
	victim, name, faceDown := c.Controller, c.Name, c.FaceDownIsPermanent()
	return true, ExileAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, o game.Card) bool {
			if o.InstanceID == target {
				return true
			}
			return !faceDown && !o.FaceDownIsPermanent() && o.IsCreature() &&
				o.Controller == victim && game.HasName(o, name)
		},
		ThenTail: then,
	}.Apply(ctx)
}

// OldB157YouOrYourPermanentIsTargeted is "you and/or at least one permanent
// you control becomes the target of a spell or ability an opponent controls":
// an EventBecomesTarget whose actor is an opponent and whose target is the
// source's controller or a permanent they control (CR 115.1).
func OldB157YouOrYourPermanentIsTargeted(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == you {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.Actor); p == nil {
		return false
	}
	if ev.CardID == uuid.Nil {
		return ev.Target == you
	}
	if !OnBattlefield(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == you
}
