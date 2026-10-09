package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "5e4d69de-1876-49a8-8ec6-a891d4a84ccf",
		Name:            "Komainu Battle Armor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisOrEquippedCreatureDealsCombatDamageToAPlayer,
				"Komainu Battle Armor — goad each creature that player controls", komainuBattleArmorGoad),
		},
		Activated: Reconfigure("{4}"),
	})
}

func komainuBattleArmorGoad(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Trigger().Event.Target
	return GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool {
		return c.Controller == player
	})
}
