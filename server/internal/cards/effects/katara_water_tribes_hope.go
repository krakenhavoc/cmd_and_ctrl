package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Katara, Water Tribe's Hope — Legendary Creature — Human Warrior
// Ally {2}{W}{U}{U}, 3/3:
//
//	"Vigilance
//	 When Katara enters, create a 1/1 white Ally creature token.
//	 Waterbend {X}: Creatures you control have base power and
//	 toughness X/X until end of turn. X can't be 0. Activate only
//	 during your turn. (While paying a waterbend cost, you can tap
//	 your artifacts and creatures to help. Each one pays for {1}.)"
//
// The proof card for waterbend {X} on an ACTIVATED ability (#1310):
// the cost composes three things the engine already had and one it
// did not.
//
//   - `WaterbendCost("{X}")` is the new one. The {X} goes in the mana
//     component, where the X prompt, the cost-modifier pass and every
//     affordability check read it; the waterbend clause beside it lets
//     the activator tap untapped artifacts and creatures they control,
//     each paying {1} of that X (CR 701.67a). So X=5 with five
//     creatures tapped costs no mana at all.
//   - `MinX(1)` is "X can't be 0", the floor Helm of Obedience
//     introduced — an announcement below it is refused, not cheap.
//   - `DuringYourTurn()` is "Activate only during your turn" (#743).
//   - The effect is Mass Diminish's layer 7b set, sized by the X the
//     activation announced (ctx.X, locked at CR 602.2b) and with its
//     affected set locked at resolution (CR 611.2c): a creature that
//     enters afterwards keeps its printed size.
//
// The ability doesn't target, so it has nothing to fizzle on. Tapping
// your creatures to pay does not stop them being affected — they are
// still creatures you control — which is the printed interaction the
// card is built around: pay with the team, make the team X/X.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "234fb291-0b62-4092-9071-81311c71bd53",
		Name:            "Katara, Water Tribe's Hope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Katara, Water Tribe's Hope — create a 1/1 white Ally",
				Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "Waterbend {X}: Creatures you control have base power and toughness X/X until end of turn. " +
				"X can't be 0. Activate only during your turn.",
			Cost:      Plus(WaterbendCost("{X}"), MinX(1)),
			Condition: DuringYourTurn(),
			Effect:    kataraBasePTX,
		}},
	})
}

// kataraBasePTX is "creatures you control have base power and
// toughness X/X until end of turn", with X read off the activation.
func kataraBasePTX(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := ctx.X()
	applies := SnapshotAffected(ctx, And(Creature(), ControlledBy(ctx.Controller())))
	if applies == nil {
		return nil
	}
	return StaticForDuration{
		Ability: game.StaticAbility{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: applies,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = x
				c.Toughness = x
			},
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Katara, Water Tribe's Hope — base X/X until end of turn",
	}.Apply(ctx)
}
