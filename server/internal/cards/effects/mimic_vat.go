package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mimic Vat — Artifact {3}:
//
//	"Imprint — Whenever a nontoken creature dies, you may exile that
//	 card. If you do, return each other card exiled with this artifact
//	 to its owner's graveyard.
//	 {3}, {T}: Create a token that's a copy of a card exiled with this
//	 artifact. It gains haste. Exile it at the beginning of the next end
//	 step."
//
// The imprint is a "may" on any nontoken creature's death, whoever
// controlled it. The exile is linked to the Vat object (ExileTarget.Linked,
// CR 607.2a) and happens only while the card is still the object that died
// (CR 400.7); once it has landed, the tail returns every OTHER card linked
// to this Vat to its owner's graveyard, so at most one card is ever
// imprinted. The activated ability copies the card where it sits in exile
// (its printed copiable values, CR 707.2); haste is part of the copy's
// values (CR 707.9a) and the exile is a delayed trigger on that token
// alone (CR 603.7). With nothing imprinted it does nothing.
//
// No simplification.
func init() {
	imprint := cmdr2S47VatImprint()
	Register(Spec{
		OracleID:     "6ebdc9f3-6b23-4716-8011-48f00d7e30c8",
		Name:         "Mimic Vat",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{imprint},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}: Create a token that's a copy of a card exiled with this artifact. It gains haste. Exile it at the beginning of the next end step.",
			Cost:  Plus(ManaCost("{3}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				linked := g.ExiledWithForEffect(game.PermissionCardRef(ref))
				if len(linked) == 0 {
					return nil
				}
				return trekB40CreateCopyExiledAtNextEndStep(ctx, item.Controller,
					linked[len(linked)-1].InstanceID, TokenCopyGainsHaste, "Mimic Vat — exile the token")
			},
		}},
	})
}

// cmdr2S47VatImprint is Mimic Vat's imprint trigger.
func cmdr2S47VatImprint() game.TriggeredAbility {
	tail := CardTail("mimic-vat/return-others", func(ctx *Context, o game.TailOutcome) error {
		if !oneWentThisWay(o) {
			return nil
		}
		ref, ok := ctx.SourceRef()
		if !ok {
			return nil
		}
		var others []uuid.UUID
		for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
			if c.InstanceID != o.IDs[0] {
				others = append(others, c.InstanceID)
			}
		}
		for _, id := range others {
			if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
				return err
			}
		}
		return nil
	})
	return Optional(On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		dead, ok := diedCreature(ev, g)
		return ok && !dead.IsToken()
	}, "Mimic Vat — exile that card; if you do, return each other card exiled with this artifact to its owner's graveyard",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			dead := item.Trigger.Event.CardID
			if !diedCardStillInGraveyard(ctx, dead) {
				return nil
			}
			return ExileTarget{Target: dead, Linked: true, ThenTail: ctx.ItemTail(tail)}.Apply(ctx)
		}), "Mimic Vat — exile that card?")
}
