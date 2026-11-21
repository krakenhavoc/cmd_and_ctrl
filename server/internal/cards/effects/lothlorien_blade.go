package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lothlórien Blade — Artifact — Equipment {3} (Tales of Middle-earth
// Commander, #5948):
//
//	"Whenever equipped creature attacks, it deals damage equal to its power
//	 to target creature defending player controls.
//	 Equip Elf {2}
//	 Equip {5}"
//
// Mage Slayer's trigger (attachedCreatureAttacked, the EQUIPPED CREATURE
// deals the damage, by object, so its lifelink and deathtouch apply, and its
// power is read as the trigger resolves, its last-known power once it has
// left, CR 608.2h), aimed at a creature the defending player controls, as
// Armix's target clause is: the player attacked, read off the attack event.
// The two equips are Bilbo's Ring's pair: Equip Elf {2} (EquipOnlyAbility over
// an Elf you control) and the plain Equip {5}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2a39123-8b70-4678-8675-ba79f37a5db8",
		Name:         "Lothlórien Blade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			},
			TargetsFrom: func(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
				return TargetCreature("target creature defending player controls", ControlledBy(b17DefendingPlayer(g, tc.Event)))
			},
			Key:    "Lothlórien Blade — equipped creature deals damage equal to its power to target creature defending player controls",
			Effect: lothlorienBladeDamage,
		}},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip Elf {2}", "{2}",
				TargetCreature("target Elf you control", YouControl(), Subtype("Elf"))),
			EquipAbility("{5}"),
		},
	})
}

// lothlorienBladeDamage is the trigger body: the equipped creature, named by
// object, deals damage equal to its power to the target if still legal.
func lothlorienBladeDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	creature, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	ref := ctx.Trigger().Object.Ref()
	return DealDamage{
		SourceObject: &ref,
		Target:       target,
		Amount:       creature.Power,
	}.Apply(ctx)
}
