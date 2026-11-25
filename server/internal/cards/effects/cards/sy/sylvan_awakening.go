package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvan Awakening — Sorcery {2}{G} (Dominaria, S82 OLD, #5530):
//
//	"Until your next turn, all lands you control become 2/2 Elemental
//	 creatures with reach, indestructible, and haste. They're still lands."
//
// Rude Awakening's land animation, lasting until the caster's next turn begins
// (CR 611.2b) and with the Elemental type, reach, indestructible and haste added
// in their layers. The affected set is the lands the caster controls as the
// spell resolves (CR 611.2c): a land that enters later stays a land. The lands
// are summoning sick only if they came under the caster's control this turn
// (CR 302.6) and haste lifts even that.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66b46257-b64c-4672-bf47-12bf0bcd6f8f",
		Name:         "Sylvan Awakening",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ScopedEffectFor{
				Match: effects.And(effects.Land(), effects.YouControl()),
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.AddSubtypesMod(ctx.TextSubtype("Elemental")),
					game.SetBasePowerMod(2),
					game.SetBaseToughnessMod(2),
					game.AddKeywordsMod("reach", "indestructible", "haste"),
				},
				Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
				Label:    "Sylvan Awakening — lands you control are 2/2 Elemental creatures with reach, indestructible and haste",
			}.Apply(ctx)
		},
	})
}
