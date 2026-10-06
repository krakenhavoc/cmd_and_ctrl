package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wishmonger — Creature — Unicorn Monger {3}{W}, 3/3:
//
//	"{2}: Target creature gains protection from the color of its
//	 controller's choice until end of turn. Any player may activate
//	 this ability."
//
// ADR 0106 PR 6 (#1793). Mother of Runes' body with two players in it
// instead of one:
//
//   - The any-player row (CR 602.2, 602.1b): the activator pays the {2}
//     and chooses the target (CR 602.1a, 602.2b).
//   - The COLOUR is chosen by the target creature's controller, as the
//     ability resolves — "its controller's choice" — which need not be
//     the activator. The controller is read off the creature then, so
//     one that changed hands in response is chosen for by its new
//     controller. Five colours, as CR 105.4 says; the prompt declares
//     ColorForProtection (#780).
//   - The grant is protection from that colour until end of turn
//     (CR 702.16a, 514.2), through the same token Mother of Runes
//     grants (game.ProtectionFromColor). A target gone by resolution is
//     not asked about (CR 608.2b).
//
// No purpose for the bot: what the protection is worth depends on the
// colour its controller then picks, which no Purpose field
// can say in advance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f68b90af-9502-4d99-aadf-2bfcea8a655a",
		Name:         "Wishmonger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{2}: Target creature gains protection from the color of its controller's choice until end of turn. Any player may activate this ability.",
			Cost:      game.AbilityCost{Mana: "{2}"},
			Targets:   TargetCreature("target creature"),
			AnyPlayer: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					chooser, ok := controllerOfTarget(ctx, t.ID)
					if !ok {
						return nil
					}
					target := t.ID
					ChooseColorThen(game.ColorForProtection, g, chooser, item.SourceCardID,
						"Wishmonger — choose a color for your creature's protection",
						func(g *game.Game, color string) error {
							token := game.ProtectionFromColor(color)
							if token == "" {
								return nil
							}
							return GrantKeywordUntilEOT{
								Target:   target,
								Keywords: []string{token},
								Label:    "Wishmonger — " + token,
							}.Apply(NewContext(g, item))
						})
					return nil
				}
				return nil
			},
		}},
	})
}
