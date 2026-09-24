package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gogo, Master of Mimicry — Legendary Creature — Wizard {2}{U}, 2/4
// (Edea steal-and-sac deck, #1565):
//
//	"{X}{X}, {T}: Copy target activated or triggered ability you
//	 control X times. You may choose new targets for the copies. This
//	 ability can't be copied and X can't be 0. (Mana abilities can't
//	 be targeted.)"
//
// Strionic Resonator's ability-copy seam (#1223), X times. The target
// clause is AbilityOnStack narrowed to "you control" and nothing else,
// because Gogo copies activated AND triggered abilities (Lithoform
// Engine's reach, not the Resonator's). A mana ability never uses the
// stack (CR 605.3b), so the reminder text is enforced by the ability
// not being there to click.
//
// The cost is ManaCost("{X}{X}") with MinX(1): one announced X pays
// both {X}s (CR 107.3), so X = 2 costs four mana and makes two
// copies, and X = 0 is refused at announce, as printed. CopyAbility's
// Count is X, and each copy gets its own CR 707.10c re-target offer.
//
// "This ability can't be copied", half of it: Gogo's own clause will
// not target a Gogo activation — its own, or one from a copy of Gogo
// (the check is on the ability source's catalog key, which a Clone or
// token copy carries). The other half, another card's copy effect
// (Lithoform Engine, Rings of Brighthearth) copying a Gogo activation,
// needs a "this ability can't be copied" flag the engine's ability
// copy does not have, and ships as the declared caveat.
func init() {
	Register(Spec{
		OracleID:     gogoMasterOfMimicryOracleID,
		Name:         "Gogo, Master of Mimicry",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Another card that copies abilities, such as Lithoform Engine or Rings of Brighthearth, can still copy Gogo's ability.",
		},
		Activated: []ActivatedAbility{{
			Label: "{X}{X}, {T}: Copy target activated or triggered ability you control X times. You may choose new targets for the copies. This ability can't be copied and X can't be 0.",
			Cost:  Plus(ManaCost("{X}{X}"), TapCost(), MinX(1)),
			Targets: AbilityOnStack("target activated or triggered ability you control",
				AnAbilityYouControl(), notAGogoActivation()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok || ctx.X() < 1 {
					return nil
				}
				return CopyAbility{
					ItemID:           id,
					Controller:       item.Controller,
					Count:            ctx.X(),
					ChooseNewTargets: true,
				}.Apply(ctx)
			},
		}},
	})
}

const gogoMasterOfMimicryOracleID = "61586052-7d69-489c-84c4-0359228d131b"

// notAGogoActivation keeps "this ability can't be copied" inside
// Gogo's own clause: an activated ability whose source is a Gogo (or a
// copy of one) is not a legal target.
func notAGogoActivation() AbilityPredicate {
	return func(g *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		if item.Kind != game.StackItemActivated {
			return true
		}
		src, ok := g.LookupCardForEffect(item.SourceCardID)
		return !ok || game.BaseCatalogKey(game.CatalogKey(src)) != gogoMasterOfMimicryOracleID
	}
}
