package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nettling Nuisance — Creature — Faerie Rogue {2}{B}, 3/1:
//
//	"Flying
//	 Whenever one or more Faeries you control deal combat damage to a
//	 player, that player creates a 4/2 red Pirate creature token with
//	 'This token can't block.' The token is goaded for the rest of the
//	 game. (It attacks each combat if able and attacks a player other than
//	 you if able.)"
//
// One trigger per player the Faeries connect with in a damage step (CR
// 603.2c), the Faeries judged on the damage source's effective subtypes.
// The token is the DAMAGED player's, a catalog Pirate that prints its own
// "can't block", and it is goaded for the rest of the game on behalf of
// Nettling Nuisance's controller (game.GoadForRestOfGameForEffect): it must
// attack each combat and may not attack that controller while another
// player is open, for as long as it is on the battlefield. The goad is put
// on as the token lands, even when a token doubler pauses the creation to
// order its replacements.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9662c579-f414-4c2d-b133-0f5f21d5c28a",
		Name:            "Nettling Nuisance",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(HasSubtype("Faerie"),
				"Nettling Nuisance — that player creates a 4/2 Pirate that can't block, goaded for the rest of the game",
				nettlingNuisanceHit),
		},
	})
}

func nettlingNuisanceHit(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	victim := item.Trigger.Event.Target
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	ctx := NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: victim,
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(cmdr2S34PirateToken()), Count: 1}},
	}, ctx.ItemTail(cmdr2S34GoadTokensForRestOfGame))
}
