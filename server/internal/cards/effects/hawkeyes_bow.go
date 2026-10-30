package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hawkeye's Bow — Artifact — Equipment {R} (Marvel Super Heroes, #4455):
//
//	"Equipped creature gets +1/+0 and has reach.
//	 Whenever equipped creature becomes tapped, it deals 1 damage to each
//	 opponent.
//	 Equip {1}"
//
// Bone Saw's pump and the ordinary equip ability, plus a reach grant. The
// trigger is Contaminated Ground's "becomes tapped" (for mana, by an
// effect, or by attacking, and not for a vigilance attacker, CR 701.26a)
// pointed at whatever the Bow is attached to. The damage is dealt by the
// equipped creature, not by the Bow ("it deals"), so lifelink or deathtouch
// on the creature applies and a creature that has left is still the source
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83798eb8-b14f-4620-bce2-1a6804a24dc7",
		Name:         "Hawkeye's Bow",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0), GrantToAttached("reach")},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: EnchantedPermanentBecameTapped,
			Key:       "Hawkeye's Bow — the equipped creature deals 1 damage to each opponent",
			Effect:    mshA3TappedCreatureDamagesEachOpponent,
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}

// mshA3TappedCreatureDamagesEachOpponent is "it deals 1 damage to each
// opponent", said of the creature whose tapping fired the trigger.
func mshA3TappedCreatureDamagesEachOpponent(g *game.Game, item *game.StackItem) error {
	creature := item.Trigger.Event.CardID
	ctx := NewContext(g, item)
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if err := g.DealDamageToPlayerForEffect(creature, opp, 1); err != nil {
				return err
			}
		}
		return nil
	})
}
