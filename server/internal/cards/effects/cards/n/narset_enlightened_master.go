package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Narset, Enlightened Master — Legendary Creature — Human Monk
// {3}{U}{R}{W}, 3/2:
//
//	"First strike, hexproof
//	 Whenever Narset attacks, exile the top four cards of your library.
//	 Until end of turn, you may cast noncreature spells from among those
//	 cards without paying their mana costs."
//
// The attack trigger is ExileTopWithPermission: the top four cards of
// its controller's library go to exile face up, and each carries a
// cast-only, until-end-of-turn permission whose cost is {0} (so an X is
// 0, CR 107.3b) and whose filter refuses a creature card
// (PermissionFilter.NoncreatureOnly). The casts keep their normal timing
// (CR 601.3), so a sorcery waits for the second main phase. A land
// among the four stays in exile, since playing a land is not casting a
// spell. A card not cast this turn stays in exile.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aeaf7768-ed67-487d-8506-e667babb88ba",
		Name:            "Narset, Enlightened Master",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Narset, Enlightened Master — exile the top four cards of your library; cast noncreature spells from among them free this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{
						From:     item.Controller,
						GrantTo:  item.Controller,
						N:        4,
						CastOnly: true,
						Free:     true,
						Filter:   game.PermissionFilter{NoncreatureOnly: true},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
