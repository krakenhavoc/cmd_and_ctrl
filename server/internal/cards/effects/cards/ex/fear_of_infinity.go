package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fear of Infinity — Enchantment Creature — Nightmare {1}{U}{B}, 2/2:
//
//	"Flying, lifelink
//	 This creature can't block.
//	 Eerie — Whenever an enchantment you control enters and whenever
//	 you fully unlock a Room, you may return this card from your
//	 graveyard to your hand."
//
// Flying and lifelink ride PrintedKeywords, and "can't block" is the
// Bloodghast restriction. The eerie ability watches from the GRAVEYARD
// (InGraveyard, #925), the only zone it is ever useful from; "you" and
// "your graveyard" read the card's owner there (CR 108.4). The "you may"
// is a CR 603.5 prompt, so declining leaves the card where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "018327d1-a3ba-4912-9c88-f0f0c54a1750",
		Name:            "Fear of Infinity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InGraveyard(effects.Eerie("Fear of Infinity — return it from your graveyard to your hand (eerie)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(effects.NewContext(g, item))
				})), "Fear of Infinity — return it to your hand?"),
		},
	})
}
