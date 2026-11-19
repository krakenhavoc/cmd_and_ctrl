package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander Mustard — Legendary Creature — Human Soldier {3}{R}{W}, 5/5
// (Ravnica: Clue Edition, #5929):
//
//	"Vigilance
//	 Other Soldiers you control have vigilance, trample, and haste.
//	 {2}{R}{W}: Until end of turn, Soldiers you control gain "Whenever this
//	 creature attacks, it deals 1 damage to defending player.""
//
// The lord half is Hordewing Skaab's TribeFilter, one grant per keyword. The
// activated ability gives a bundle (Spec.Grants) to the Soldiers you control
// as it resolves, the set locked then (CR 611.2c, GrantAbilitiesFor): a
// Soldier that arrives later does not have it, and Mustard is a Soldier
// himself. The granted trigger names the defending player off the attack
// event, so it survives the attacker leaving, and the damage is dealt by the
// attacking creature ("it") as an object.
//
// No simplification.
func init() {
	const grant = "commander-mustard/attack-ping"
	soldiers := TribeFilter{Tribes: []string{"Soldier"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "3fdbe867-63cd-42c5-9843-46b0992b9628",
		Name:            "Commander Mustard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(soldiers, "vigilance"),
			TribalKeywordGrant(soldiers, "trample"),
			TribalKeywordGrant(soldiers, "haste"),
		},
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "Whenever this creature attacks, it deals 1 damage to defending player.",
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Soldier — it deals 1 damage to defending player", func(g *game.Game, item *game.StackItem) error {
					defender := g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target)
					if defender == uuid.Nil || g.PlayerByIDForEffect(defender) == nil {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: defender, Amount: 1}.Apply(NewContext(g, item))
				}),
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}{W}: Until end of turn, Soldiers you control gain \"Whenever this creature attacks, it deals 1 damage to defending player.\"",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{R}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantAbilitiesFor{
					Match: And(Creature(), YouControl(), Subtype("Soldier")),
					Keys:  []string{grant},
					Label: "Commander Mustard — Soldiers gain the attack ping until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
