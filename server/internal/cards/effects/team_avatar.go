package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Team Avatar — Enchantment {2}{W} (Avatar: The Last Airbender, #4459):
//
//	"Whenever a creature you control attacks alone, it gets +X/+X until
//	 end of turn, where X is the number of creatures you control.
//	 {2}{W}, Discard this card: It deals damage equal to the number of
//	 creatures you control to target creature."
//
// The trigger is Agents of S.H.I.E.L.D.'s: "alone" is game.AttackedAlone
// (CR 506.5), read when the trigger is harvested, so a creature declared
// beside another never triggers it even if the other is removed from
// combat (the printed ruling). X is counted as the trigger resolves
// (CR 608.2h, the printed ruling), the attacker itself included, and does
// not change later in the turn. An attacker that left in response gets
// nothing.
//
// The second ability is a hand-zone activation with the discard in its
// cost (Action News Crew's channel shape): the card itself, now in the
// graveyard, is the source of the damage, and the count is read as the
// ability resolves (the printed ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f95f7b0c-2dd6-46e1-ae02-f51d275e7120",
		Name:         "Team Avatar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone,
				"Team Avatar — the lone attacker gets +X/+X until end of turn, where X is the number of creatures you control",
				teamAvatarPump),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}, Discard this card: It deals damage equal to the number of creatures you control to target creature.",
			Cost:    game.AbilityCost{Mana: "{2}{W}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target := FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: target, Amount: teamAvatarCreatures(g, item)}.Apply(ctx)
			},
		}},
	})
}

// teamAvatarCreatures is "the number of creatures you control".
func teamAvatarCreatures(g *game.Game, item *game.StackItem) int {
	return countControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
}

func teamAvatarPump(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !onBattlefield(g, attacker) {
		return nil
	}
	x := teamAvatarCreatures(g, item)
	return BoostUntilEOT{
		Target:    attacker,
		Power:     x,
		Toughness: x,
		Label:     "Team Avatar — +X/+X",
	}.Apply(NewContext(g, item))
}
