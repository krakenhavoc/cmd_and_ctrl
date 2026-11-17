package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moira and Teshar — Legendary Creature — Phyrexian Spirit Bird {3}{W}{B},
// 4/5 (March of the Machine Commander):
//
//	"Flying
//	 Whenever you cast a historic spell, return target nonland permanent card
//	 from your graveyard to the battlefield. It gains haste. Exile it at the
//	 beginning of the next end step. If it would leave the battlefield, exile
//	 it instead of putting it anywhere else."
//
// Teshar's historic cast trigger with Whip of Erebos's return: the card comes
// back, gains haste until end of turn, is exiled at the next end step by a
// delayed trigger, and the shared "exile it instead" replacement
// (ExileInsteadOfLeavingBattlefield) covers it leaving any other way. A card
// that did not reach the battlefield (a target that left the graveyard, an
// Aura with nothing to enchant) gets none of the rider.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "445ad01c-c899-424d-8666-ee7bb2a6e535",
		Name:            "Moira and Teshar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17HistoricSpellCastByYou(ev, source, g)
			},
			Targets: TargetCardInGraveyard("target nonland permanent card from your graveyard",
				YouOwn(), Permanent(), Nonland()),
			Key:    "Moira and Teshar — return a nonland permanent card to the battlefield; it gains haste and is exiled at the next end step",
			Effect: moiraAndTesharReturn,
		}},
	})
}

func moiraAndTesharReturn(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	id := item.Targets[0].ID
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return moiraAndTesharRider(ctx, id)
}

// moiraAndTesharRider is the returned card's haste, its exile at the next end
// step and its exile-instead replacement.
func moiraAndTesharRider(ctx *Context, id uuid.UUID) error {
	if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
		Label: "Moira and Teshar — the returned permanent has haste"}).Apply(ctx); err != nil {
		return err
	}
	if err := (ScheduleDelayedTrigger{
		Label: "Moira and Teshar — exile the returned permanent",
		Cards: []uuid.UUID{id},
		Body:  exileListedCardsBody,
	}).Apply(ctx); err != nil {
		return err
	}
	ExileInsteadOfLeavingBattlefield(ctx.Game, id, ctx.Controller(),
		"Moira and Teshar: if it would leave the battlefield, exile it instead")
	return nil
}
