package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sheoldred, Whispering One — Legendary Creature — Phyrexian Praetor
// {5}{B}{B}, 6/6:
//
//	"Swampwalk
//	 At the beginning of your upkeep, return target creature card from
//	 your graveyard to the battlefield.
//	 At the beginning of each opponent's upkeep, that player sacrifices
//	 a creature of their choice."
//
// Three parts, all existing shapes:
//
//   - Swampwalk is a printed keyword, enforced by the block check
//     (game/landwalk.go).
//   - The reanimation is a targeted upkeep trigger (Oversold Cemetery's
//     shape). It is mandatory, so with no creature card in the graveyard
//     the ability has no target and is simply not put on the stack
//     (CR 603.3d). The card comes back under its OWNER's control, which
//     for "from YOUR graveyard" is the controller.
//   - The edict fires on an opponent's upkeep and makes THAT player
//     choose, not each opponent: ev.Actor is the upkeep's player, and
//     the prompt goes to them (Mana Vortex's per-upkeep sacrifice). It
//     is not targeted, so hexproof and protection stay out of it, and a
//     player with no creature is never asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9218b56d-aaec-482f-99e9-d95d227bfe25",
		Name:            "Sheoldred, Whispering One",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Sheoldred, Whispering One — return a creature card from your graveyard to the battlefield",
					sheoldredReanimate),
				TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature())),
			On(game.EventBeginUpkeep, ByAnOpponent,
				"Sheoldred, Whispering One — that player sacrifices a creature",
				sheoldredEdict),
		},
	})
}

// sheoldredReanimate returns the target to the battlefield, if it is
// still a legal target (CR 608.2b).
func sheoldredReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
		}
	}
	return nil
}

// sheoldredEdict asks the player whose upkeep it is to sacrifice a
// creature of their choice.
func sheoldredEdict(g *game.Game, item *game.StackItem) error {
	player := NewContext(g, item).Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	g.PlayerSacrificesForEffect(item.SourceCardID, player, sacrificeSpec("a creature", Creature()),
		"Sheoldred, Whispering One — sacrifice a creature")
	return nil
}
