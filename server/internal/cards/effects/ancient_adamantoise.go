package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Adamantoise — Creature — Turtle {5}{G}{G}{G}, 8/20:
//
//	"Vigilance, ward {3}
//	 Damage isn't removed from this creature during cleanup steps.
//	 All damage that would be dealt to you and other permanents you
//	 control is dealt to this creature instead.
//	 When this creature dies, exile it and create ten tapped Treasure
//	 tokens."
//
// Four lines, one shape each:
//
//   - Vigilance is a printed keyword and ward {3} the shared Ward trigger.
//   - THE CLEANUP LINE is Spec.DamageStaysThroughCleanup (#2058): the
//     CR 514.2 sweep skips this permanent while it has the ability. A
//     permanent that lost its abilities is cleaned as usual, a
//     phased-out one is always cleaned (CR 702.26b), and the damage is
//     ordinary marked damage otherwise, so regeneration and leaving
//     the battlefield clear it and it counts for lethal damage on a
//     later turn.
//   - THE REDIRECTION is the same static Palisade Giant ships with
//     (CR 614.9, ADR 0108 §9).
//   - THE DIES TRIGGER exiles the card if it is still in the graveyard,
//     then makes ten tapped Treasures whether or not the exile worked
//     (a commander that went back to the command zone, CR 903.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                  "e6873252-653a-47a8-99b7-b2ef70aa1f7f",
		Name:                      "Ancient Adamantoise",
		Completeness:              CompletenessFull,
		PrintedKeywords:           []string{"vigilance"},
		DamageStaysThroughCleanup: true,
		Replacements: []game.ReplacementEffect{
			staticRedirection("Ancient Adamantoise — damage to you and your other permanents is dealt to it instead",
				redirectWhere{applies: damageToYouOrYourOtherPermanents, to: toThisPermanent}),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Ancient Adamantoise — ward {3}"),
			WhenThisDies("Ancient Adamantoise — exile it and create ten tapped Treasure tokens",
				ancientAdamantoiseDies),
		},
	})
}

func ancientAdamantoiseDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	treasures := func(ctx *Context) error {
		return b13CreateTappedTreasures(ctx, item.Controller, 10)
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return treasures(ctx)
	}
	return ExileTarget{
		Target: item.SourceCardID,
		Then:   func(ctx *Context, _ bool) error { return treasures(ctx) },
	}.Apply(ctx)
}
