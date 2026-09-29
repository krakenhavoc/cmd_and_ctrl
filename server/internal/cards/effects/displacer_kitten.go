package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Displacer Kitten — Creature — Cat Beast {3}{U}, 2/2:
//
//	"Avoidance — Whenever you cast a noncreature spell, exile up to
//	 one target nonland permanent you control, then return that card
//	 to the battlefield under its owner's control."
//
// "Avoidance" is the ability's printed NAME, not a keyword the engine
// tracks — it carries no rules content of its own, so nothing is
// declared for it beyond the triggered ability it names. "Whenever you
// cast a noncreature spell" is Flux Channeler's own condition
// (b10NoncreatureSpellCastByYou), shared rather than re-written. The
// body is an ordinary immediate Flicker (S22): exile and return
// resolve together, the permanent comes back a new object (CR 400.7)
// and re-fires its ETBs, which is the whole reason this is a combo
// piece. "Up to one" is TargetPermanent's (0, 1) count, and leaving
// Flicker.Controller unset returns it "under its owner's control" as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09ef446c-a13d-49d9-a94c-cd5f5a2d440b",
		Name:         "Displacer Kitten",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: b10NoncreatureSpellCastByYou,
			Targets: TargetPermanent("up to one target nonland permanent you control",
				Not(Land()), YouControl()).WithCount(0, 1),
			Key:    "Displacer Kitten — exile and return a nonland permanent you control",
			Effect: displacerKittenFlicker,
		}},
	})
}

// displacerKittenFlicker is the "exile it, then return it" body. "Up
// to one" means the target may be empty — nothing exiled, nothing
// returned — which is a legal, unremarkable answer.
func displacerKittenFlicker(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return Flicker{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
}
