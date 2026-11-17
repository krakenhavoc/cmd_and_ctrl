package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Repository Skaab — Creature — Zombie {3}{U}, 3/3:
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, return target instant or
//	 sorcery card from your graveyard to your hand."
//
// Exploit is the shared constructor (exploit.go); the payoff is a
// reflexive trigger whose target is chosen once a creature has really
// been sacrificed, among the instant and sorcery cards in its
// controller's graveyard then. A target that has left the graveyard by
// resolution is not returned (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7932ddef-b274-4f56-be7f-72e82d17fc59",
		Name:         "Repository Skaab",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Exploit("Repository Skaab", repositorySkaabReturnBody),
		},
	})
}

// repositorySkaabReturnBody is "When this creature exploits a creature,
// return target instant or sorcery card from your graveyard to your
// hand." An on-disk key: never renamed, never reused.
var repositorySkaabReturnBody = game.ReflexiveBody("repository-skaab/return", simpleBody(returnTargetCardToHand),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target instant or sorcery card from your graveyard", YouOwn(), Or(Instant(), Sorcery()))
	}))
