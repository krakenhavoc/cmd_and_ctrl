package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Endless Foot Assault — Enchantment {2}{W} (TMNT Commander, #5011):
//
//	"Squad {1}{W} (As an additional cost to cast this spell, you may pay
//	 {1}{W} any number of times. When this enchantment enters, create
//	 that many tokens that are copies of it.)
//	 Whenever you attack, for each opponent, create a 1/1 black Ninja
//	 creature token that's tapped and attacking that player."
//
// Squad is CR 702.157a's two linked abilities: an optional mana cost
// paid any number of times (SquadMana), and an enters trigger that makes
// a token copy for each payment (Squad; both in tmc_b4_helpers.go). The
// cap of 20 payments on one announcement is the engine's, multikicker's
// reason. The copies have both of
// this card's abilities, so each one makes its own Ninjas.
//
// "Whenever you attack" is CR 508.3d: it triggers once when one or more
// creatures you control are declared as attackers, at a player, a
// planeswalker or a battle (OncePerBatch over the attack events of one
// declaration). As it resolves, every opponent still in the game gets a
// Ninja attacking them, attacked or not (the card's ruling). The Ninjas
// were never declared as attackers (CR 508.4), so they trigger nothing
// that reads a creature attacking, this card included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77ccafd9-6d76-4658-81ab-bde0454fcc15",
		Name:         "Endless Foot Assault",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			SquadMana("{1}{W}", 20),
		},
		Triggered: []game.TriggeredAbility{
			Squad("Endless Foot Assault"),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Endless Foot Assault — a tapped and attacking Ninja for each opponent", endlessFootAssaultNinjas)),
		},
	})
}

func endlessFootAssaultNinjas(g *game.Game, item *game.StackItem) error {
	return tokenTappedAndAttackingEachOpponent(NewContext(g, item), func() game.Card {
		return TokenCard("1/1 black Ninja")
	})
}
