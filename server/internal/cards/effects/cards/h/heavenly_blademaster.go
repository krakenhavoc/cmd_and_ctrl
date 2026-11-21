package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heavenly Blademaster — Creature — Angel {5}{W}, 3/6 (Commander 2018,
// #5526):
//
//	"Flying, double strike
//	 When this creature enters, you may attach any number of Auras and
//	 Equipment you control to it.
//	 Other creatures you control get +1/+1 for each Aura and Equipment
//	 attached to this creature."
//
// The enters trigger is Ardenn's attach with the Blademaster as the host:
// as it resolves, the controller picks any number of the Auras and
// Equipment they control that could legally be attached to it (CR 701.3b:
// an Aura only where its enchant clause and the host's protection allow),
// none being the "you may", and each picked one is attached to it through
// Ardenn's keyed pick (ardennAttachPick). One already on the Blademaster is
// not offered. A Blademaster that has left, or is a new object, gets
// nothing.
//
// The anthem is a layer 7c modification of the other creatures you control
// whose size is the Auras and Equipment attached to the Blademaster now,
// counted on every recompute (CR 611.3a), whoever controls the attachments.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "56940b6c-90ab-491b-a88a-4cd4a8adfbdb",
		Name:            "Heavenly Blademaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Heavenly Blademaster — you may attach any number of Auras and Equipment you control to it",
				heavenlyBlademasterAttach),
		},
		Static: []game.StaticAbility{{
			Label:    "Other creatures you control get +1/+1 for each Aura and Equipment attached to this creature.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, a := range g.BattlefieldCardsForEffect() {
					if a.IsAttachedTo(source.InstanceID) && (a.IsAura() || a.HasSubtype("Equipment")) {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}

// heavenlyBlademasterAttach offers the Auras and Equipment you control that
// could go on the Blademaster and attaches the ones picked.
func heavenlyBlademasterAttach(g *game.Game, item *game.StackItem) error {
	host := item.SourceCardID
	if !effects.OnBattlefield(g, host) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || c.InstanceID == host || c.IsAttachedTo(host) {
			continue
		}
		if !c.HasSubtype("Aura") && !c.HasSubtype("Equipment") {
			continue
		}
		if !g.AttachLegalForEffect(c.InstanceID, host) {
			continue
		}
		candidates = append(candidates, c.InstanceID)
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   item.Controller,
		Source:    host,
		Question:  "Heavenly Blademaster — attach any number of Auras and Equipment you control to Heavenly Blademaster",
		Cards:     candidates,
		Min:       0,
		Max:       len(candidates),
		ThenKey:   effects.ArdennAttachPick,
		LiveZone:  game.ZoneBattlefield,
		Carry:     []uuid.UUID{host},
		CarryWord: "permanent",
	})
	return nil
}
