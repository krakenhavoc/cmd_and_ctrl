package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Senu, Keen-Eyed Protector — Legendary Creature — Bird Scout {1}{W}, 2/1
// (Assassin's Creed, #5948):
//
//	"Flying, vigilance
//	 {T}, Exile Senu: You gain 2 life and scry 2.
//	 When a legendary creature you control attacks and isn't blocked, if this
//	 card is exiled, put it onto the battlefield attacking."
//
// The activated ability's exile is a cost (the ability exiles Senu whether or not
// she was the only legendary creature). The third ability works only from exile
// (CR 113.6; InExile, with "you" the card's owner there): it triggers when a
// defender's block declaration completes with a legendary creature you control
// attacking them and nothing blocking it (attacks_unblocked.go's reading of "isn't
// blocked"), and is an intervening "if" (CR 603.4): the card must still be
// exiled as it resolves. She enters untapped, attacking a player of your choice
// (CR 508.4: you choose, so the choice is asked when there is more than one
// opponent); she was never declared as an attacker, so "whenever attacks" triggers
// don't see her, and she is not blocked (blockers were already declared).
//
// No simplification.
func init() {
	const label = "Senu, Keen-Eyed Protector — put it onto the battlefield attacking"
	effects.Register(effects.Spec{
		OracleID:        "bf2dc318-7b51-451e-ac4d-0f1b756dc860",
		Name:            "Senu, Keen-Eyed Protector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Exile Senu: You gain 2 life and scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.ExileThis()),
			Effect:  effects.Do(effects.GainLife{Amount: 2}, effects.Scry{N: 2}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.InExile(effects.On(game.EventBlockersDeclared, senuLegendaryUnblocked, label, senuEnters)),
		},
	})
}

// senuLegendaryUnblocked is "a legendary creature you control attacks and isn't
// blocked": a completed block declaration by the player a legendary creature of
// yours is attacking, with nothing declared blocking it.
func senuLegendaryUnblocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBlockersDeclared {
		return false
	}
	legend := effects.Legendary()
	for _, c := range g.Battlefield.Cards {
		if c.Controller != source.Controller || !c.IsCreature() || !legend(g, source.Controller, c) || c.AttackingTarget == uuid.Nil {
			continue
		}
		if g.DefendingPlayerForAttackerForEffect(c.InstanceID) == ev.Actor && g.UnblockedAttackerForEffect(c.InstanceID) {
			return true
		}
	}
	return false
}

// senuEnters puts the card onto the battlefield, if it is still exiled, and then
// attacking.
func senuEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return effects.ReturnFromExile{Target: item.SourceCardID, Controller: item.Controller, ThenTail: ctx.ItemTail(senuEntered)}.Apply(ctx)
}

// senuEntered is told the new object, and puts it into combat. An on-disk key.
var senuEntered = effects.EnteredTail("senu-keen-eyed-protector/entered", func(ctx *effects.Context, id uuid.UUID, _ game.TailOutcome) error {
	if id == uuid.Nil {
		return nil
	}
	targets := ctx.Game.AttackTargetsForEffect(ctx.Controller())
	switch len(targets) {
	case 0:
		return nil
	case 1:
		ctx.Game.PutOnBattlefieldAttackingForEffect(id, targets[0].ID)
		return nil
	}
	options := make([]game.ChoiceOption, 0, len(targets))
	for _, t := range targets {
		options = append(options, game.ChoiceOption{Label: "Attack " + effects.AttackTargetName(ctx.Game, t.ID)})
	}
	return effects.PickOption{
		Question: "Senu, Keen-Eyed Protector — which player or planeswalker is she attacking?",
		Options:  options,
		ThenKey:  senuAttackPicked,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
})

// senuAttackPicked puts her into combat against the chosen defender. An on-disk
// key.
var senuAttackPicked = effects.OptionPickThen("option-pick/senu-keen-eyed-protector-attack", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 {
		return nil
	}
	targets := ctx.Game.AttackTargetsForEffect(ctx.Controller())
	if r.Index >= len(targets) {
		return nil
	}
	ctx.Game.PutOnBattlefieldAttackingForEffect(r.Carry[0], targets[r.Index].ID)
	return nil
})
