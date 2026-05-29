package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bitterblossom — Kindred Enchantment — Faerie for {1}{B}:
//
//	"At the beginning of your upkeep, you lose 1 life and create a
//	1/1 black Faerie Rogue creature token with flying."
//
// S19 sub-PR 5: a "your upkeep" trigger combining a life cost with
// a token. Mandatory; Build runs inline. The Faerie token's flying
// is cosmetic (token keywords aren't enforced yet).
func init() {
	Register(Spec{
		OracleID: "fb868840-09fa-49b1-85cb-b08ad065e972",
		Name:     "Bitterblossom",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				ctx := NewContext(g, nil)
				_ = g.ChangePlayerLifeForEffect(source.InstanceID, source.Controller, -1)
				_ = CreateToken{
					Controller: source.Controller,
					Template:   FaerieRogueToken(),
					N:          1,
				}.Apply(ctx)
				return nil
			},
		}},
	})
}
