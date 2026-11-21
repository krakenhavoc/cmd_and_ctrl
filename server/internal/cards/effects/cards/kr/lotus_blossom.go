package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lotus Blossom — Artifact {2} (Dominaria Remastered):
//
//	"At the beginning of your upkeep, you may put a petal counter on this
//	 artifact.
//	 {T}, Sacrifice this artifact: Add X mana of any one color, where X is
//	 the number of petal counters on this artifact."
//
// The upkeep counter is Barrin's Codex's "you may" trigger. The mana
// ability is Astral Cornucopia's one-colour pick over the counters on the
// source, read from the Blossom as it last existed once the sacrifice cost has
// taken it away (CR 608.2h; ProducedForPaid), or from the artifact itself
// while it is still on the battlefield. It is a mana ability (CR 605.1a): it uses no
// stack and no target.
//
// No simplification.
func init() {
	const petal = "petal"
	effects.Register(effects.Spec{
		OracleID:     "32e66ea0-5f35-4df8-8d00-580403c51bb5",
		Name:         "Lotus Blossom",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Lotus Blossom — put a petal counter on this artifact", effects.PutACounterOnThis(petal)),
				"Lotus Blossom — put a petal counter on it?"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{Tap: true, Sacrifice: true},
			ProducedForPaid: func(g *game.Game, _, source uuid.UUID, _ game.PaidCost) game.ManaOutput {
				return game.ManaOutput{Amount: game.ManaAmount{Options: game.AllColors, N: lotusBlossomPetals(g, source), OneColor: true}}
			},
			Label: "{T}, Sacrifice this artifact: Add X mana of any one color, where X is the number of petal counters on this artifact",
		}},
	})
}

// lotusBlossomPetals is the number of petal counters on the Blossom: live
// while it is on the battlefield, last-known once its cost sacrificed it.
func lotusBlossomPetals(g *game.Game, source uuid.UUID) int {
	if c, ok := g.LookupCardForEffect(source); ok && effects.OnBattlefield(g, source) {
		return c.Counters["petal"]
	}
	if info, ok := g.LastKnownPermanentForEffect(source); ok {
		return info.Counters["petal"]
	}
	return 0
}
