package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magitek Scythe — Artifact — Equipment {4} (Final Fantasy, #4462):
//
//	"A Test of Your Reflexes! — When this Equipment enters, you may
//	 attach it to target creature you control. If you do, that creature
//	 gains first strike until end of turn and must be blocked this turn
//	 if able.
//	 Equipped creature gets +2/+1.
//	 Equip {2}"
//
// Barbed Bloodletter's enters-attach made a "you may" (CR 603.5, asked
// as the trigger resolves; the target is chosen as it goes on the
// stack). "If you do" is read off the board after the attach: only when
// the Scythe is now attached to that creature does it gain first strike
// and the must-be-blocked requirement (Emergent Growth's
// BlockRequirementUntilEOT), both until end of turn. A Scythe that left
// in response, or a "no", gives the creature nothing.
//
// No simplification.
func init() {
	const name = "Magitek Scythe"
	effects.Register(effects.Spec{
		OracleID:     "0d5bbde7-6f33-4708-b3ab-34d4528af649",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("A Test of Your Reflexes! — you may attach it to target creature you control; it gains first strike and must be blocked this turn",
					magitekScytheAttach),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			), name+" — attach it to that creature?"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// magitekScytheAttach attaches the Scythe to the still-legal target and,
// if it is now attached there, gives that creature first strike and
// "must be blocked this turn if able".
func magitekScytheAttach(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := effects.AttachSourceToTarget(g, item); err != nil {
		return err
	}
	self, found := g.LookupCardForEffect(item.SourceCardID)
	if !found || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) || !self.IsAttachedTo(id) {
		return nil
	}
	if err := (effects.GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"first strike"},
		Label:    "Magitek Scythe — first strike until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.BlockRequirementUntilEOT{
		Target: id,
		Kind:   game.BlockRequirementMustBeBlocked,
		Label:  "Magitek Scythe — must be blocked this turn",
	}.Apply(ctx)
}
