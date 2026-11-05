package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gwyndala, the Progeny — Legendary Creature — Vau-N'Akat Advisor
// {1}{U}, 1/1 (Star Trek, #4973):
//
//	"When Gwyndala enters, create a colorless Equipment artifact token
//	 named Neuroflux Heirloom with "Equipped creature gets +1/+1" and
//	 equip {2}. Attach it to Gwyndala.
//	 {2}: Each Equipment token you control becomes a copy of target
//	 Equipment you control until end of turn."
//
// The enters trigger is U.S.Agent's (createTokenAttachedToSource): the
// Heirloom is its own token template, and if Gwyndala has left by then
// it stays unattached (CR 701.3b).
//
// The activated ability is a CR 707.2 copy effect with a duration
// (BecomeCopy, CR 611.2): each Equipment TOKEN its controller controls
// as it resolves takes the target's copiable values until end of turn.
// A copy stays a token and stays attached (neither is a copiable value,
// CR 707.2), so an Heirloom on a creature becomes, say, a second
// Bonesplitter on that creature. The target itself, if it is an
// Equipment token, is already a copy of itself and is left alone. A
// target gone by resolution counters the ability (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "852c38fe-8bfa-420b-b961-7b91af9b38d1",
		Name:         "Gwyndala, the Progeny",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gwyndala, the Progeny — create a Neuroflux Heirloom Equipment token and attach it to Gwyndala",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAttachedToSource(g, item, NeurofluxHeirloomEquipmentToken())
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}: Each Equipment token you control becomes a copy of target Equipment you control until end of turn.",
			Cost:    ManaCost("{2}"),
			Targets: TargetPermanent("target Equipment you control", Subtype("Equipment"), YouControl()),
			Effect:  gwyndalaEquipmentTokensBecomeCopies,
		}},
	})
}

// NeurofluxHeirloomEquipmentToken is the Neuroflux Heirloom as PRINTED.
func NeurofluxHeirloomEquipmentToken() game.Card {
	return tokenFromCatalog(printedNeurofluxHeirloomToken)
}

// printedNeurofluxHeirloomToken is the catalog's entry for the Neuroflux
// Heirloom: a colorless Equipment artifact token, "Equipped creature
// gets +1/+1" and equip {2}. Its slug is an on-disk identity and is
// never renamed.
func printedNeurofluxHeirloomToken() tokenTemplate {
	return tokenTemplate{
		Slug: "neuroflux-heirloom-equipment",
		Card: game.Card{
			Name:     "Neuroflux Heirloom",
			TypeLine: "Token Artifact — Equipment",
		},
		Text:      "Equipped creature gets +1/+1.\nEquip {2}",
		Static:    []game.StaticAbility{PumpAttached(1, 1)},
		Activated: activatedShapes([]ActivatedAbility{EquipAbility("{2}")}),
	}
}

func gwyndalaEquipmentTokensBecomeCopies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	model := targets[0].ID
	var tokens []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != model && c.Controller == item.Controller && c.IsToken() && c.HasSubtype("Equipment") {
			tokens = append(tokens, c.InstanceID)
		}
	}
	if len(tokens) == 0 {
		return nil
	}
	return BecomeCopy{
		Targets: tokens,
		Of:      model,
		Label:   "Gwyndala, the Progeny — becomes a copy until end of turn",
	}.Apply(ctx)
}
