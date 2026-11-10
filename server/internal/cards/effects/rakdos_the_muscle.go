package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos, the Muscle — Legendary Creature — Demon Mercenary {2}{B}{B}{R},
// 6/5 (Outlaws of Thunder Junction, #4468):
//
//	"Flying, trample
//	 Whenever you sacrifice another creature, exile cards equal to its mana
//	 value from the top of target player's library. Until your next end
//	 step, you may play those cards, and mana of any type can be spent to
//	 cast those spells.
//	 Sacrifice another creature: Rakdos gains indestructible until end of
//	 turn. Tap it. Activate only once each turn."
//
// The sacrificed creature is read as it last existed (CR 603.10a, 608.2h),
// so a token (mana value 0) exiles nothing and a copy of a creature counts
// the mana value it had. The exile is an impulse exile (ExileTopWithPermission)
// from the chosen player's library whose permission runs until the beginning
// of the controller's next end step and lets any type of mana pay. The
// activated ability is once each turn per object (the engine's activation
// tally); it is not sacrifice-itself, so the cost names another creature.
//
// No simplification.
func init() {
	const label = "Sacrifice another creature: Rakdos gains indestructible until end of turn. Tap it. Activate only once each turn."
	Register(Spec{
		OracleID:        "e2c141af-925e-42bd-8d83-e1a71f4c9794",
		Name:            "Rakdos, the Muscle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventSacrifice, b35YouSacrificedAnotherCreature,
				"Rakdos, the Muscle — exile cards equal to its mana value from the top of target player's library; you may play them until your next end step",
				rakdosExile),
				TargetPlayer("target player")),
		},
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerSurvive},
			Cost:      game.AbilityCost{SacrificeOther: Another(sacrificeSpec("another creature", Creature()))},
			Condition: OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				if err := (GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"},
					Label: "Rakdos, the Muscle — indestructible"}).Apply(ctx); err != nil {
					return err
				}
				return TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}

// b35YouSacrificedAnotherCreature is "you sacrifice another creature": the
// source's controller's sacrifice of a creature (as it last existed) other
// than the source.
func b35YouSacrificedAnotherCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil || source == nil || ev.CardID == source.InstanceID || ev.Actor != source.Controller {
		return false
	}
	if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
		for _, t := range info.Characteristic.Types {
			if strings.EqualFold(t, "creature") {
				return true
			}
		}
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// rakdosExile exiles cards equal to the sacrificed creature's mana value
// off the top of the target player's library and lets the controller play
// them until their next end step, with mana of any type.
func rakdosExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := TargetedPlayer(ctx)
	if target == uuid.Nil {
		return nil
	}
	n := 0
	if info, ok := g.LastKnownPermanentForEffect(ctx.Trigger().Event.CardID); ok {
		n = info.ManaValue
	}
	if n <= 0 {
		return nil
	}
	window := g.UntilYourNextEndStepDuration(item.Controller)
	return ExileTopWithPermission{
		From:     target,
		GrantTo:  item.Controller,
		N:        n,
		AnyColor: true,
		AnyType:  true,
		Window:   &window,
	}.Apply(ctx)
}
