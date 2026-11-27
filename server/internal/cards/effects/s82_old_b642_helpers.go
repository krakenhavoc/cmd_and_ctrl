package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b642_helpers.go — shared bodies for Planeshift's domain cards and
// life-gain cards, batch 642 (#5615).

// OldB642TargetPlayerDomain is DomainCount of the first target player that is
// still legal as the spell resolves (CR 608.2b): "the number of basic land types
// among lands THEY control" (Allied Strategies). ok is false when the target
// is gone.
func OldB642TargetPlayerDomain(ctx *Context) (player uuid.UUID, n int, ok bool) {
	p := TargetedPlayer(ctx)
	if p == uuid.Nil {
		return uuid.Nil, 0, false
	}
	return p, DomainCount(ctx.Game, p), true
}

// OldB642DomainLandwalk is Magnigoth Treefolk's "for each basic land type among
// lands you control, this creature has landwalk of that type" (CR 702.14): a
// layer 6 grant (CR 613.1f) of one landwalk word per type its controller
// controls, each counted once however many lands carry it.
func OldB642DomainLandwalk(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			for _, t := range game.BasicLandTypes {
				if !OldB642ControlsLandType(g, source.Controller, t) {
					continue
				}
				c.Abilities = game.AppendKeywordAbility(c.Abilities, source.TextKeyword(OldB642LandwalkWord(t)))
			}
		},
	}
}

// OldB642LandwalkWord is the keyword word of a basic land type's landwalk.
func OldB642LandwalkWord(landType string) string {
	switch landType {
	case "Plains":
		return "plainswalk"
	case "Island":
		return "islandwalk"
	case "Swamp":
		return "swampwalk"
	case "Mountain":
		return "mountainwalk"
	default:
		return "forestwalk"
	}
}

// OldB642ControlsLandType reports whether `controller` controls a land of the
// basic land type, read from the layered land.
func OldB642ControlsLandType(g *game.Game, controller uuid.UUID, landType string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsLand() && c.HasSubtype(landType) {
			return true
		}
	}
	return false
}

// OldB642GainPerCreatureOfColors is "you gain `per` life for each creature
// target opponent controls that is any of `colors`" as an enters trigger body
// (Honorable Scout): a creature of both colors counts once, and the count is
// of the target opponent's creatures as the trigger resolves (CR 608.2c).
func OldB642GainPerCreatureOfColors(per int, colors ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		p := TargetedPlayer(ctx)
		if p == uuid.Nil {
			return nil
		}
		n := CountControlled(g, p, func(c game.Card) bool {
			if !c.IsCreature() {
				return false
			}
			for _, col := range colors {
				if c.HasColorWord(col) {
					return true
				}
			}
			return false
		})
		return GainLife{Player: item.Controller, Amount: per * n}.Apply(ctx)
	}
}

// OldB642HalfLifeRoundedUp is half of `life`, rounded up (CR 107.1a); nothing
// for a life total of zero or less.
func OldB642HalfLifeRoundedUp(life int) int {
	if life <= 0 {
		return 0
	}
	return (life + 1) / 2
}
