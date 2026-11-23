package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Odric, Blood-Cursed — Legendary Creature — Vampire Soldier {1}{R}{W}, 3/3
// (Innistrad: Crimson Vow, #5495):
//
//	"When Odric enters, create X Blood tokens, where X is the number of
//	 abilities from among flying, first strike, double strike, deathtouch,
//	 haste, hexproof, indestructible, lifelink, menace, reach, trample, and
//	 vigilance found among creatures you control. (Count each ability only
//	 once.)"
//
// X is read as the trigger resolves: each of the twelve listed keywords
// counts once if any creature the controller controls has it as the
// layers now stand (so a granted keyword counts, and a keyword two
// creatures share counts once). Odric's own abilities are included, as he is
// on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "690f79f3-e6a5-4542-8401-61a9bf645d55",
		Name:         "Odric, Blood-Cursed",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Odric, Blood-Cursed — create X Blood tokens, X the number of listed abilities among creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.VowB05ListedAbilitiesAmongCreatures(g, item.Controller)
					if x <= 0 {
						return nil
					}
					return effects.CreateToken{Controller: item.Controller, Template: effects.BloodToken(), N: x}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
