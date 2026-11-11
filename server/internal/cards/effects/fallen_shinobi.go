package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallen Shinobi — Creature — Zombie Ninja {3}{U}{B}, 5/4 (Outlaws of Thunder
// Junction Commander):
//
//	"Ninjutsu {2}{U}{B}
//	 Whenever this creature deals combat damage to a player, that player
//	 exiles the top two cards of their library. Until end of turn, you may
//	 play those cards without paying their mana costs."
//
// The exile is face up, off the damaged player's library, and the grant is
// the controller's: a free play ({0} in place of the mana cost, a land taken
// with the turn's land drop) at each card's normal timing, ending with the
// turn (Fire Lord Ozai's grant, Free).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fc5f871-a809-4a4a-adfb-eb4cd52837c8",
		Name:         "Fallen Shinobi",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{2}{U}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Fallen Shinobi — that player exiles the top two cards of their library; you may play them without paying their mana costs",
				cmdr2S45FallenShinobiExile),
		},
	})
}

func cmdr2S45FallenShinobiExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return ExileTopWithPermission{
		From:    item.Trigger.Event.Target,
		GrantTo: item.Controller,
		N:       2,
		Free:    true,
	}.Apply(NewContext(g, item))
}
