package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phylactery Lich — Creature {B}{B}{B}, 5/5:
//
//	"Indestructible
//	 As this creature enters, put a phylactery counter on an artifact you
//	 control.
//	 When you control no permanents with phylactery counters on them,
//	 sacrifice this creature."
//
// #1893, ADR 0170. The middle sentence is CR 614.1c's as-enters
// replacement with a choice of another permanent inside it
// (AsEntersChoosePermanents): the prompt opens before the Lich is on the
// battlefield, and the counter is on the artifact by the time it
// arrives. The counter is placed inline (mustSettleNow): a Doubling
// Season still doubles it, and a CR 616 order between two counter
// replacements is applied in the gathered order rather than asked,
// because the entry it sits inside cannot pause twice.
//
// "An artifact you control" is not a "may": with an artifact, one must
// be named. With none, nothing is asked, the Lich enters with no
// phylactery anywhere, and the last sentence sacrifices it. That last
// sentence is a CR 603.8 state trigger (ADR 0107 §1): it triggers
// whenever its controller controls no permanent with a phylactery
// counter, whatever took the counter away.
//
// No simplification.

const phylacteryCounter = "phylactery"

func init() {
	const name = "Phylactery Lich"
	game.RegisterEntryChoiceSuggestion(name, phylacteryLichSuggest)
	Register(Spec{
		OracleID:        "520c4d88-34b6-4a26-81ca-65eb8b10461c",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements: []game.ReplacementEffect{
			AsEntersChoosePermanents(name,
				"Phylactery Lich — as it enters, put a phylactery counter on an artifact you control.",
				&game.EntryCardChoice{
					Action:     game.EntryCardChoose,
					Matches:    func(c game.Card) bool { return c.IsArtifact() },
					Min:        1,
					Max:        1,
					Then:       phylacteryLichPutCounter,
					SuggestKey: name,
				}),
		},
		Triggered: []game.TriggeredAbility{
			WhenState("Phylactery Lich — sacrifice it", func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == controller && c.Counters[phylacteryCounter] > 0 {
						return false
					}
				}
				return true
			}, SacrificeThisIfStillOnBattlefield),
		},
	})
}

// phylacteryLichPutCounter is the entry's Then: one phylactery counter
// on the artifact named, placed by its controller (the Lich's).
func phylacteryLichPutCounter(g *game.Game, _ *game.ReplacementEvent, picked []uuid.UUID) error {
	for _, id := range picked {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if _, err := g.AddCounterByMustSettleNowForEffect(c.Controller, id, phylacteryCounter, 1); err != nil {
			return err
		}
	}
	return nil
}

// phylacteryLichSuggest is the artifact a seat would rather keep the
// phylactery on: the Lich dies with it, so the one least exposed to
// removal. A noncreature artifact before an artifact creature (creature
// removal and combat are everywhere), and a nontoken before a token;
// otherwise battlefield order.
func phylacteryLichSuggest(g *game.Game, _ uuid.UUID, candidates []uuid.UUID) []uuid.UUID {
	best, bestScore := uuid.Nil, -1
	for _, c := range cardsStillHere(g, candidates) {
		score := 0
		if !c.IsCreature() {
			score += 2
		}
		if !c.IsToken() {
			score++
		}
		if score > bestScore {
			best, bestScore = c.InstanceID, score
		}
	}
	if best == uuid.Nil {
		return nil
	}
	return []uuid.UUID{best}
}
