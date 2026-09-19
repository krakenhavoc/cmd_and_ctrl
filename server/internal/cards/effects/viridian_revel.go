package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viridian Revel — Enchantment {1}{G}{G} (EDHREC rank 4033):
//
//	"Whenever an artifact is put into an opponent's graveyard from
//	 the battlefield, you may draw a card."
//
// Green's artifact-hate payoff: it does not remove anything itself,
// it rewards a deck that already does. In a pod full of Signets,
// Treasures and Clues it draws on every cracked Treasure the table
// makes and on every artifact a sweeper takes down.
//
// It is in the batch as the AUDIENCE-NARROWED twin of Disciple of the
// Vault, which watches the same event over the whole table. The
// clause here is "an OPPONENT's graveyard", and the graveyard a
// permanent goes to is its OWNER's (CR 404.3) — not its controller's.
// So a Treasure token you gave an opponent with a Curse effect is
// still yours by ownership and draws nothing, while an opponent's
// artifact you stole with Gilded Drake and then sacrificed goes to
// their graveyard and does draw. That is the printed reading, and it
// is the only place the two ways of writing this condition come apart.
//
// A token counts as long as it is still findable — it exists in the
// graveyard until the state-based sweep removes it (CR 111.7), which
// is after the trigger has already seen it, the same read Agent of
// the Iron Throne makes for a Treasure.
//
// "You may" is a real prompt asked of the Revel's controller, so the
// one case where you would decline — an empty library — is answerable
// rather than lethal.
//
// A sweeper that puts five of an opponent's artifacts into their
// graveyard fires this five times, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c24b0c7d-1a4b-4eca-8d83-9e81ed3b3b27",
		Name:         "Viridian Revel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b38ArtifactPutIntoAnOpponentsGraveyard(ev, source, g)
			}, "Viridian Revel — draw a card", Do(DrawCards{N: 1})),
				"Viridian Revel — draw a card?"),
		},
	})
}
