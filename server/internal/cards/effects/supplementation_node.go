package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Supplementation Node — Artifact Creature — Borg {2}{W}, 2/2 (Star Trek
// Commander, #4974):
//
//	"Artifact creatures you control have "Whenever this creature deals
//	 combat damage to a player, put a +1/+1 counter on it.""
//
// Hive Mind Coprocessor's shape: a layer-6 grant of a catalog bundle (ADR
// 0093), so the trigger is each recipient's own and "it" is the artifact
// creature that connected. The Node is an artifact creature, so it has the
// ability too. Two Nodes grant two instances, which trigger separately.
//
// No simplification.
const supplementationNodeGrant = "supplementation-node/combat-plus-one-counter"

func init() {
	const text = "Whenever this creature deals combat damage to a player, put a +1/+1 counter on it."
	Register(Spec{
		OracleID:     "24296042-2963-405b-88d7-28808280ac00",
		Name:         "Supplementation Node",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: supplementationNodeGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Supplementation Node — put a +1/+1 counter on this creature",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
		}, supplementationNodeGrant), "Artifact creatures you control have \""+text+"\"")},
	})
}
