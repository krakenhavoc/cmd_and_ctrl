package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tyrannical Pitlord — Creature — Demon {4}{B}{B}, 6/6 (Dominaria United):
//
//	"Flying, trample
//	 As this creature enters, choose another creature you control.
//	 The chosen creature gets +3/+3 and has flying.
//	 When this creature leaves the battlefield, sacrifice the chosen creature."
//
// The choice is made as it enters (CR 614.12) and is mandatory when there is
// another creature to pick: it is stored as the Pitlord's linked permanent, an
// object with its CR 400.7 epoch (game.QueueChooseLinkedPermanentAsEnters…),
// so a chosen creature that leaves and returns is a new object nobody chose.
// With no other creature, nothing is chosen. The two statics (a layer 7c +3/+3
// and a layer 6 flying grant) apply to that object only, for as long as the
// Pitlord is on the battlefield.
//
// The leave trigger reads the choice from the Pitlord as it last existed
// (CR 608.2h, PermanentInfo.LinkedPermanent), so it still names the creature
// when the Pitlord is the one that died. Only a permanent its controller
// controls can be sacrificed (CR 701.21a): a creature that has changed
// controllers since is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "776ea0c7-fc5b-44b4-9302-5a3514ff843b",
		Name:            "Tyrannical Pitlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		AsEnters:        tyrannicalPitlordChoose,
		Static: []game.StaticAbility{
			{
				Label:     "The chosen creature gets +3/+3.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: tyrannicalPitlordChosen,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 3
					c.Toughness += 3
				},
			},
			func() game.StaticAbility {
				ab := KeywordGrant(tyrannicalPitlordChosen, "flying")
				ab.Label = "The chosen creature has flying."
				return ab
			}(),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Tyrannical Pitlord — sacrifice the chosen creature", tyrannicalPitlordSacrifice),
		},
	})
}

// tyrannicalPitlordChoose is "As this creature enters, choose another creature
// you control."
func tyrannicalPitlordChoose(card *game.Card, ctx *Context) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID != card.InstanceID && c.Controller == card.Controller && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	ctx.Game.QueueChooseLinkedPermanentAsEntersForEffect(card.Controller, card.InstanceID,
		"Tyrannical Pitlord — choose another creature you control", ids)
	return nil
}

// tyrannicalPitlordChosen is "the chosen creature": the object the Pitlord's
// linked permanent names, still on the battlefield as that object.
func tyrannicalPitlordChosen(target *game.Card, _ *game.Game, source *game.Card) bool {
	l := source.LinkedPermanent
	return l.ID != uuid.Nil && l.ID == target.InstanceID && l.Epoch == target.ObjectEpoch
}

func tyrannicalPitlordSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	chosen := g.LinkedPermanentForEffect(ref)
	if chosen == nil {
		return nil
	}
	// The controller as the layers have it now (a control change is a layer 2 effect).
	if live, ok := g.LookupCardForEffect(chosen.InstanceID); !ok || live.Controller != ctx.Controller() {
		return nil
	}
	return SacrificePermanent{Target: chosen.InstanceID}.Apply(ctx)
}
