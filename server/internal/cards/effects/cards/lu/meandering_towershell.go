package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Meandering Towershell — Creature — Turtle {3}{G}{G}, 5/9 (Secret Lair Drop,
// #5462):
//
//	"Islandwalk
//	 Whenever this creature attacks, exile it. Return it to the battlefield
//	 under your control tapped and attacking at the beginning of the declare
//	 attackers step on your next turn."
//
// Islandwalk is a printed keyword. The attack trigger exiles the Towershell
// (it leaves combat with it) and, once it is really in exile, schedules a delayed
// trigger (CR 603.7) for the beginning of the declare attackers step of the
// controller's next turn, strictly after this one (AfterThisTurn, Kav
// Landseeker's reading), about that exiled object (CR 400.7): a Towershell
// that left exile before then is not returned. The return puts it onto the
// battlefield tapped and attacking a player, planeswalker or battle its
// controller chooses (CR 508.4), as Ilharg's put does, never declared as an
// attacker, so it triggers nothing that watches attacks and its own attack
// trigger does not fire. With nothing to attack it enters tapped, not
// attacking (CR 508.4a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "45154dc6-ff0a-4e99-8528-689757d1778e",
		Name:            "Meandering Towershell",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Meandering Towershell — exile it; return it tapped and attacking at the beginning of the declare attackers step on your next turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(meanderingTowershellExiled)}.Apply(ctx)
				}),
		},
	})
}

// meanderingTowershellExiled schedules the return once the Towershell is in
// exile. An on-disk key: never renamed, never reused.
var meanderingTowershellExiled = effects.CardTail("meandering-towershell/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:                 game.StepDeclareAttackers,
		ControllerTurnOnly: true,
		AfterThisTurn:      true,
		Label:              "Meandering Towershell — return it to the battlefield tapped and attacking",
		Cards:              []uuid.UUID{c.InstanceID},
		Body:               meanderingTowershellReturnBody,
		Params:             game.EffectParams{Object: game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

// meanderingTowershellReturnBody is the delayed trigger's resolution. An
// on-disk key: never renamed, never reused.
var meanderingTowershellReturnBody = game.DelayedBody("meandering-towershell/return", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	targets := g.AttackTargetsForEffect(item.Controller)
	switch len(targets) {
	case 0:
		return meanderingTowershellPut(g, item.Controller, id, uuid.Nil)
	case 1:
		return meanderingTowershellPut(g, item.Controller, id, targets[0].ID)
	}
	options := make([]game.ChoiceOption, len(targets))
	carry := []uuid.UUID{id}
	for i, t := range targets {
		options[i] = game.ChoiceOption{Label: "Attack " + effects.AttackTargetName(g, t.ID)}
		carry = append(carry, t.ID)
	}
	return effects.PickOption{
		Question: "Meandering Towershell — what does it attack?",
		Options:  options,
		ThenKey:  meanderingTowershellAttacks,
		WithItem: true,
		Carry:    carry,
	}.Apply(effects.NewContext(g, item))
})

// meanderingTowershellAttacks is the answer to "what does it attack?". An
// on-disk key: never renamed, never reused.
var meanderingTowershellAttacks = effects.OptionPickThen("option-pick/meandering-towershell-attacks", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) < 2 {
		return nil
	}
	card, targets := r.Carry[0], r.Carry[1:]
	index := r.Index
	if index < 0 || index >= len(targets) {
		index = 0
	}
	if z := ctx.Game.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return meanderingTowershellPut(ctx.Game, ctx.Controller(), card, targets[index])
})

// meanderingTowershellPut moves the card from exile, tapped and attacking
// `attacking` (uuid.Nil: not attacking), under `controller`'s control.
func meanderingTowershellPut(g *game.Game, controller, card, attacking uuid.UUID) error {
	return g.PutOntoBattlefieldTogetherKeyedForEffect([]game.BatchEntry{{CardID: card, From: game.ZoneExile}},
		game.ZoneEntryOptions{Controller: controller, Tapped: true, Attacking: attacking}, game.KeyedTail{})
}
