package effects

import (
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Exchange — Sorcery {B}{B} (#3402, deck request #3209):
//
//	"As an additional cost to cast this spell, exile a creature you
//	 control. Return target creature card from your graveyard to the
//	 battlefield. Put a +2/+2 counter on that creature if the exiled
//	 creature was a Thrull."
//
// The additional cost is the activated ability's exile-a-permanent
// component on a cast (ExileCreatureYouControlCost,
// AdditionalCost.ExilePermanents): the caster names the creature on
// cast_spell as exile_permanent_ids, it is exiled with the spell on the
// stack (nothing dies, nothing is sacrificed), and it is recorded on the
// payment as the battlefield object it was. "If the exiled creature was
// a Thrull" is CR 608.2h's last-known information, read through
// ctx.ExiledPermanents(): a Thrull that was a Thrull when it left still
// counts, and a changeling is every creature type.
//
// The creature is exiled while the spell is cast, so it is never in the
// graveyard to be the spell's own target. The counter goes on "that
// creature" only if it actually came back (a replacement can keep it
// off the battlefield), and the reanimation is the family's body
// (reanimateSingleTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "366ea3b4-5e7e-4693-9f4b-1e549972609e",
		Name:           "Soul Exchange",
		Completeness:   CompletenessFull,
		AdditionalCost: ExileCreatureYouControlCost(),
		Targets:        TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
		Purpose:        ForTargets(Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			wasThrull := false
			// "Thrull" is a word of the spell's text (CR 612.2, #3984).
			thrull := ctx.TextSubtype("Thrull")
			for _, info := range ctx.ExiledPermanents() {
				if permanentWasCreatureType(info, thrull) {
					wasThrull = true
				}
			}
			targets := ctx.Targets()
			if _, ok := reanimateSingleTarget(ctx, ctx.Controller()); !ok || !wasThrull || len(targets) == 0 {
				return nil
			}
			id := targets[0].ID
			if !slices.ContainsFunc(ctx.Game.BattlefieldCardsForEffect(), func(c game.Card) bool { return c.InstanceID == id }) {
				return nil
			}
			return AddCounter{Target: id, Kind: "+2/+2", N: 1}.Apply(ctx)
		},
	})
}

// permanentWasCreatureType reports whether `info` was a creature of the
// given type as it last existed on the battlefield: the type is among its
// subtypes (case-insensitively), or it had every creature type
// (changeling).
func permanentWasCreatureType(info game.PermanentInfo, subtype string) bool {
	if info.Characteristic.AllCreatureTypes {
		return true
	}
	return slices.ContainsFunc(info.Characteristic.Subtypes, func(s string) bool {
		return strings.EqualFold(s, subtype)
	})
}
