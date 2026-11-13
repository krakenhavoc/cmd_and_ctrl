package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cabal Executioner — Creature — Human Cleric {2}{B}{B}, 2/2 (Onslaught):
//
//	"Whenever this creature deals combat damage to a player, that player
//	 sacrifices a creature of their choice.
//	 Morph {3}{B}{B}"
//
// The damaged player is read off the damage event, and they pick the
// creature as the trigger resolves, Sheoldred, Whispering One's edict
// prompt; a player with no creature sacrifices nothing. The edict is not
// targeted, so hexproof does not stop it. Face down it is a 2/2 with no
// text (CR 708.2), so it connects without the trigger until it is turned
// face up for {3}{B}{B}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b7817579-df4d-41db-bb90-2aabac58f299",
		Name:             "Cabal Executioner",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{3}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WheneverThisDealsCombatDamageToAPlayer("Cabal Executioner — that player sacrifices a creature",
				func(g *game.Game, item *game.StackItem) error {
					victim := NewContext(g, item).Trigger().Event.Target
					if g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					g.PlayerSacrificesForEffect(item.SourceCardID, victim, sacrificeSpec("a creature", Creature()),
						"Cabal Executioner — sacrifice a creature")
					return nil
				})),
		},
	})
}
