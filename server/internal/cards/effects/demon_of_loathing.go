package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demon of Loathing — Creature — Demon {5}{B}{B}, 7/7 (Theros Beyond Death, #5514):
//
//	"Flying, trample
//	 Whenever this creature deals combat damage to a player, that
//	 player sacrifices a creature of their choice."
//
// Flying and trample are the printed keywords. The combat-damage trigger
// is Cabal Executioner's edict: the damaged player sacrifices a creature
// of their choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2e55fc03-779c-4a71-beb7-aa0ade738c24",
		Name:            "Demon of Loathing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WheneverThisDealsCombatDamageToAPlayer("Demon of Loathing — that player sacrifices a creature",
				func(g *game.Game, item *game.StackItem) error {
					victim := NewContext(g, item).Trigger().Event.Target
					if g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					g.PlayerSacrificesForEffect(item.SourceCardID, victim, sacrificeSpec("a creature", Creature()),
						"Demon of Loathing — sacrifice a creature")
					return nil
				})),
		},
	})
}
