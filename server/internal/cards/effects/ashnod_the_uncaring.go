package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashnod the Uncaring — Legendary Creature — Human Artificer {2}{U}{B}{R},
// 1/4 (The Brothers' War Commander):
//
//	"Deathtouch
//	 Whenever you activate an ability of an artifact or creature that isn't
//	 a mana ability, if one or more permanents were sacrificed to activate
//	 it, you may copy that ability. You may choose new targets for the
//	 copy. (Sacrificing an artifact for mana to activate an ability doesn't
//	 count.)"
//
// The trigger watches CR 602 activations only (a mana ability never uses
// the stack, CR 605.3b, and announces a different event). Its intervening
// "if" (CR 603.4) is read off the activation's paid record
// (PaidCost.Sacrificed): the permanents the ability's own cost took,
// itself included. A sacrifice that paid for MANA the cost needed (a
// Treasure) is a cost of that mana ability, not of the activated one, and
// is not in the record. The ability's source counts as an artifact or a
// creature if it was one as the ability was activated (CR 113.7a), so a
// creature sacrificed to pay for its own ability still qualifies. The copy
// is a copy of the ability on the stack and is created, not activated
// (CR 707.10), so it never triggers Ashnod again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dcd5ab1b-adc5-486e-a1e0-636af6ee9470",
		Name:            "Ashnod the Uncaring",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventActivateAbility, ashnodActivatedWithASacrifice,
				"Ashnod the Uncaring — you may copy that ability",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CopyAbility{
						ItemID:           ctx.Trigger().Event.StackItemID,
						Controller:       item.Controller,
						ChooseNewTargets: true,
					}.Apply(ctx)
				}),
				"Ashnod the Uncaring — copy that ability?"),
		},
	})
}

// ashnodActivatedWithASacrifice is "you activate an ability of an artifact
// or creature that isn't a mana ability, if one or more permanents were
// sacrificed to activate it".
func ashnodActivatedWithASacrifice(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventActivateAbility || ev.Actor != source.Controller {
		return false
	}
	item := g.StackItemForEffect(ev.StackItemID)
	if item == nil || item.Paid.Sacrificed <= 0 {
		return false
	}
	if activationSourceMatches(g, ev, Or(Artifact(), Creature())) {
		return true
	}
	info, ok := g.LastKnownPermanentForEffect(item.SourceObject.ID)
	return ok && info.Epoch == item.SourceObject.Epoch &&
		(slices.Contains(info.Characteristic.Types, "Artifact") || slices.Contains(info.Characteristic.Types, "Creature"))
}
