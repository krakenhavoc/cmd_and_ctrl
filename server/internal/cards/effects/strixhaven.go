package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strixhaven — Plane — Arcavios (Commander Masters, #4870):
//
//	"Instant and sorcery spells players cast have demonstrate. (Whenever a
//	 player casts an instant or sorcery spell, they may copy it. If they do,
//	 they choose an opponent to also copy it. Players may choose new targets
//	 for their copies.)
//	 Whenever chaos ensues, return up to one target instant or sorcery card
//	 from a graveyard to its owner's hand."
//
// A plane (ADR 0216 §6). Demonstrate (CR 702.144a) is a cast trigger of the
// spell; the plane's version is the same trigger for every player's instant
// or sorcery spell, controlled by the CASTER (the stack item is built with
// them as its controller, so "you" and the question are theirs, and a copy
// they make is theirs), and it resolves before the spell: the copy goes on the
// stack above the original, and the opponent's copy above that, per the
// rulings. Declining means no opponent copies. A copy is not cast, so it
// doesn't demonstrate again. The chaos ability (CR 311.7) returns up to
// one instant or sorcery card from any graveyard to its owner's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "022e38c3-fcea-428f-a0d0-a6cca057faa3",
		Name:         "Strixhaven",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			strixhavenDemonstrate(),
			Targeting(WheneverChaosEnsues("Strixhaven — return up to one target instant or sorcery card from a graveyard to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ReturnFromGraveyard{Target: t, Dest: game.ZoneHand}.Apply(ctx)
				}), TargetCardInGraveyard("up to one target instant or sorcery card from a graveyard", Or(Instant(), Sorcery())).WithCount(0, 1)),
		},
	})
}

const strixhavenDemonstrateLabel = "Strixhaven — demonstrate: the caster may copy that spell; if they do, they choose an opponent to also copy it"

// strixhavenDemonstrate is demonstrate granted to every instant or sorcery
// spell: the keyword's cast trigger (see Demonstrate), whose source is the
// plane and whose controller is the caster.
func strixhavenDemonstrate() game.TriggeredAbility {
	t := On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		spell, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (spell.HasCardType("instant") || spell.HasCardType("sorcery"))
	}, strixhavenDemonstrateLabel, strixhavenDemonstrateEffect)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, strixhavenDemonstrateLabel)
		item.Controller, item.Owner = ev.Actor, ev.Actor
		return item
	}
	t = Optional(t, "Strixhaven — demonstrate: copy this spell? If you do, choose an opponent to also copy it.")
	t.OptionalPrompt.Chooser = func(ev game.Event, _ *game.Card, _ *game.Game) uuid.UUID { return ev.Actor }
	return t
}

// strixhavenDemonstrateEffect copies the spell that was just cast for its
// caster and has them choose an opponent who also copies it (Demonstrate's
// body, over the spell the trigger's event names).
func strixhavenDemonstrateEffect(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	spell := item.Trigger.Event.CardID
	if err := (CopySpell{StackID: spell, ChooseNewTargets: true, FromLastKnown: true}).Apply(ctx); err != nil {
		return err
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Demonstrate — choose an opponent to also copy the spell",
		ThenKey:  demonstrateOpponentCopies,
		Carry:    []uuid.UUID{spell},
	}.Apply(ctx)
}
