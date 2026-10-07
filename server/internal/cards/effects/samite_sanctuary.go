package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samite Sanctuary — Enchantment {2}{W}:
//
//	"{2}: Prevent the next 1 damage that would be dealt to target
//	 creature this turn. Any player may activate this ability."
//
// ADR 0106 PR 6 (#1793). The whole card is one any-player row
// (CR 602.2, 602.1b) over Mending Hands' charged shield
// (PreventNextDamage, CR 615.7) with a charge of one. The {2} is the
// activator's to pay and the target the activator's to choose
// (CR 602.1a, 602.2b); a target gone by resolution leaves nothing to
// shield (CR 608.2b). The shield is pinned to that creature as an
// object and ends at cleanup (CR 514.2).
//
// No purpose for the bot: it shields a creature for whoever chose it,
// which no Purpose field describes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e6e1ea6-4682-4056-b70e-fe45d0c9fdee",
		Name:         "Samite Sanctuary",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{2}: Prevent the next 1 damage that would be dealt to target creature this turn. Any player may activate this ability.",
			Cost:      game.AbilityCost{Mana: "{2}"},
			Targets:   TargetCreature("target creature"),
			AnyPlayer: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return PreventNextDamage{
						Target: t.ID,
						Amount: 1,
						Label:  "Samite Sanctuary: prevent the next 1 damage",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
