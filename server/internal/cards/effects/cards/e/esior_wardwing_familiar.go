package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Esior, Wardwing Familiar — Legendary Creature — Bird {1}{U}, 1/3
// (Commander Legends, #6163):
//
//	"Flying
//	 Spells your opponents cast that target one or more commanders you
//	 control cost {3} more to cast.
//	 Partner (You can have two commanders if both have partner.)"
//
// A cost increase (CR 601.2f) that reads the announced targets
// (ReadsTargets), as Kopala's does: it applies once however many of the
// targets are commanders Esior's controller controls. A commander is read
// off the targeted object (Card.IsCommander), on the battlefield or on the
// stack. Partner is a deck-construction rule (CR 702.124) the deck checker
// reads off the printed text; it does nothing in play.
//
// No simplification.
func init() {
	tax := effects.CostsMore(3,
		"Spells your opponents cast that target one or more commanders you control cost {3} more to cast.",
		effects.OpponentsSpell(), esiorTargetsACommanderYouControl)
	tax.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:        "ddf38a32-302a-4658-9eaf-b813ebe275c9",
		Name:            "Esior, Wardwing Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers:   []game.CostModifier{tax},
	})
}

// esiorTargetsACommanderYouControl passes when one of the spell's targets
// is a commander controlled by the tax's source's controller.
func esiorTargetsACommanderYouControl(q game.CostQuery) bool {
	if q.Game == nil {
		return false
	}
	for _, t := range q.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if c, ok := q.Game.LookupCardForEffect(t.ID); ok && c.IsCommander && c.Controller == q.Source.Controller {
			return true
		}
	}
	return false
}
