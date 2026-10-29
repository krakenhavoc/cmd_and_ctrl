package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const tomBertAndWilliamOracle = "0ea58cfe-b37c-49a6-a3be-7e60065b8238"

// Tom, Bert, and William — Legendary Creature — Troll {3}{B}{G}, 5/5
// (The Hobbit, #4454, #4506):
//
//		"{1}, Sacrifice another creature: Draw cards equal to the sacrificed
//		 creature's power, then discard a card.
//		 When Tom, Bert, and William die, if they were a creature, return
//		 them to the battlefield. They're an artifact. (They're no longer a
//		 creature.)"
//
//	  - The first ability is Ghoulcaller Gisa's cost with Greater Good's
//	    effect: the power is the sacrificed creature's last-known power
//	    (CR 608.2h), and the discard prompt opens once every draw is done
//	    (drawThenDiscard), so a drawn card can be discarded.
//	  - "If they were a creature" is CR 603.4's intervening if, read once
//	    from the last-known information the dies trigger is handed. They
//	    return under their owner's control, only if the card in the
//	    graveyard is still the object that died (CR 400.7).
//	  - "They're an artifact" is ADR 0210's entry overwrite: the card is an
//	    artifact from the moment it lands (CR 614.1c), never a creature on
//	    the battlefield, so nothing can kill it as one and a "whenever a
//	    creature enters" trigger does not see it. The card type is SET
//	    (CR 205.1a): Artifact, with Troll gone; Legendary stays (CR 205.4b).
//	    It keeps its abilities, so the sacrifice outlet still works. When
//	    the artifact dies, it was not a creature, the intervening if is
//	    false, and it stays in the graveyard: no loop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     tomBertAndWilliamOracle,
		Name:         "Tom, Bert, and William",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: Draw cards equal to the sacrificed creature's power, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return drawThenDiscard(g, item.Controller, item.SourceCardID, ctx.SacrificedPower(), 1, "")
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source) && keywordSliceContains(lki.Types, "Creature")
			}, "Tom, Bert, and William — return them to the battlefield; they're an artifact, not a creature",
				returnThemAsAnArtifact),
		},
	})
}

// returnThemAsAnArtifact is the dies trigger's effect: the card that
// died comes back under its owner's control as a noncreature artifact.
func returnThemAsAnArtifact(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		// It left the graveyard in response, and anything that came
		// back is a new object the trigger does not name (CR 400.7).
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{item.SourceCardID},
		Controller: uuid.Nil,
		EntersAs: &game.EntersAs{
			Types: []string{"Artifact"},
			Label: "Tom, Bert, and William — an artifact (no longer a creature)",
		},
	}.Apply(NewContext(g, item))
}
