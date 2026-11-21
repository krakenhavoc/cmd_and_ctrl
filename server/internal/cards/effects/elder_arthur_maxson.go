package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elder Arthur Maxson — Legendary Creature — Human Knight {1}{W}{B}, 4/2
// (Fallout, #5948):
//
//	"Creature tokens you control have training. (Whenever a creature token
//	 you control attacks with another creature with greater power, put a
//	 +1/+1 counter on that token.)
//	 Blind Betrayal — Sacrifice another creature: Elder Arthur Maxson gains
//	 indestructible until end of turn."
//
// Training is the shared trigger (Training), granted as a layer 6 bundle to
// every creature token its controller controls (Warrior's Resolve's shape),
// so it is each token's own ability and compares that token's power with
// the other attackers' (CR 702.149a). The ability is Bladed Ambassador's
// indestructible with a "sacrifice another creature" cost (SacrificeAnotherN):
// Maxson is never his own payment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d95cd013-6f32-4a8b-9c46-a28695eea4f1",
		Name:         "Elder Arthur Maxson",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       elderArthurMaxsonTrainingGrant,
			Triggered: []game.TriggeredAbility{Training("Training — put a +1/+1 counter on this creature")},
			Text:      "Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsToken() && target.Controller == source.Controller
			}, elderArthurMaxsonTrainingGrant), "Creature tokens you control have training."),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another creature: Elder Arthur Maxson gains indestructible until end of turn.",
			Cost:    SacrificeAnotherN(1, "another creature", Creature()),
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Effect:  thisGainsKeywordUntilEndOfTurn("indestructible", "Elder Arthur Maxson — indestructible until end of turn"),
		}},
	})
}

// elderArthurMaxsonTrainingGrant is the bundle key of the training Maxson
// gives. On disk.
const elderArthurMaxsonTrainingGrant = "elder-arthur-maxson/training"
