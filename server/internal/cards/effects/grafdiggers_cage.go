package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grafdigger's Cage — {1} Artifact:
//
//	"Creature cards in graveyards and libraries can't enter the
//	 battlefield.
//	 Players can't cast spells from graveyards or libraries."
//
// The ZONE-shaped cast restriction, and the one that proves the gate
// sees where a cast is coming FROM. It is here rather than Meddling
// Mage because Meddling Mage needs a choose-a-card-name prompt the
// engine does not have — Card.NamedTribe and Card.ChosenColor exist,
// a chosen NAME does not (ADR 0073 §8).
//
// Sentence two is the cast gate. It costs nothing that the zones in
// question are exactly the two ADR 0066 just opened: a flashed-back
// Faithless Looting and a Bolas's Citadel cast off the top are both
// refused, and the enumerator and the view agree because all three read
// the same gate.
//
// Sentence one is a CR 614 replacement that cancels the entry: a
// creature card moving from a graveyard or a library onto the
// battlefield (reanimation, a library search, a put from the top of the
// library) is refused and the card stays where it was, which is what
// "can't enter" means (CR 614.17). A creature that comes from anywhere
// else (hand, exile, the stack) is not touched, and neither is a
// noncreature card, nor a token (it has no zone to come from).
func init() {
	Register(Spec{
		OracleID:     "753cb2b2-24ce-484f-a2d0-be6fd2c67ebd",
		Name:         "Grafdigger's Cage",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			PlayersCantCastFrom(
				"Grafdigger's Cage — players can't cast spells from graveyards or libraries.",
				game.ZoneGraveyard, game.ZoneLibrary),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
					return false
				}
				if ev.OldZone != game.ZoneGraveyard && ev.OldZone != game.ZoneLibrary {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Grafdigger's Cage: creature cards in graveyards and libraries can't enter the battlefield",
		}},
	})
}
