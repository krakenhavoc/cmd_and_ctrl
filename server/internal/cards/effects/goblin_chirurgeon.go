package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Chirurgeon — Creature — Goblin Shaman {R}, 0/2:
//
//	"Sacrifice a Goblin: Regenerate target creature."
//
// The goblin deck's repeatable shield: every spare token is another
// regeneration (CR 701.19a) for whatever matters. The Chirurgeon is
// itself a Goblin, so it can eat ITSELF to regenerate something else
// — which is why the cost is "Sacrifice a Goblin" rather than
// "Sacrifice another Goblin", and why SacrificeACreature's own
// comment about Carrion Feeder applies here too.
//
// Eating itself to regenerate itself does nothing useful: the
// sacrifice is a cost paid at announce (CR 601.2h), so the Chirurgeon
// is in the graveyard before the ability resolves and CR 701.19b
// regenerates a permanent that is not on the battlefield — nothing.
// The engine reaches that answer on its own; the shield is simply
// never created.
//
// Sacrifice is never destruction (CR 701.21a), so the Goblin the cost
// eats is not saved by a shield of its own.
func init() {
	Register(Spec{
		OracleID:     "ea55db87-a5c3-49f7-b968-77510cdeb469",
		Name:         "Goblin Chirurgeon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Goblin: Regenerate target creature.",
			Cost:    game.AbilityCost{SacrificeOther: sacrificeSpec("a Goblin", Subtype("Goblin"))},
			Targets: TargetCreature("target creature"),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
