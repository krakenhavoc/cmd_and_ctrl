package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Backdraft Hellkite — Creature — Dragon {3}{R}{R}, 4/4 (Commander 2019,
// #5517):
//
//	"Flying
//	 Whenever this creature attacks, each instant and sorcery card in your
//	 graveyard gains flashback until end of turn. The flashback cost is
//	 equal to its mana cost."
//
// The grant is Past in Flames's (GrantCastFromYourGraveyard): the set is
// locked as the trigger resolves (CR 611.2c), so a card that reaches the
// graveyard later in the turn gains nothing, and each granted card is cast
// for its mana cost under the "flashback" key and exiled whenever it would
// leave the stack (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fcf9057a-4975-4a7d-85f5-3fdc80942fcd",
		Name:            "Backdraft Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Backdraft Hellkite — each instant and sorcery card in your graveyard gains flashback until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantCastFromYourGraveyard{
						Filter:            game.PermissionFilter{InstantOrSorceryOnly: true},
						AltCostKey:        "flashback",
						ExileOnResolution: true,
						Label:             "Flashback — its mana cost (Backdraft Hellkite)",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
