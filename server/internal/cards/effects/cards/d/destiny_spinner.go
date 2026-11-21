package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Destiny Spinner — Enchantment Creature — Human {1}{G}, 2/3:
//
//	"Creature and enchantment spells you control can't be countered.
//	 {3}{G}: Target land you control becomes an X/X Elemental creature
//	 with trample and haste until end of turn, where X is the number of
//	 enchantments you control. It's still a land."
//
// The first line is ADR 0106 §4's battlefield static (#1806).
//
// The activated ability is one ScopedEffectFor record pinned to the
// land, Wrenn and Realmbreaker's +1 shape with a counted base P/T. X is
// counted once, as the ability resolves (CR 608.2h), so it does not
// move with the number of enchantments later in the turn; the
// Spinner counts itself while it is on the battlefield. The land keeps
// its types, so it is still a land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7dd189f8-a23f-4b43-b0ef-fb174a9664ba",
		Name:         "Destiny Spinner",
		Completeness: effects.CompletenessFull,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			effects.SpellsYouControlCantBeCountered("Creature and enchantment spells you control can't be countered.", effects.Or(effects.Creature(), effects.Enchantment())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Target land you control becomes an X/X Elemental creature with trample and haste until end of turn, where X is the number of enchantments you control. It's still a land.",
			Cost:    effects.ManaCost("{3}{G}"),
			Targets: effects.TargetPermanent("target land you control", effects.And(effects.Land(), effects.YouControl())),
			Effect:  destinySpinnerAnimateLand,
		}},
	})
}

// destinySpinnerAnimateLand makes the targeted land an X/X Elemental
// with trample and haste until end of turn, X the enchantments the
// activator controls now. A target that left in response does nothing.
func destinySpinnerAnimateLand(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	x := effects.CountControlled(g, ctx.Controller(), func(c game.Card) bool { return c.IsEnchantment() })
	return effects.ScopedEffectFor{
		Target: targets[0].ID,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
			game.SetBasePowerMod(x),
			game.SetBaseToughnessMod(x),
			game.AddKeywordsMod("trample", "haste"),
		},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Destiny Spinner — becomes an X/X Elemental creature until end of turn",
	}.Apply(ctx)
}
