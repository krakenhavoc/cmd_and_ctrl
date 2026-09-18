package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// March of the World Ooze — Enchantment {3}{G}{G}{G} (EDHREC rank
// 4277):
//
//	"Creatures you control have base power and toughness 6/6 and are
//	 Oozes in addition to their other types.
//	 Whenever an opponent casts a spell, if it's not their turn, you
//	 create a 3/3 green Elephant creature token."
//
// Six mana that turns a board of mana dorks into a board of 6/6s, and
// taxes the table for interacting on your turn. The second ability is
// the one that wins games in Commander: at a four-player table,
// everybody's instant-speed answer during your turn hands you another
// blocker.
//
// The two static clauses are DIFFERENT LAYERS and are declared
// separately for that reason:
//
//   - "Base power and toughness 6/6" is layer 7b, a SET. It runs
//     before 7c and 7d, so a creature with two +1/+1 counters under
//     this is an 8/8, not a 6/6 — the counters are added on top of the
//     new base. That is the printed behaviour, and it is why the
//     sub-layer is not decoration.
//   - "Are Oozes IN ADDITION to their other types" is layer 4, and it
//     ADDS the subtype rather than replacing the type line. An Elf
//     Druid stays an Elf Druid and is also an Ooze, so it keeps
//     feeding an Elvish Archdruid.
//
// Both apply to creatures you control on EVERY recompute — unlike a
// one-shot pump, a creature cast after the March lands is a 6/6 Ooze
// too, which is what a battlefield static means.
//
// "If it's not their turn" is an intervening-if (CR 603.4), so it is
// checked when the spell is cast and again when the token trigger
// resolves. It is the active-player read, not "your turn": at a
// four-player table an opponent casting on a THIRD player's turn also
// triggers it, which is the printed text and is often overlooked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a04696c4-4138-47e2-8da4-81d61748519f",
		Name:         "March of the World Ooze",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: b41CreatureYouControlStatic,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power = 6
					c.Toughness = 6
				},
			},
			{
				Layer:     game.Layer4Type,
				AppliesTo: b41CreatureYouControlStatic,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, s := range c.Subtypes {
						if equalFoldASCIIEffects(s, "Ooze") {
							return
						}
					}
					c.Subtypes = append(c.Subtypes, "Ooze")
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != source.Controller && b29SpellCastOffTurn(ev, g)
			}, "March of the World Ooze — create a 3/3 green Elephant",
				Do(CreateToken{Template: TokenCard("3/3 green Elephant"), N: 1})),
		},
	})
}

// b41CreatureYouControlStatic is the March's scope: a creature the
// source's controller controls, the source itself included when it is
// one (the March is an enchantment, so it never is).
//
// Post-layer types, which is what lets the two clauses compose: the
// layer-4 pass that adds Ooze and the layer-7b pass that sets 6/6 both
// ask this, and an artifact something animated in between is caught by
// the later one.
func b41CreatureYouControlStatic(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller
}
