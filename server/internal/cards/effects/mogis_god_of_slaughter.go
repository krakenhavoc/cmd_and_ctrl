package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mogis, God of Slaughter — Legendary Enchantment Creature — God
// {2}{B}{R}, 7/5:
//
//	"Indestructible
//	 As long as your devotion to black and red is less than seven,
//	 Mogis isn't a creature.
//	 At the beginning of each opponent's upkeep, Mogis deals 2 damage to
//	 that player unless they sacrifice a creature of their choice."
//
// The God clause is the shared Theros static with a two-colour,
// threshold-seven count (godUnlessDevotionToColors, theros_gods.go;
// CR 700.5: a hybrid symbol of both colours is one symbol).
//
// The upkeep trigger is a punisher: the opponent whose upkeep it is
// chooses at resolution between taking 2 damage and sacrificing a
// creature of their choice. Damage is first, because the option list's
// first branch must be one the chooser can always take; a player with
// no creature is not offered the sacrifice and just takes the damage
// (CR 608.2, as much as possible). The damage is Mogis's, as printed,
// whether or not it is a creature right now. The victim is read at
// resolution off the item's carried trigger context.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e5912d21-e188-4b9e-8e92-a6e8196353c6",
		Name:            "Mogis, God of Slaughter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{godUnlessDevotionToColors(7, "B", "R")},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, ByAnOpponent,
				"Mogis, God of Slaughter — 2 damage to that player unless they sacrifice a creature",
				mogisUpkeepPunisher),
		},
	})
}

const mogisBranchSacrifice = 1

// mogisUpkeepPunisher asks the opponent whose upkeep it is whether to
// take the damage or sacrifice a creature.
func mogisUpkeepPunisher(g *game.Game, item *game.StackItem) error {
	victim := item.Trigger.Event.Actor
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	ctx := NewContext(g, item)
	if len(creaturesControlledByPlayer(g, victim)) == 0 {
		return mogisDamage(ctx, victim)
	}
	return PickOption{
		Player:   victim,
		Question: "Mogis, God of Slaughter — take 2 damage, or sacrifice a creature?",
		Options: []game.ChoiceOption{
			{Label: "Take 2 damage"},
			{Label: "Sacrifice a creature"},
		},
		Then: mogisAnswered(victim),
	}.Apply(ctx)
}

func mogisAnswered(victim uuid.UUID) func(ctx *Context, index int) error {
	return func(ctx *Context, index int) error {
		if index == mogisBranchSacrifice {
			creatures := creaturesControlledByPlayer(ctx.Game, victim)
			if len(creatures) > 0 {
				return SacrificeChoice{
					Player:     victim,
					Candidates: creatures,
					Question:   "Mogis, God of Slaughter — sacrifice a creature",
				}.Apply(ctx)
			}
		}
		return mogisDamage(ctx, victim)
	}
}

func mogisDamage(ctx *Context, victim uuid.UUID) error {
	return DealDamage{Source: ctx.Source(), Target: victim, Amount: 2}.Apply(ctx)
}
