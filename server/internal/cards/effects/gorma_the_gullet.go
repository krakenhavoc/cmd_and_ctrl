package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gorma, the Gullet — Legendary Creature — Pest Frog {1}{B}{G}, 1/1
// (Secrets of Strixhaven Commander, #4834):
//
//	"Lifelink
//	 Whenever another creature you control dies, put a +1/+1 counter on
//	 Gorma.
//	 Nontoken creatures you control enter with an additional +1/+1
//	 counter on them for each creature that died under your control this
//	 turn."
//
// The trigger is Garna's condition (another creature its controller
// controlled as it died) with Wildwood Scourge's counter on the source.
//
// The last ability is a CR 614 entry replacement, Grumgully's shape: the
// counters ride the entry event, so the creature is never on the
// battlefield without them and Hardened Scales or Doubling Season apply.
// The count is the controller's per-turn death tally
// (PlayerTurnTally.CreaturesDied, CR 700.4), read as the creature enters;
// a token gets nothing. The replacement works only while Gorma is on the
// battlefield, so Gorma itself, and any creature entering at the same
// time as Gorma, gets nothing (CR 614.12, and the card's ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ae5f30cc-a22e-4fbb-9510-e9429de3623c",
		Name:            "Gorma, the Gullet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return anotherCreatureYouControlDied(ev, source, g)
			}, "Gorma, the Gullet — put a +1/+1 counter on Gorma", b35PutCounterOnSelf),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.IsCreature() && !entering.IsToken() && entering.Controller == src.Controller &&
					g.TurnTallyFor(src.Controller).CreaturesDied > 0
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, g.TurnTallyFor(src.Controller).CreaturesDied)
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Gorma, the Gullet: enters with an additional +1/+1 counter for each creature that died under your control this turn",
		}},
	})
}
