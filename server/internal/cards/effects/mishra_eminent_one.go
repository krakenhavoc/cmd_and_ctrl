package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mishra, Eminent One — Legendary Creature — Human Artificer
// {2}{U}{B}{R}, 5/4:
//
//	"At the beginning of combat on your turn, create a token that's a
//	 copy of target noncreature artifact you control, except its name
//	 is Mishra's Warform and it's a 4/4 Construct artifact creature in
//	 addition to its other types. It gains haste until end of turn.
//	 Sacrifice it at the beginning of the next end step."
//
// The Fire Crystal's shape (a token copy plus a CR 603.7 delayed
// sacrifice), with an except clause and a haste that lasts only for
// the turn.
//
// The except clause is part of the copy (CR 707.9b): the token's
// copiable name is Mishra's Warform, its P/T is 4/4, and Artifact,
// Creature and Construct are ADDED to the copied type line. The copied
// artifact's own subtypes (Equipment, Vehicle, Treasure) stay, so a
// copy of an Equipment is an Equipment Construct artifact creature,
// and CR 301.5c says it can't equip a creature.
//
// "It gains haste until end of turn" is NOT part of the copy: it is a
// layer-6 effect pinned to that one token for the turn
// (ScopedEffectFor), so a copy of the Warform made later does not have
// haste, and the token would lose it at cleanup if it survived its
// sacrifice.
//
// The target is re-read at resolution (CR 608.2b): an artifact that
// left or became a creature in response makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3438037-3efd-4ce0-88ec-6d48ab521992",
		Name:         "Mishra, Eminent One",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Mishra, Eminent One — create a Mishra's Warform token copy of target noncreature artifact you control",
					mishraEminentOneWarform),
				TargetPermanent("target noncreature artifact you control", Artifact(), Noncreature(), YouControl()),
			),
		},
	})
}

// mishraEminentOneWarform makes the Warform, gives it haste for the
// turn and schedules its sacrifice.
func mishraEminentOneWarform(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	copyOf := FirstLegalBattlefieldTarget(ctx)
	if copyOf == uuid.Nil {
		return nil
	}
	cursor := b25LastEventSeq(g)
	if err := (CreateTokenCopy{
		Controller: item.Controller,
		Copy:       copyOf,
		N:          1,
		Except:     mishrasWarformException,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	for _, token := range tokens {
		if err := (ScopedEffectFor{
			Target:   token,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Mishra, Eminent One — the Warform gains haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		Label: "Mishra, Eminent One — sacrifice the Warform",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}

// mishrasWarformException is "except its name is Mishra's Warform and
// it's a 4/4 Construct artifact creature in addition to its other
// types" (CR 707.9b).
func mishrasWarformException(t *game.Card) {
	v := game.PrintedValues{TypeLine: t.TypeLine}
	v.AddCardType("Artifact")
	v.AddCardType("Creature")
	v.AddSubtype("Construct")
	t.TypeLine = v.TypeLine
	t.Name = "Mishra's Warform"
	t.Power = 4
	t.Toughness = 4
	t.VariableToughness = false
	t.PrintedPTKnown = true
}
