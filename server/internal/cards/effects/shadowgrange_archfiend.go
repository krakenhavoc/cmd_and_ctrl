package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shadowgrange Archfiend — Creature — Demon {6}{B}, 8/4:
//
//	"When this creature enters, each opponent sacrifices a creature with
//	 the greatest power among creatures they control. You gain life equal
//	 to the greatest power among creatures sacrificed this way.
//	 Madness—{2}{B}, Pay 8 life. (If you discard this card, discard it
//	 into exile. When you do, cast it for its madness cost or put it into
//	 your graveyard.)"
//
// Each opponent chooses their own victim among the creatures tied for the
// greatest power (GreatestPowerYouControl is read for the chooser, as Will
// of the Abzan's is), through the same sacrifice prompt every "of their
// choice" effect uses, so nothing targets and hexproof is irrelevant. An
// opponent with no creature is skipped.
//
// The life is the greatest of the sacrificed creatures' powers as each
// last stood on the battlefield (CR 608.2h), read after the last prompt is
// answered, so a pump or an anthem counts and a negative power gains
// nothing.
//
// Madness is the Spec.Madness string plus Spec.MadnessLife for the "Pay 8
// life" half: the offer charges {2}{B} and 8 life, and a player with fewer
// than 8 life cannot cast it that way (CR 119.4). The life is paid as part
// of casting, so it is paid whether or not the spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f3705ff-e8e5-48e7-ab68-df6849d78867",
		Name:         "Shadowgrange Archfiend",
		Completeness: CompletenessFull,
		Madness:      "{2}{B}",
		MadnessLife:  8,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self,
				"Shadowgrange Archfiend — each opponent sacrifices a creature with the greatest power; you gain that much life",
				shadowgrangeArchfiendETB),
		},
	})
}

func shadowgrangeArchfiendETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := item.Controller
	gain := func(g *game.Game, sacrificed game.PromptedSacrifices) error {
		best := 0
		for _, id := range sacrificed.Cards() {
			if info, ok := g.LastKnownPermanentForEffect(id); ok && info.Power > best {
				best = info.Power
			}
		}
		return GainLife{Player: controller, Amount: best}.Apply(NewContext(g, item))
	}
	opponents := ctx.Opponents()
	if len(opponents) == 0 {
		return nil
	}
	return g.PlayersSacrificeThenForEffect(item.SourceCardID, opponents,
		sacrificeSpec("a creature with the greatest power among creatures you control", GreatestPowerYouControl()),
		"Shadowgrange Archfiend — sacrifice a creature with the greatest power", gain)
}
