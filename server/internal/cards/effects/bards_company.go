package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bard's Company — Creature — Human Citizen {2}{W}{U}, 2/3:
//
//	"You may cast this spell as though it had flash if you control a
//	 Human.
//	 Other creatures you control get +1/+1.
//	 Whenever this creature enters or attacks, recruit. (Draw a card,
//	 then discard a card. If you discarded a nonland card, create a 1/1
//	 white Human Soldier creature token.)"
//
// The flash clause is Fated Clash's shape: an offer at the printed cost
// that opens the instant-speed window for its own claim
// (AlternativeCost.AsThoughFlash), gated by Condition, read as the cast
// begins (CR 601.3c). The Company is in the hand then, so it is not the
// Human it asks about.
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:     "5c5bfbb2-0e63-4e43-b441-c4878983288f",
		Name:         "Bard's Company",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:           "flash-if-you-control-a-human",
			Label:         "Cast with flash — you control a Human ({2}{W}{U})",
			ManaCost:      "{2}{W}{U}",
			AsThoughFlash: true,
			Condition:     bardsCompanyYouControlAHuman,
		}},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "Other creatures you control get +1/+1.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Bard's Company — recruit", Do(Recruit{})),
		},
	})
}

// bardsCompanyYouControlAHuman is "if you control a Human".
func bardsCompanyYouControlAHuman(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.HasSubtype("Human") {
			return true
		}
	}
	return false
}
