package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Star Athlete — Creature — Human Warrior {1}{R}{R}, 3/2:
//
//	"Menace
//	 Whenever this creature attacks, choose up to one target nonland
//	 permanent. Its controller may sacrifice it. If they don't, this
//	 creature deals 5 damage to that player.
//	 Blitz {3}{R}"
//
// Blitz (CR 702.152, #3772). The punisher choice belongs to the target's
// controller as the trigger resolves, asked through MayChoice. "If they
// don't" is read off what actually happened: a permanent that can't be
// sacrificed (or left in the meantime) was not sacrificed, so the 5
// damage is dealt. The damage's source is this creature, from last-known
// information if it has left (CR 608.2h). With no target chosen, nothing
// happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "fb03befb-b4e4-4ae9-908d-a16d5b03b054",
		Name:             "Star Athlete",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{Blitz("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Star Athlete — sacrifice it, or take 5 damage", starAthletePunish),
				TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1)),
		},
	})
}

// starAthletePunish asks the target's controller whether to sacrifice
// it.
func starAthletePunish(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		return MayChoice{
			Player:   c.Controller,
			Question: "Star Athlete — sacrifice " + c.Name + "? If you don't, Star Athlete deals 5 damage to you.",
			YesLabel: "Sacrifice it",
			NoLabel:  "Take 5 damage",
			OnYes:    starAthleteSacrifice(t.ID, c.Controller),
			OnNo:     starAthleteDamage(c.Controller),
		}.Apply(ctx)
	}
	return nil
}

// starAthleteSacrifice is the yes branch: sacrifice it if it is still
// theirs, and deal the damage if it did not go.
func starAthleteSacrifice(target, player uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		c, ok := ctx.Game.LookupCardForEffect(target)
		if !ok || c.Controller != player || !ctx.Game.Battlefield.Contains(target) {
			return starAthleteDamage(player)(ctx)
		}
		return ctx.Game.SacrificeThenForEffect(ctx.Item.SourceCardID, target, func(g *game.Game, sacrificed bool) error {
			if sacrificed {
				return nil
			}
			return starAthleteDamage(player)(NewContext(g, ctx.Item))
		})
	}
}

// starAthleteDamage is "this creature deals 5 damage to that player".
func starAthleteDamage(player uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		return DealDamage{Source: ctx.Item.SourceCardID, Target: player, Amount: 5}.Apply(ctx)
	}
}
