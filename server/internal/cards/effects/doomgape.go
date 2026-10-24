package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doomgape — Creature — Elemental {4}{B/G}{B/G}{B/G}, 10/10:
//
//	"Trample
//	 At the beginning of your upkeep, sacrifice a creature. You gain
//	 life equal to that creature's toughness."
//
// Consuming Vapors' sacrifice-and-read body run by the controller on
// themselves. The sacrifice is mandatory and may be Doomgape itself;
// the life is the sacrificed creature's last-known toughness.
func init() {
	Register(Spec{
		OracleID:        "ab94d8f7-f811-436a-9a22-9a12bd929e98",
		Name:            "Doomgape",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Doomgape — sacrifice a creature, gain life equal to its toughness",
				func(g *game.Game, item *game.StackItem) error {
					// Keyed (#4290): the controller rides the run as its
					// carry, so a table waiting on the sacrifice is a
					// restore point.
					return g.PlayerSacrificesThenKeyedForEffect(item.SourceCardID, item.Controller,
						sacrificeSpec("a creature", Creature()),
						"Doomgape — sacrifice a creature", 1,
						doomgapeGain, []uuid.UUID{item.Controller})
				}),
		},
	})
}

// doomgapeGain is "you gain life equal to that creature's toughness",
// once the creature has left the battlefield (CR 608.2h).
var doomgapeGain = game.RegisterPromptRunThen("prompt-run/doomgape-gain", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	me := r.Carry[0]
	ids := r.Sacrifices().By(me)
	if len(ids) == 0 {
		return nil
	}
	t := departedCreatureToughness(g, ids[0])
	if t <= 0 {
		return nil
	}
	return GainLife{Player: me, Amount: t}.Apply(keyedContext(g, me, r.Source))
})
