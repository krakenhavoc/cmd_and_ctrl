package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caparocti Sunborn — Legendary Creature — Human Soldier {2}{R}{W}, 4/4:
//
//	"Whenever Caparocti Sunborn attacks, you may tap two untapped
//	 artifacts and/or creatures you control. If you do, discover 3."
//
// The tap is a choice made as the trigger resolves: a card-set pick of
// exactly two (or none, which is the "you may" declined), over the
// untapped artifacts and creatures you control — Caparocti itself
// included if it has vigilance. Each pick is re-read before it is
// tapped, because the board can move under an open prompt, and the
// discover happens only if both were really tapped ("if you do").
// Discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "7f87a770-af47-4345-8dc1-0a1e8bae75c5",
		Name:         "Caparocti Sunborn",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Caparocti Sunborn — tap two artifacts and/or creatures to discover 3", caparoctiTapTwo),
		},
	})
}

func caparoctiTapTwo(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	candidates := caparoctiUntappedArtifactsOrCreatures(g, controller)
	if len(candidates) < 2 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   item.SourceCardID,
		Question: "Caparocti Sunborn — tap two untapped artifacts and/or creatures you control to discover 3?",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		Zone:     game.ZoneBattlefield,
		Validate: func(picked []game.Card) bool { return len(picked) == 2 },
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) != 2 {
				return nil
			}
			still := caparoctiUntappedArtifactsOrCreatures(g, controller)
			tapped := 0
			for _, id := range picked {
				if !slices.Contains(still, id) {
					continue
				}
				if err := g.TapTargetForEffect(id); err != nil {
					return err
				}
				tapped++
			}
			if tapped != 2 {
				return nil
			}
			return Discover{N: 3}.Apply(NewContext(g, item))
		},
	})
	return nil
}

// caparoctiUntappedArtifactsOrCreatures is every untapped artifact or
// creature `controller` controls.
func caparoctiUntappedArtifactsOrCreatures(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || c.Tapped {
			continue
		}
		if c.IsArtifact() || c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
