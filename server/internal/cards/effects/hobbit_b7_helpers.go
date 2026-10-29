package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hobB7ArtifactsByOpponents is "the artifacts your opponents control": every
// artifact on the battlefield, tokens included, whose controller is not
// `you`. An eliminated player's permanents are already gone.
func hobB7ArtifactsByOpponents(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you && c.IsArtifact() {
			n++
		}
	}
	return n
}

// hobB7GreatestArtifactsAnOpponentControls is Cavern-Hoard Dragon's "X is the
// greatest number of artifacts an opponent controls": the largest per-player
// count among the caster's opponents (not their sum), zero when none has one.
func hobB7GreatestArtifactsAnOpponentControls() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil || q.Game.Battlefield == nil {
			return 0
		}
		perPlayer := map[uuid.UUID]int{}
		greatest := 0
		for _, c := range q.Game.Battlefield.Cards {
			if c.Controller == q.Controller || !c.IsArtifact() {
				continue
			}
			perPlayer[c.Controller]++
			if perPlayer[c.Controller] > greatest {
				greatest = perPlayer[c.Controller]
			}
		}
		return greatest
	}
}

// hobB7CounterLand is the shape of The Hobbit's two-colour "counter" lands
// (Iron Hills, Goblin-town):
//
//	"This land enters tapped.
//	 {T}: Add {A} or {B}.
//	 {2}{A}{B}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 <creature> you control. Activate only as a sorcery."
//
// The tapped entry is a CR 614 self-replacement, the mana is a plain tap
// ability, and the second ability is a CR 602 ability whose cost (mana, tap,
// sacrifice) is paid at announcement. The target is checked again at
// resolution (CR 608.2b); if it is gone the counters are simply not put
// (b31CountersOnChosenCreature, shared with Fain and Room of Refuge).
func hobB7CounterLand(oracleID, name, produced, manaLabel, activationCost, targetLabel string, victim CardPredicate) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: produced,
			Label:    manaLabel,
		}},
		Activated: []ActivatedAbility{{
			Label:        activationCost + ", {T}, Sacrifice this land: Put two +1/+1 counters on " + targetLabel + ". Activate only as a sorcery.",
			Cost:         Plus(ManaCost(activationCost), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Targets:      TargetCreature(targetLabel, YouControl(), victim),
			Effect:       b31CountersOnChosenCreature,
		}},
	}
}
