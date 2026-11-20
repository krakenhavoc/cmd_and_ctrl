package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Komainu Battle Armor — Artifact Creature — Equipment Dog {2}{R}, 2/2:
//
//	"Menace
//	 Equipped creature gets +2/+2 and has menace.
//	 Whenever this creature or equipped creature deals combat damage to
//	 a player, goad each creature that player controls.
//	 Reconfigure {4}"
//
// The player is the one the damage was dealt to (the event's Target),
// and "each creature that player controls" is read as the trigger
// resolves. The goad is the trigger controller's (GoadAllMatching, CR
// 701.15). Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e4d69de-1876-49a8-8ec6-a891d4a84ccf",
		Name:            "Komainu Battle Armor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisOrEquippedCreatureDealsCombatDamageToAPlayer,
				"Komainu Battle Armor — goad each creature that player controls", komainuBattleArmorGoad),
		},
		Activated: effects.Reconfigure("{4}"),
	})
}

func komainuBattleArmorGoad(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Trigger().Event.Target
	return effects.GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool {
		return c.Controller == player
	})
}
