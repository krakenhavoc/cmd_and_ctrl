package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b307_helpers.go — shared shapes for the Battle for Zendikar Devoid
// cards of S82 batch 307 (#5547).

// OldB307Ingest is the ingest keyword (CR 702.115a): "whenever this creature
// deals combat damage to a player, that player exiles the top card of their
// library".
func OldB307Ingest(name string) game.TriggeredAbility {
	return WheneverThisDealsCombatDamageToAPlayer(name+" — ingest: that player exiles the top card of their library",
		func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			return MillToZone{Player: item.Trigger.Event.Target, N: 1, To: game.ZoneExile}.Apply(NewContext(g, item))
		})
}

// OldB307WhenThisEntersCreateScions is "When this creature enters, create N
// 1/1 colorless Eldrazi Scion creature tokens. They have "Sacrifice this
// token: Add {C}."" (the catalog's Eldrazi Scion).
func OldB307WhenThisEntersCreateScions(name string, n int) game.TriggeredAbility {
	what := "a 1/1 colorless Eldrazi Scion"
	if n == 2 {
		what = "two 1/1 colorless Eldrazi Scions"
	}
	return TriggerWithPurpose(WhenThisEnters(name+" — create "+what,
		Do(CreateToken{Template: B28EldraziScionToken(), N: n})),
		game.Purpose{Tokens: n})
}
