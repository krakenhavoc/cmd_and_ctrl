package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thada Adel, Acquisitor — Legendary Creature — Merfolk Rogue {1}{U}{U},
// 2/2 (Dominaria United Commander reprint, #5908):
//
//	"Islandwalk
//	 Whenever Thada Adel deals combat damage to a player, search that
//	 player's library for an artifact card and exile it. Then that player
//	 shuffles. Until end of turn, you may play that card."
//
// Islandwalk is a printed keyword (CR 702.14). You search the damaged
// player's library (Bribery's LibraryOwner), exile the card, and that
// player shuffles it. The card stays in exile and its owner stays the
// opponent; you are granted permission to play it (CR 601, or CR 305 for an
// artifact land) until end of turn, paying its costs. The search may find
// nothing (CR 701.23b); the permission is then granted to nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9a84fbb-0972-4a3d-bdae-a3b01219b35d",
		Name:            "Thada Adel, Acquisitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Thada Adel — search that player's library for an artifact card, exile it, you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					victim := item.Trigger.Event.Target
					ctx := NewContext(g, item)
					return SearchLibrary{
						Player:       ctx.Controller(),
						LibraryOwner: victim,
						Predicate:    func(c game.Card) bool { return c.IsArtifact() },
						Dest:         game.ZoneExile,
						Limit:        1,
						Shuffle:      true,
						Reason:       "Thada Adel — search that player's library for an artifact card",
						ThenTail:     ctx.Tail(thadaAdelPlayIt),
					}.Apply(ctx)
				}),
		},
	})
}

// thadaAdelPlayIt is "until end of turn, you may play that card" over the
// card the search exiled. An on-disk key.
var thadaAdelPlayIt = CardTail("thada-adel/play-the-exiled-card", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     ctx.Controller(),
			Zone:       game.ZoneExile,
			Duration:   ctx.Game.UntilEndOfTurnDuration(),
			Source:     ctx.Source(),
			SourceName: "Thada Adel, Acquisitor",
			Label:      "Thada Adel — play the exiled artifact this turn",
		}, []game.Card{c})
	}
	return nil
})
