package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Berserk Murlodont — Creature — Beast {4}{G}, 3/3 (Legions, #5609):
//
//	"Whenever a Beast becomes blocked, it gets +1/+1 until end of turn
//	 for each creature blocking it."
//
// Any Beast, whoever controls it, the Murlodont included. "Becomes
// blocked" is the once-per-attacker EventBecomesBlocked (CR 509.1h), so
// a double block is one trigger. The blockers are counted as it resolves,
// the way rampage counts them (game.CreaturesBlockingForEffect): a
// blocker removed from combat in response no longer counts, and the
// bonus is fixed then. The Beast is the object that became blocked: one
// that left the battlefield in response gets nothing (CR 400.7).
// "Beast" is read through the Murlodont's own text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a762e8cb-9d33-41fc-b2e8-0c5ad4d7268e",
		Name:         "Berserk Murlodont",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesBlocked, berserkMurlodontABeastBecameBlocked,
				"Berserk Murlodont — it gets +1/+1 until end of turn for each creature blocking it",
				berserkMurlodontPump),
		},
	})
}

// berserkMurlodontABeastBecameBlocked is "whenever a Beast becomes
// blocked": the attacker the event names is a Beast creature as it
// becomes blocked.
func berserkMurlodontABeastBecameBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesBlocked {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && OfCreatureType("Beast")(g, source.Controller, game.ThroughTextOf(source, c))
}

// berserkMurlodontPump gives the Beast +1/+1 for each creature blocking
// it now.
func berserkMurlodontPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
		return nil
	}
	id := obj.Ref().ID
	n := g.CreaturesBlockingForEffect(id)
	return BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Berserk Murlodont — +1/+1 for each creature blocking it"}.Apply(ctx)
}
