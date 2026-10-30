package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const emeritusOfConflictOracleID = "5963eef1-1022-42b1-8a0c-fc9850bfc2a3"

func init() {
	Register(Spec{
		OracleID:        emeritusOfConflictOracleID,
		Name:            "Emeritus of Conflict",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, emeritusOfConflictThirdSpell, "Emeritus of Conflict — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     emeritusOfConflictOracleID + "#1",
		Name:         "Lightning Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve:    damageToFirstTarget(3),
	})
}

// emeritusOfConflictThirdSpell is "whenever you cast your third spell each
// turn" (YouCastYourSecondSpellEachTurn's shape).
func emeritusOfConflictThirdSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	return g.CastTallyFor(ev.Actor).Total == 3
}
