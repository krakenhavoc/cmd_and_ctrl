package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merry, Esquire of Rohan — Legendary Creature — Halfling Knight
// {R}{W}, 2/2 (EDHREC rank 4340):
//
//	"Haste
//	 Merry has first strike as long as it's equipped.
//	 Whenever you attack with Merry and another legendary creature,
//	 draw a card."
//
// A two-mana legend that draws a card an attack in a deck full of
// other legends — which is every Boros legends deck, the only place
// this card is played. The first strike is conditional on an
// Equipment, which in that deck is not much of a condition.
//
// The conditional static is the interesting half. "As long as it's
// equipped" is a continuous effect whose CONDITION is read on every
// layer recompute, not a one-shot grant when a sword lands: equip a
// sword and the first strike appears in the same pass, destroy the
// sword and it is gone. It checks for an EQUIPMENT attached to Merry
// specifically — an Aura on him is not equipment (CR 301.5c), and
// neither is a Fortification.
//
// The attack trigger is anchored on MERRY's own declaration, which is
// what makes it fire exactly once: the engine emits one EventAttack
// per creature, and Merry gets exactly one. The "and another legendary
// creature" half is read off the battlefield, which is safe because
// the whole declaration is staged — every attacker has its
// AttackingTarget stamped — before the first EventAttack is emitted.
// So a single-creature attack by Merry sees no other legend and does
// not trigger, and an attack by Merry plus any other legendary
// creature its controller controls does.
//
// "Another LEGENDARY CREATURE", so a legendary artifact that is not a
// creature does not count, and a second copy of Merry would (before
// the legend rule settles it). They need not be attacking the same
// player — the card says "you attack with", and at a multiplayer
// table splitting an attack across two opponents still triggers it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4dce1f49-ddad-4748-893a-e10d831ee7d7",
		Name:            "Merry, Esquire of Rohan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if !b41SourceIsEquipped(g, source) {
					return
				}
				if !keywordSliceContains(c.Abilities, "first strike") {
					c.Abilities = append(c.Abilities, "first strike")
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && b41AnotherLegendaryCreatureYouControlIsAttacking(g, source)
			}, "Merry, Esquire of Rohan — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
