package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bone Shaman — Creature — Giant Shaman {2}{R}{R}, 3/3:
//
//	"{B}: Until end of turn, this creature gains "Creatures dealt damage
//	 by this creature this turn can't be regenerated this turn.""
//
// The gained ability is an ADR 0093 bundle given by a duration grant, so
// it is a real layer-6 ability of Bone Shaman: a later "loses all
// abilities" takes it away, and it is gone if Bone Shaman leaves the
// battlefield. The bundle's one slot is
// AbilityGrant.DealtDamageCantBeRegenerated (#1946), which the
// regeneration gate reads live against the turn's record of which
// objects Bone Shaman damaged. That makes it cover a creature Bone Shaman
// damaged before the ability was activated (block, then pay {B} with the
// regeneration on the stack), combat or not, and stop covering every
// creature the moment the ability is gone. A creature that was damaged
// and then flickered is a new object (CR 400.7). A refused regeneration
// shield is not used up (CR 701.19c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "690ba846-964e-44ba-8637-9a1ce62f5931",
		Name:         "Bone Shaman",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:                          boneShamanNoRegenGrant,
			DealtDamageCantBeRegenerated: DealtDamageByThisCantBeRegenerated("Creatures dealt damage by this creature this turn can't be regenerated this turn."),
			Text:                         "Creatures dealt damage by this creature this turn can't be regenerated this turn.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B}: Until end of turn, this creature gains \"Creatures dealt damage by this creature this turn can't be regenerated this turn.\"",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return GrantAbilitiesFor{
					Target: item.SourceCardID,
					Keys:   []string{boneShamanNoRegenGrant},
					Label:  "Bone Shaman — creatures it damaged can't be regenerated",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

const boneShamanNoRegenGrant = "bone-shaman/no-regeneration"
