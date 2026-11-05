package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extraction Integrant — Artifact Creature — Borg {1}{B}, 1/1 (Star Trek
// Commander, #4974):
//
//	"Artifact creatures you control have menace and "Whenever a creature
//	 blocks this creature, that creature gets -1/-1 until end of turn.""
//
// Two layer-6 grants over the same set, read on post-layer types every
// pass: the keyword menace, and a catalog bundle (ADR 0093) holding the
// block trigger, so "this creature" is each recipient itself and the
// trigger is its own. The Integrant is an artifact creature, so it has
// both. A creature that stops being an artifact, or changes controller,
// loses them with the next layer pass; two Integrants grant two instances
// of the trigger, which trigger separately (CR 603.2), while a second
// menace is redundant (CR 702.111b).
//
// The trigger fires once for each creature that blocks the recipient
// (EventBlock names the blocker), and shrinks that blocker until end of
// turn if it is still on the battlefield as it resolves.
//
// No simplification.
const extractionIntegrantGrant = "extraction-integrant/blocker-shrinks"

func init() {
	const text = "Whenever a creature blocks this creature, that creature gets -1/-1 until end of turn."
	Register(Spec{
		OracleID:     "24964ce8-1eec-49b3-9db1-b895bc686ec5",
		Name:         "Extraction Integrant",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: extractionIntegrantGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Target == source.InstanceID && ev.CardID != uuid.Nil
				}, "Extraction Integrant grant — the creature that blocked it gets -1/-1 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						blocker := item.Trigger.Event.CardID
						if !onBattlefield(g, blocker) {
							return nil
						}
						return BoostUntilEOT{Target: blocker, Power: -1, Toughness: -1, Label: "Extraction Integrant — -1/-1"}.Apply(NewContext(g, item))
					}),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(trekB36ArtifactCreaturesYouControl, "menace"), "Artifact creatures you control have menace"),
			Labelled(GrantAbilities(trekB36ArtifactCreaturesYouControl, extractionIntegrantGrant), "Artifact creatures you control have \""+text+"\""),
		},
	})
}
