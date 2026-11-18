package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vrondiss's 5/4 red and green Dragon Spirit creature token: "When this
// token deals damage, sacrifice it." The ability is printed on the token,
// so it is a catalog template (ADR 0083), not a plain table row. One
// trigger per damage event: a second event after the sacrifice finds it
// gone and does nothing.
func afrDragonSpiritToken() game.Card { return tokenFromCatalog(printedAfrDragonSpiritToken) }

func printedAfrDragonSpiritToken() tokenTemplate {
	return tokenTemplate{
		Slug: "afr-dragon-spirit-5-4",
		Card: game.Card{
			Name:      "Dragon Spirit",
			TypeLine:  "Token Creature — Dragon Spirit",
			Power:     5,
			Toughness: 4,
			Colors:    []string{"R", "G"},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventDealDamage && ev.Source == source.InstanceID && ev.Amount > 0
			}, "Dragon Spirit — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		Text: "When this token deals damage, sacrifice it.",
	}
}
