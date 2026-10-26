package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ultra Magnus, Tactician // Ultra Magnus, Armored Carrier (#3195, ADR
// 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {4}{R}{G}{W}, 7/7:
//
//	"More Than Meets the Eye {2}{R}{G}{W}
//	 Ward {2}
//	 Whenever Ultra Magnus attacks, you may put an artifact creature
//	 card from your hand onto the battlefield tapped and attacking. If
//	 you do, convert Ultra Magnus at end of combat."
//
// Back face, Legendary Artifact — Vehicle, 4/7:
//
//	"Living metal
//	 Haste
//	 Formidable — Whenever Ultra Magnus attacks, attacking creatures you
//	 control gain indestructible until end of turn. If those creatures
//	 have total power 8 or greater, convert Ultra Magnus."
//
// The card enters tapped and attacking a player, planeswalker or battle
// its controller chooses (CR 508.4, 506.3c), Magnus's own defender first.
// It was never declared as an attacker, so it triggers no "whenever
// attacks" ability. "If you do" is a card actually entering; the convert
// is a delayed trigger for the beginning of the end of combat step, and
// it records Magnus's turn-over count when it is created, so a Magnus
// that turned over meanwhile ignores it (CR 701.28e). "Those creatures"
// are the attackers the indestructible was given to, fixed as the ability
// resolves (CR 611.2c); their power is added up after that.
//
// No simplification.
const ultraMagnusOracleID = "48e913c0-efeb-47f5-9dcb-b795751e6b6d"

func init() {
	Register(Spec{
		OracleID:         ultraMagnusOracleID,
		Name:             "Ultra Magnus, Tactician",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{2}{R}{G}{W}")},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Ultra Magnus, Tactician — ward {2}"),
			WheneverThisAttacks("Ultra Magnus, Tactician — you may put an artifact creature card from your hand onto the battlefield tapped and attacking",
				ultraMagnusPutAttacker),
		},
	})
	Register(Spec{
		OracleID:        ultraMagnusOracleID + "#1",
		Name:            "Ultra Magnus, Armored Carrier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Ultra Magnus, Armored Carrier — formidable: attacking creatures you control gain indestructible; convert Ultra Magnus if their total power is 8 or greater",
				ultraMagnusFormidable),
		},
	})
}

func ultraMagnusPutAttacker(g *game.Game, item *game.StackItem) error {
	me := item.Controller
	candidates := handCardsMatching(g, me, And(Artifact(), Creature()))
	if len(candidates) == 0 {
		return nil
	}
	prefer := triggerAttackTarget(item)
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefieldTapped,
		Chooser:     me,
		Source:      item.SourceCardID,
		Question:    "Ultra Magnus, Tactician — you may put an artifact creature card from your hand onto the battlefield tapped and attacking",
		Cards:       candidates,
		Min:         0,
		Max:         1,
		Zone:        game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return ultraMagnusChooseDefender(g, item, picked[0], prefer)
		},
	})
	return nil
}

// ultraMagnusChooseDefender asks what the entering creature attacks when
// there is more than one choice (CR 508.4), Magnus's own defender first.
func ultraMagnusChooseDefender(g *game.Game, item *game.StackItem, card, prefer uuid.UUID) error {
	var targets []uuid.UUID
	for _, ref := range g.AttackTargetsForEffect(item.Controller) {
		targets = append(targets, ref.ID)
	}
	if i := slices.Index(targets, prefer); i > 0 {
		targets = append([]uuid.UUID{prefer}, slices.Delete(targets, i, i+1)...)
	}
	switch len(targets) {
	case 0:
		// Nothing to attack: it enters tapped, not attacking (CR 508.4a).
		return ultraMagnusPut(g, item, card, uuid.Nil)
	case 1:
		return ultraMagnusPut(g, item, card, targets[0])
	}
	options := make([]game.ChoiceOption, len(targets))
	for i, id := range targets {
		options[i] = game.ChoiceOption{Label: "Attack " + attackTargetName(g, id)}
	}
	return PickOption{
		Question: "Ultra Magnus, Tactician — what does it attack?",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(targets) {
				index = 0
			}
			return ultraMagnusPut(ctx.Game, item, card, targets[index])
		},
	}.Apply(NewContext(g, item))
}

func ultraMagnusPut(g *game.Game, item *game.StackItem, card, attacking uuid.UUID) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	return g.PutFromHandOntoBattlefieldKeyedForEffect(card, game.HandEntryOptions{
		Controller: item.Controller,
		Tapped:     true,
		Attacking:  attacking,
	}, NewContext(g, item).ItemTail(ultraMagnusConvert))
}

// ultraMagnusConvert is "if you do, convert Ultra Magnus at end of
// combat" as a keyed tail (#4253 slice 6). An on-disk key.
var ultraMagnusConvert = EnteredTail("ultra-magnus-tactician/convert-at-end-of-combat", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil // "If you do": nothing entered
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEndCombat,
		Label: "Ultra Magnus, Tactician — convert Ultra Magnus at end of combat",
		Body:  transformersConvertThisBody,
	}.Apply(ctx)
})

func ultraMagnusFormidable(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	attacking := func(_ *game.Game, controller uuid.UUID, c game.Card) bool {
		return c.Controller == controller && c.IsCreature() && c.AttackingTarget != uuid.Nil
	}
	if err := (GrantKeywordUntilEOT{
		Match:    attacking,
		Keywords: []string{"indestructible"},
		Label:    "Ultra Magnus, Armored Carrier — attacking creatures gain indestructible until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	g.RecomputeLayersIfStaleLocked()
	total := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if attacking(g, item.Controller, c) {
			total += c.CurrentPower()
		}
	}
	if total < 8 {
		return nil
	}
	return ConvertThis{}.Apply(ctx)
}
