package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emeritus of Conflict // Lightning Bolt — Creature — Human Wizard {1}{R},
// 2/2 // Instant {R} (Secrets of Strixhaven, #4456; preparation card,
// CR 722):
//
//	"First strike
//	 Whenever you cast your third spell each turn, this creature becomes
//	 prepared. (While it's prepared, you may cast a copy of its spell.
//	 Doing so unprepares it.)"
//
//	Lightning Bolt — "Lightning Bolt deals 3 damage to any target."
//
// It does not enter prepared. "Your third spell each turn" is Alphinaud
// Leveilleur's "second spell" clock moved on by one: the per-turn cast
// tally of the controller, which counts the spell being cast, reads
// exactly three. A copy is not cast (CR 707.10), so a copied spell does
// not move the count, and a fourth spell does not trigger it again. A
// creature that is already prepared gains nothing (CR 722.3a). The
// prepare spell is Lightning Bolt word for word, so it shares that card's
// body (lightning_bolt.go).
//
// No simplification.
const EmeritusOfConflictOracleID = "5963eef1-1022-42b1-8a0c-fc9850bfc2a3"

func init() {
	effects.Register(effects.Spec{
		OracleID:        EmeritusOfConflictOracleID,
		Name:            "Emeritus of Conflict",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.R37YouCastYourThirdSpellEachTurn, "Emeritus of Conflict — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BecomePrepared{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     EmeritusOfConflictOracleID + "#1",
		Name:         "Lightning Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
