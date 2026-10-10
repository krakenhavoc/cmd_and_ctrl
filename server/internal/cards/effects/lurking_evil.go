package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lurking Evil — {B}{B}{B} Enchantment:
//
//	"Pay half your life, rounded up: This enchantment becomes a 4/4
//	 Phyrexian Horror creature with flying."
//
// The life is a computed cost (#1594, ADR 0020 Decision 47), read at
// announce (CR 601.2f–g via CR 602.2b): half of what the activator has
// then, rounded up.
//
// The effect has no printed duration, so it lasts as long as the
// permanent is this object (CR 611.2a) and is pinned to it (CR 611.2c
// / 400.7 — a flickered Lurking Evil comes back an enchantment). It is
// one record at one timestamp, a data ScopedEffect (ADR 0041 phase 3),
// so a table holding it is still a restore point.
//
// "Becomes a … creature" with no "in addition to its other types" and
// no "artifact creature" SETS the card type (CR 205.1a / 205.1b): the
// animated permanent is a Phyrexian Horror creature and no longer an
// enchantment. Its colour is untouched — it stays black.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81d6b42e-9c41-4d1e-bd28-4662e1fde572",
		Name:         "Lurking Evil",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Pay half your life, rounded up: This enchantment becomes a 4/4 Phyrexian Horror creature with flying",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    PayLifeCount(LifeHalfYoursRoundedUp),
			Effect:  lurkingEvilBecomesAHorror,
		}},
	})
}

// lurkingEvilBecomesAHorror is the ability's effect. A package-level
// func reading the source off the item, so it captures nothing.
func lurkingEvilBecomesAHorror(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self := ctx.Source()
	if !g.Battlefield.Contains(self) {
		return nil
	}
	return ScopedEffectFor{
		Target: self,
		Mods: []game.Mod{
			game.RemoveTypesMod("Enchantment"),
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod("Phyrexian", "Horror"),
			game.SetBasePowerMod(4),
			game.SetBaseToughnessMod(4),
			game.AddKeywordsMod("flying"),
		},
		Duration: g.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Lurking Evil — becomes a 4/4 Phyrexian Horror creature with flying",
	}.Apply(ctx)
}
