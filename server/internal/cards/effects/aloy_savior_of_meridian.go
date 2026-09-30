package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aloy, Savior of Meridian — Legendary Creature — Human Warrior
// {3}{G}{U}, 3/5:
//
//	"Vigilance, reach
//	 In You, All Things Are Possible — Whenever one or more artifact
//	 creatures you control attack, discover X, where X is the greatest
//	 power among them."
//
// "One or more … attack" is one trigger per attack declaration
// (OncePerBatch, CR 603.2c). "Them" are the artifact creatures you
// control that are attacking as it resolves — the declaration's set,
// less any that have since left combat — and X is the greatest power
// among them then (CR 608.2h). Discover is ADR 0099's
// (game/discover.go).
func init() {
	Register(Spec{
		OracleID:        "f0554a8f-32de-4069-9f47-5e06ceb3f09d",
		Name:            "Aloy, Savior of Meridian",
		Completeness:    CompletenessFull,
		Discovers:       true,
		PrintedKeywords: []string{"vigilance", "reach"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, aloyArtifactCreatureAttacked,
				"Aloy, Savior of Meridian — In You, All Things Are Possible — discover X",
				func(g *game.Game, item *game.StackItem) error {
					return Discover{N: aloyGreatestAttackingArtifactPower(g, item.Controller)}.Apply(NewContext(g, item))
				})),
		},
	})
}

// aloyArtifactCreatureAttacked is "an artifact creature you control
// attacks".
func aloyArtifactCreatureAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsArtifact() && c.IsCreature()
}

// aloyGreatestAttackingArtifactPower is the greatest power among the
// attacking artifact creatures `controller` controls, zero when none
// are attacking any more.
func aloyGreatestAttackingArtifactPower(g *game.Game, controller uuid.UUID) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || c.AttackingTarget == uuid.Nil {
			continue
		}
		if !c.IsArtifact() || !c.IsCreature() {
			continue
		}
		if p := c.CurrentPower(); p > best {
			best = p
		}
	}
	return best
}
