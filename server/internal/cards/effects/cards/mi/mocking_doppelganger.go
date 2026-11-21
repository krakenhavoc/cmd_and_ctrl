package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mocking Doppelganger — Creature — Shapeshifter {3}{U}, 0/0 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Flash
//	 You may have this creature enter as a copy of a creature an opponent
//	 controls, except it has "Other creatures with the same name as this
//	 creature are goaded." (They attack each combat if able and attack a player
//	 other than you if able.)"
//
// Mirrorhall Mimic's enters-as-a-copy replacement, with the candidates
// narrowed to creatures an opponent controls, and an exception (CR 707.9) that
// gives the copy a granted static: Koloth's continuous goad (CR 701.15a, a
// static goad has no "until your next turn" end), read for the copy's
// controller on every OTHER creature on the battlefield with the copy's
// current name (a Doppelganger copying a Bear is a Bear, so other Bears are
// goaded, whoever controls them). Declined, it is a 0/0 that dies as a state
// check (CR 704.5f).
//
// No simplification.
func init() {
	const grant = "mocking-doppelganger/goad-same-name"
	effects.Register(effects.Spec{
		OracleID:        "3b8eb602-01e8-432e-8dfc-8fdb62c3aa43",
		Name:            "Mocking Doppelganger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Static: []game.StaticAbility{{
				Label: "Other creatures with the same name as this creature are goaded.",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.InstanceID != source.InstanceID && target.Name == source.Name
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.AttackRequirements = append(c.AttackRequirements, game.GoadRequirements(source, source.Controller)...)
				},
			}},
			Text: "Other creatures with the same name as this creature are goaded.",
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Mocking Doppelganger",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					var out []uuid.UUID
					for _, id := range effects.AnyCreatureOnBattlefield(g, controller, self) {
						if c, ok := g.LookupCardForEffect(id); ok && c.Controller != controller {
							out = append(out, id)
						}
					}
					return out
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.GrantAbility(grant)
				}),
		},
	})
}
