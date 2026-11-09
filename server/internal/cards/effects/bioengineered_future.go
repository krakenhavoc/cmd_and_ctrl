package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bioengineered Future — Enchantment {1}{G}{G} (Edge of Eternities,
// #4461):
//
//	"When this enchantment enters, create a Lander token.
//	 Each creature you control enters with an additional +1/+1 counter on
//	 it for each land that entered the battlefield under your control this
//	 turn."
//
// The enters trigger is Glacier Godmaw's Lander (b16LanderToken). The
// second ability is a CR 614.1c replacement on the enchantment over every
// creature entering under its controller's control (Gev, Scaled Scorch's
// shape, but "each creature" with no "other"): the count is the controller's
// landfall tally (PlayerTurnTally.LandsEntered), read as each creature
// enters, so lands that entered before the enchantment did still count, a
// land that is itself a creature entering is not counted for itself, and
// no land this turn means no counter. A creature that enters at the same
// time as the enchantment sees no replacement from it (the rulings: the
// source is not yet on the battlefield). A Lander token or the Lander's
// land fetched this turn raises the count for creatures that enter later.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6fa78914-85b3-46fd-977a-0e5b4d29eca4",
		Name:         "Bioengineered Future",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bioengineered Future — create a Lander token",
				Do(CreateToken{Template: b16LanderToken(), N: 1})),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.IsCreature() && entering.Controller == src.Controller &&
					g.TurnTallyFor(src.Controller).LandsEntered > 0
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				if n := g.TurnTallyFor(src.Controller).LandsEntered; n > 0 {
					ev.AddCounterAtETB(game.CounterPlusOne, n)
				}
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Bioengineered Future: each creature you control enters with an additional +1/+1 counter on it for each land that entered the battlefield under your control this turn",
		}},
	})
}
