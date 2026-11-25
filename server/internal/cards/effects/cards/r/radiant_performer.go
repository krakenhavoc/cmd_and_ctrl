package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Radiant Performer — Creature — Human Wizard {3}{R}{R}, 2/2 (Commander
// Legends; OLD b057, #6159):
//
//	"Flash
//	 When this creature enters, if you cast it from your hand, choose
//	 target spell or ability that targets only a single permanent or
//	 player. Copy that spell or ability for each other permanent or player
//	 the spell or ability could target. Each copy targets a different one
//	 of those permanents and players."
//
// The intervening "if" is checked as the creature enters and again as the
// trigger resolves (CR 603.4). The copies are CopyForEachOtherPermanentOrPlayer
// (CR 707.10d): one for each other permanent or player the copied spell or
// ability could target, judged for the Performer's controller, who
// controls the copies and orders them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad16f918-5603-437b-b143-49c69ceae17d",
		Name:            "Radiant Performer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.Self(ev, source, lki, g) && effects.B30CastFromHand(g, source.InstanceID)
				}, "Radiant Performer — copy target spell or ability for each other permanent or player it could target",
					radiantPerformerCopies),
				effects.TargetSpellOrAbility("target spell or ability that targets only a single permanent or player",
					effects.ItemTargetsOnlyASinglePermanentOrPlayer()),
			),
		},
	})
}

func radiantPerformerCopies(g *game.Game, item *game.StackItem) error {
	if !effects.B30CastFromHand(g, item.SourceCardID) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			effects.CopyForEachOtherPermanentOrPlayer(ctx, t.ID, "Radiant Performer")
			return nil
		}
	}
	return nil
}
