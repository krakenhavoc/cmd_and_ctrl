package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graven Dominator — Creature — Gargoyle {4}{W}{W}, 4/4 (Guildpact):
//
//	"Flying
//	 Haunt (When this creature dies, exile it haunting target creature.)
//	 When this creature enters or the creature it haunts dies, each other
//	 creature has base power and toughness 1/1 until end of turn."
//
// Haunt is CR 702.55 (#6083). The second ability is one ability with two
// conditions. Its set of creatures is locked as it resolves (CR 611.2c);
// the Dominator itself, on the battlefield or not, is not in it. Base
// power and toughness is layer 7b, so counters and pumps still apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5462d711-037b-4c62-9238-c557a66534f8",
		Name:            "Graven Dominator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Haunt("Graven Dominator"),
			effects.WhenThisEntersOrTheCreatureItHauntsDies("Graven Dominator — each other creature has base power and toughness 1/1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					self := item.SourceCardID
					other := effects.And(effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Match:    other,
						Mods:     game.SetBasePTMods(1, 1),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Graven Dominator — base power and toughness 1/1 until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
