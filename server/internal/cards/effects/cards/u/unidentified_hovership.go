package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unidentified Hovership — Artifact — Vehicle {1}{W}{W}, 2/2 (Duskmourn:
// House of Horror, #4466):
//
//	"Flying
//	 When this Vehicle enters, exile up to one target creature with
//	 toughness 5 or less.
//	 When this Vehicle leaves the battlefield, the exiled card's owner
//	 manifests dread.
//	 Crew 1"
//
// Skyclave Apparition's linked pair (CR 607.2a): the leave trigger reads
// what THIS object's enter trigger exiled, off the event record
// (b27EverExiledWith), so a Hovership that left and came back has no
// memory of its first exile, and one whose enter trigger exiled nothing
// pays nothing. Nothing comes back from exile; the owner of the exiled
// card manifests dread (CR 701.62a) instead, even if the card has since
// left exile, as the card is remembered rather than looked up there.
//
// Any creature with toughness 5 or less can be the target, your own
// included. Crew 1 is the shared crew ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bde67dfb-706b-4942-a365-984ea6e97a2b",
		Name:            "Unidentified Hovership",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(unidentifiedHovershipExileLabel, effects.B27ExileChosenTarget),
				effects.TargetCreature("up to one target creature with toughness 5 or less", effects.ToughnessLE(5)).WithCount(0, 1)),
			effects.On(game.EventLTB, effects.Self, "Unidentified Hovership — the exiled card's owner manifests dread",
				unidentifiedHovershipOwnerManifests),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Unidentified Hovership"),
		}},
	})
}

const unidentifiedHovershipExileLabel = "Unidentified Hovership — exile up to one target creature with toughness 5 or less"

// unidentifiedHovershipOwnerManifests has the exiled card's owner
// manifest dread. The enter trigger exiles at most one card.
func unidentifiedHovershipOwnerManifests(g *game.Game, item *game.StackItem) error {
	for _, id := range effects.B27EverExiledWith(g, item.SourceCardID, unidentifiedHovershipExileLabel) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		return effects.ManifestDread{Player: c.Owner}.Apply(effects.NewContext(g, item))
	}
	return nil
}
