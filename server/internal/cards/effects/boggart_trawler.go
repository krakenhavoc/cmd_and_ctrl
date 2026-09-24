package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boggart Trawler // Boggart Bog — the FRONT face, Creature — Goblin
// {2}{B}, 3/1:
//
//	"When this creature enters, exile target player's graveyard."
//
// The land back (pay 3 life or enter tapped; {T}: Add {B}) is the
// mdfc_lands.go row under "<oracle>#1"; this is face 0, which keeps
// the bare oracle ID (game.CatalogKey).
//
// Bojuka Bog's shape exactly — the same "target player" clause on a
// creature instead of a land, so exileTargetPlayersGraveyard is
// reused whole. Mandatory: there is always a legal player (including
// its own controller), so it never fizzles for want of a target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "727f3201-1cfc-4ab2-9dfe-be4f7251f42f",
		Name:         "Boggart Trawler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPlayer("target player"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Boggart Trawler — exile target player's graveyard",
					exileTargetPlayersGraveyard)
			},
		}},
	})
}
