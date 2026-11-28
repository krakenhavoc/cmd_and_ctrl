package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Royal Decree — Enchantment {2}{W}{W} (Alliances, S82 OLD batch 818):
//
//	"Cumulative upkeep {W}
//	 Whenever a Swamp, Mountain, black permanent, or red permanent becomes
//	 tapped, this enchantment deals 1 damage to that permanent's controller."
//
// A permanent becomes tapped when it is tapped for any reason, attacking
// included (a vigilance attacker does not). The damage goes to the
// controller the permanent has as the trigger resolves, and Royal Decree is
// the source.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fca14cbd-5ffc-441f-bccb-7ec22b03acbb",
		Name:         "Royal Decree",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Royal Decree — cumulative upkeep {W}", "{W}"),
			{
				Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
				AppliesTo: effects.OldB818SwampMountainBlackOrRedBecameTapped,
				Key:       "Royal Decree — 1 damage to that permanent's controller",
				Effect:    effects.OldB818DamageToTheTappedPermanentsController(1),
			},
		},
	})
}
