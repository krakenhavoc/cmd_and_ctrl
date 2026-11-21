package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deluxe Dragster — Artifact — Vehicle {4}{U}, 4/3 (March of the Machine
// Commander):
//
//	"This Vehicle can't be blocked except by Vehicles.
//	 Whenever this Vehicle deals combat damage to a player, you may cast
//	 target instant or sorcery card from that player's graveyard without
//	 paying its mana cost. If that spell would be put into a graveyard,
//	 exile it instead.
//	 Crew 2"
//
// The damage trigger names the damaged player's graveyard in its target
// clause (Fire Nation Salvagers' TargetsFrom), and the cast is Arcane
// Heist's: a free, exile-instead grant offered during the resolution
// (CR 608.2g). A blocker is a creature, so "except by Vehicles" is a
// Vehicle that has been crewed (CR 702.122).
//
// No simplification.
func init() {
	damage := effects.FreeCastFromDamagedPlayersGraveyard("Deluxe Dragster")
	effects.Register(effects.Spec{
		OracleID:     "56ba8fe5-8693-4783-8406-9ced99de0b2a",
		Name:         "Deluxe Dragster",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.OfSubtype("Vehicle"), "Vehicles"),
		},
		Triggered: []game.TriggeredAbility{damage},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Deluxe Dragster"),
		}},
	})
}
