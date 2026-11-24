package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Detention Vortex — Enchantment — Aura {W} (Strixhaven: School of Mages,
// #5502):
//
//	"Enchant nonland permanent
//	 Enchanted permanent can't attack or block, and its activated abilities
//	 can't be activated.
//	 {3}: Destroy this Aura. Only your opponents may activate this ability and
//	 only as a sorcery."
//
// Suppression Bonds' restriction on the attached permanent (attacking,
// blocking, and every activated ability, mana abilities included, CR 605.1a),
// with Oft-Nabbed Goat's opponents-only sorcery-speed ability: it is the
// Aura's own ability, so it is not stopped by its own restriction, and the
// activator is any opponent of the Aura's controller, who pays {3} and destroys
// the Aura as it resolves (a source that has left makes it do nothing, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5383a1c1-662f-46fe-a0e5-0f38efe1a37a",
		Name:         "Detention Vortex",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant nonland permanent", effects.Nonland()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
		Activated: []effects.ActivatedAbility{{
			Label:         "{3}: Destroy this Aura. Only your opponents may activate this ability and only as a sorcery.",
			Cost:          effects.ManaCost("{3}"),
			OpponentsOnly: true,
			SorcerySpeed:  true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DestroyTarget{Target: ctx.Source()}.Apply(ctx)
			},
		}},
	})
}
