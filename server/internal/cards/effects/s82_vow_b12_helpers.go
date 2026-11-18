package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vowB12TapUpToOneTarget is "tap up to one target creature" as an
// ability's whole effect (Timin, Youthful Geist; Breath of the Sleepless):
// zero targets taps nothing, and a target gone by resolution is skipped
// (CR 608.2b).
func vowB12TapUpToOneTarget(ctx *Context) error {
	return TapTarget{Target: firstLegalCardTarget(ctx)}.Apply(ctx)
}

// vowB12UpToOneCreature is the target clause of both tap abilities.
func vowB12UpToOneCreature() *game.TargetSpec {
	return TargetCreature("up to one target creature").WithCount(0, 1)
}

// vowB12SpiritToken is the 1/1 white Spirit with flying every card in the
// batch makes.
func vowB12SpiritToken() Applier {
	return CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1}
}

// vowB12PreventNoncombatToSpirits is "Prevent all noncombat damage that
// would be dealt to Spirits you control" (Drogskol Reinforcements): the
// plain prevention over creatures you control, narrowed to the ones that
// are Spirits as the damage would be dealt (layered subtypes, so a
// changeling counts).
func vowB12PreventNoncombatToSpirits(label string) game.ReplacementEffect {
	r := PreventAllDamageDealtTo(PreventionStatic{
		To:     ToCreaturesYouControl,
		Damage: NoncombatDamage,
		Label:  label,
	})
	inner := r.AppliesTo
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !inner(ev, g, src) {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.DamageTarget)
		return ok && c.HasSubtypeWord("Spirit")
	}
	return r
}

// vowB12ControlsAnEnchantment is "Activate only if you control an
// enchantment" (Heron-Blessed Geist).
func vowB12ControlsAnEnchantment() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return countControlled(g, controller, game.Card.IsEnchantment) > 0
	}
}

// vowB12OpponentsWithMoreLands counts the opponents who control more
// lands than `controller` (Priest of the Blessed Graf).
func vowB12OpponentsWithMoreLands(g *game.Game, controller uuid.UUID) int {
	mine := b03LandsControlled(g, controller)
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		if b03LandsControlled(g, p.ID) > mine {
			n++
		}
	}
	return n
}

// vowB12YouCastACreatureSpellOnAnOpponentsTurn is Breath of the Sleepless's
// condition: a creature spell cast by the source's controller while it is
// another player's turn.
func vowB12YouCastACreatureSpellOnAnOpponentsTurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !YouCast(Creature())(ev, source, lki, g) {
		return false
	}
	a := g.Turn.ActiveSeat
	return a >= 0 && a < len(g.Seats) && g.Seats[a] != nil && g.Seats[a].ID != source.Controller
}
