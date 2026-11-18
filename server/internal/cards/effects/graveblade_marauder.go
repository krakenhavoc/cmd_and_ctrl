package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graveblade Marauder — Creature — Human Warrior {2}{B}, 1/4:
//
//	"Deathtouch
//	 Whenever this creature deals combat damage to a player, that player
//	 loses life equal to the number of creature cards in your graveyard."
//
// "That player" is the player the combat damage was dealt to, and the
// creature cards are counted in the controller's own graveyard as the
// trigger resolves (CR 608.2h). Life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33f29fcb-c9ea-4d81-afa2-5c9575b4d58e",
		Name:            "Graveblade Marauder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Graveblade Marauder — that player loses life equal to the number of creature cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					victim := item.Trigger.Event.Target
					if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
						return nil
					}
					n := creatureCardsInGraveyard(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -n)
				}),
		},
	})
}
