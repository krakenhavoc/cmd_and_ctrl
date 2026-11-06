package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foggy Swamp Spirit Keeper — Creature — Human Druid Ally {1}{U}{B}, 2/4
// (Avatar: The Last Airbender, #4459):
//
//	"Lifelink
//	 Whenever you draw your second card each turn, create a 1/1 colorless
//	 Spirit creature token with "This token can't block or be blocked by
//	 non-Spirit creatures.""
//
// The trigger is Knights of Dol Amroth's (b41DrewYourSecondCardThisTurn):
// it fires on the second draw only, whether or not the Keeper saw the
// first, and once a turn (the rulings). The Spirit is Avatar Kuruk's
// printed ttla token, its block rule included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "faecfb8d-2f63-4b98-a842-b11448f0d538",
		Name:            "Foggy Swamp Spirit Keeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Foggy Swamp Spirit Keeper — create a 1/1 colorless Spirit",
				Do(CreateToken{Template: kurukSpiritToken(), N: 1})),
		},
	})
}
