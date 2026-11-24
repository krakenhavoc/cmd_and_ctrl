package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coffin Queen — Creature — Zombie Wizard {2}{B}, 1/1 (Secret Lair Drop):
//
//	"You may choose not to untap this creature during your untap step.
//	 {2}{B}, {T}: Put target creature card from a graveyard onto the
//	 battlefield under your control. When this creature becomes untapped or you
//	 lose control of this creature, exile that creature."
//
// The untap clause is Rubinia Soulsinger's opt-out. The ability puts the card
// onto the battlefield under the activator's control (Nighteyes the
// Desecrator's reanimation, as one entry), and as it enters the Queen remembers
// it as her linked permanent (CR 607.2e, Tyrannical Pitlord's seam). The second
// sentence is one trigger with its two conditions: this creature untaps, or its
// controller loses control of it (WhenYouLoseControlOfThis's controller rule:
// "you" is the player who lost it). It is built holding the permanent the Queen
// had linked when it triggered, so a second activation made in response (she
// has untapped) does not redirect the first trigger. The creature is exiled only
// while it is still that object on the battlefield (CR 400.7). A Queen that
// leaves the battlefield is no longer under her controller's control, so
// leaving counts as losing control.
//
// No simplification.
func init() {
	const label = "Coffin Queen — exile that creature"
	trigger := effects.OnAny([]game.EventKind{game.EventUntapCard, game.EventControlChanged, game.EventLTB},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			switch ev.Kind {
			case game.EventControlChanged:
				return effects.ThisChangedController(ev, source, lki, g)
			default:
				return effects.Self(ev, source, lki, g)
			}
		}, label, coffinQueenExile)
	trigger.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if ev.Kind == game.EventControlChanged {
			item.Controller, item.Owner = ev.Target, ev.Target
		}
		item.Params.Object = source.LinkedPermanent
		return item
	}
	effects.Register(effects.Spec{
		OracleID:     "c19ec6aa-aadd-4767-b695-384fd07b4ded",
		Name:         "Coffin Queen",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Coffin Queen — you may choose not to untap Coffin Queen")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}: Put target creature card from a graveyard onto the battlefield under your control.",
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.ReturnFromGraveyardTogether{Targets: []uuid.UUID{id}, Controller: ctx.Controller(), ThenTail: coffinQueenLinks}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{trigger},
	})
}

// coffinQueenLinks makes the creature that entered the Queen's linked
// permanent, the one "that creature" names. A Queen that has left remembers
// nothing (CR 400.7).
var coffinQueenLinks = effects.TogetherTail("coffin-queen/link", func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
	if len(entered) == 0 {
		return nil
	}
	src, ok := ctx.SourceRef()
	if !ok || !effects.OnBattlefield(ctx.Game, src.ID) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(entered[0])
	if !ok {
		return nil
	}
	ctx.Game.SetLinkedPermanentForEffect(src.ID, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
	return nil
})

// coffinQueenExile exiles the creature the trigger was built holding, while it
// is still that object on the battlefield.
func coffinQueenExile(g *game.Game, item *game.StackItem) error {
	ref := item.Params.Object
	if ref.ID == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok || c.ObjectEpoch != ref.Epoch || !effects.OnBattlefield(g, ref.ID) {
		return nil
	}
	return effects.ExileTarget{Target: ref.ID}.Apply(effects.NewContext(g, item))
}
