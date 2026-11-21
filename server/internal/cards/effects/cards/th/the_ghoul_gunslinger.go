package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Ghoul, Gunslinger — Legendary Creature — Zombie Mutant Rogue
// {1}{B}{B}, 2/3:
//
//	"First strike
//	 Whenever The Ghoul or another nontoken Zombie or Mutant you control
//	 dies, target player gets two rad counters. If that player is you,
//	 create a Treasure token."
//
// #2809. One trigger watching two conditions: The Ghoul's own death
// (ThisDied), or another NONTOKEN Zombie or Mutant that died under the
// Ghoul controller's control. The creature types, the token bit and
// the controller are read as the creature last existed on the
// battlefield (CR 603.10a), so a creature that was a Zombie only
// through a lord's grant still counts and a stolen one counts for the
// thief.
//
// The player is chosen as the trigger goes on the stack (CR 603.3d).
// "If that player is you" is read as it resolves: it is the Ghoul's
// controller, so the Treasure is theirs. The Ghoul itself is named
// without a token bar, so a token copy of it dying counts.
//
// The whole ability has one target, so if it is gone on resolution
// neither the counters nor the Treasure happen (CR 608.2b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ddfa5c7d-bc47-43e0-9607-b3baa8e6a17d",
		Name:            "The Ghoul, Gunslinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventLTB, ghoulGunslingerCondition,
				"The Ghoul, Gunslinger — target player gets two rad counters; if you, create a Treasure",
				ghoulGunslingerResolve), effects.TargetPlayer("target player")),
		},
	})
}

// ghoulGunslingerCondition is "The Ghoul or another nontoken Zombie or
// Mutant you control dies".
func ghoulGunslingerCondition(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	if effects.ThisDied(ev, source, ch, g) {
		return true
	}
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := effects.DiedCreature(ev, g)
	if !ok || effects.IsToken(dead) || effects.LeftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	// The types are words of the Ghoul's text (CR 612.2, #3984).
	return effects.LeftAsSubtype(ev, dead, source.TextSubtype("Zombie")) || effects.LeftAsSubtype(ev, dead, source.TextSubtype("Mutant"))
}

// ghoulGunslingerResolve gives the target player two rad counters, and
// a Treasure to the controller when that player is the controller.
func ghoulGunslingerResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.TargetedPlayerOf(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := effects.PlayerGetsRadCounters(g, item.Controller, target, 2); err != nil {
		return err
	}
	if target != item.Controller {
		return nil
	}
	return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
}
