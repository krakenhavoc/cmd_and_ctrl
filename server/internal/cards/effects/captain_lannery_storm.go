package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captain Lannery Storm — Legendary Creature — Human Pirate {2}{R},
// 2/2 (issue #1112):
//
//	"Haste
//	 Whenever Captain Lannery Storm attacks, create a Treasure token.
//	 Whenever you sacrifice a Treasure, Captain Lannery Storm gets
//	 +1/+0 until end of turn."
//
// Three printed abilities, three primitives:
//
//   - Haste rides PrintedKeywords; the summoning-sickness check reads
//     it.
//   - The attack trigger is WheneverThisAttacks (#579's named shape)
//     over CreateToken — a plain Treasure, untapped, the S21 fast
//     mana this deck is built around.
//   - The sacrifice trigger watches EventSacrifice directly (Mayhem
//     Devil's shape): "you sacrifice" is ev.Actor == the Captain's
//     own controller, and "a Treasure" is read off the sacrificed
//     card BEFORE it leaves the battlefield — EventSacrifice fires
//     while the permanent is still there (AGENTS.md §7's trigger
//     table), so g.LookupCardForEffect(ev.CardID) still finds it and
//     hasSubtype checks the Treasure subtype rather than the token's
//     name, so a Treasure that is a copy of something else still
//     counts. The pump is BoostUntilEOT pinned to the Captain's own
//     instance, exactly the shape a Giant Growth uses on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "235bf0ba-658c-463f-b112-7478ba27bd7b",
		Name:            "Captain Lannery Storm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Captain Lannery Storm — create a Treasure",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
			{
				Watches: []game.EventKind{game.EventSacrifice},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Actor != source.Controller || ev.CardID == uuid.Nil {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && hasSubtype(c, "Treasure")
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Captain Lannery Storm — +1/+0 until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							return BoostUntilEOT{
								Target: item.SourceCardID,
								Power:  1,
								Label:  "Captain Lannery Storm — +1/+0 until end of turn",
							}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
