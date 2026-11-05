package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scotty, Miracle Worker — Legendary Creature — Human Artificer
// {1}{U}{U}, 2/3 (Star Trek, #4973):
//
//	"During your main phase, each card in your graveyard has beam me up.
//	 The beam me up cost is equal to that spell's mana cost. Scotty can't
//	 be beamed up to cast a spell this way. (Only instant and sorcery
//	 spells can be cast this way. You may cast a card from your graveyard
//	 for its beam me up cost if you also return another creature you
//	 control to its owner's hand. Then exile that spell.)"
//
// One standing permission over its controller's graveyard, gated to that
// player's main phases (CardsInYourGraveyardHaveBeamMeUpDuringYourMainPhase,
// trek_b8_helpers.go). Its offer, game.AltCostKeyGrantedBeamMeUp, charges
// the card's own mana cost, returns a creature other than Scotty and
// exiles the spell as it leaves the stack. The grant gives no timing:
// a sorcery still waits for an empty stack, and an instant may be cast
// with something on the stack, as long as it is your main phase. A card
// that prints beam me up keeps its own offer beside this one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e5057ae-165c-4fe4-b712-a638b7eea038",
		Name:         "Scotty, Miracle Worker",
		Completeness: CompletenessFull,
		GatedCastPermissions: []game.CastPermissionGate{
			CardsInYourGraveyardHaveBeamMeUpDuringYourMainPhase("Scotty, Miracle Worker"),
		},
	})
}
