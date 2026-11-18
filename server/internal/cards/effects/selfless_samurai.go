package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selfless Samurai — Creature — Fox Samurai {1}{W}, 2/2 (Kamigawa: Neon
// Dynasty):
//
//	"Whenever a Samurai or Warrior you control attacks alone, it gains
//	 lifelink until end of turn.
//	 Sacrifice this creature: Another target creature you control gains
//	 indestructible until end of turn."
//
// The trigger grants lifelink to the lone attacker (CR 506.5). The
// activation is Selfless Savior's: the sacrifice is a cost, "another" is
// object identity, and the target must still be a creature you control as it
// resolves.
//
// No simplification.
func init() {
	const name = "Selfless Samurai"
	Register(Spec{
		OracleID:     "20c9fb06-3941-456f-ba02-4fbde9deb1e2",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, neoB09SamuraiOrWarriorAttacksAlone,
				name+" — the lone attacker gains lifelink until end of turn", selflessSamuraiLifelink),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Another target creature you control gains indestructible until end of turn.",
			Cost:    SacrificeThis(),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect:  firstTargetGainsUntilEndOfTurn(0, name+" — indestructible until end of turn", "indestructible"),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible")),
		}},
	})
}

func selflessSamuraiLifelink(g *game.Game, item *game.StackItem) error {
	attacker, ok := neoB09LoneAttacker(g, item)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{Target: attacker.InstanceID, Keywords: []string{"lifelink"}, Label: "Selfless Samurai — lifelink"}.Apply(NewContext(g, item))
}
