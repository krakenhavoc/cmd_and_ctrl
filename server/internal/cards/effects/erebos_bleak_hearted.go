package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Erebos, Bleak-Hearted — Legendary Enchantment Creature — God
// {3}{B}, 5/6 (EDHREC rank 4128):
//
//	"Indestructible
//	 As long as your devotion to black is less than five, Erebos
//	 isn't a creature.
//	 Whenever another creature you control dies, you may pay 2 life.
//	 If you do, draw a card.
//	 {1}{B}, Sacrifice another creature: Target creature gets -2/-1
//	 until end of turn."
//
// The first Theros God in the catalog, and the one an aristocrats
// deck actually wants: a four-mana indestructible enchantment that
// turns every death into a card for 2 life, plus a sacrifice outlet
// that is also removal. The two halves feed each other — sacrifice a
// creature to the ability, pay 2 life to the trigger, shrink a
// blocker, draw a card, all off one creature.
//
// THE GOD CLAUSE IS A LAYER 4 TYPE-CHANGING STATIC, not a trigger and
// not a state-based action. "As long as" is a duration on a
// continuous effect (CR 611.2) that is re-evaluated on every layer
// recompute, so Erebos flickers between creature and non-creature as
// the board changes with no bookkeeping. While the clause is on he is
// an Enchantment God and nothing else: he cannot attack, cannot
// block, cannot be targeted by "target creature" removal, and is not
// hit by a Wrath — which is exactly why Gods are hard to kill.
//
// DEVOTION COUNTS MANA SYMBOLS, NOT PERMANENTS (CR 702.98a):
// every {B} and every hybrid symbol containing black in the mana
// costs of permanents you control, Erebos's own {B} included while he
// is on the battlefield. He contributes 1, so four other black pips
// on your board turn him on. devotionTo is the shared read, the same
// one Gray Merchant and Nykthos use.
//
// "ANOTHER CREATURE YOU CONTROL DIES" excludes Erebos himself — the
// printed "another" — and he is very often not a creature anyway.
// Tokens count: the clause has no nontoken rider.
//
// "YOU MAY PAY 2 LIFE. IF YOU DO, DRAW" is a COST (CR 118.3), asked
// as the trigger RESOLVES (MayChoice, CR 608.2c) and paid through the
// life-payment path: a controller below 2 life is not asked and does not
// draw, and the payment finishes before the draw begins.
//
// The activated ability's "-2/-1 until end of turn" is the S32
// turn-scoped registry, swept at cleanup (CR 514.2). It can target
// ANY creature, including your own — occasionally the right play to
// finish off a creature you are about to sacrifice anyway. The
// sacrifice is a COST paid at announce, and "another" is written the
// way every other sac-another clause in the catalog is: by NAME,
// because the cost's candidate filter never sees the source instance
// (#350). A token copy of Erebos would also be refused as fodder,
// which is weaker than printed and never stronger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a5690dc9-f12e-4ecc-b528-f2f83c79c3c8",
		Name:            "Erebos, Bleak-Hearted",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{{
			Label: "As long as your devotion to black is less than five, this isn't a creature.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					devotionTo(g, source.Controller, source.TextColor("B")) < 5
			},
			// "Isn't a creature" removes the creature type; the Legendary
			// and Enchantment halves of the printed line stay. The God
			// subtype goes with Creature (CR 205.1a, applied by the layer
			// pass).
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { notACreature(c) },
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b39AnotherCreatureYouControlDied(ev, source, g)
			}, "Erebos, Bleak-Hearted — pay 2 life, draw a card", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if p := g.PlayerByIDForEffect(item.Controller); p == nil || p.Eliminated || p.Life < 2 {
					return nil
				}
				return MayChoice{
					Question: "Erebos, Bleak-Hearted — another creature died. Pay 2 life to draw a card?",
					YesLabel: "Pay 2 life",
					NoLabel:  "Don't pay",
					LifeCost: 2,
					Then:     mayPayLifeDraw,
					Word:     "1",
				}.Apply(ctx)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature: Target creature gets -2/-1 until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -2, -1)),
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeAnotherN(1, "another creature", Creature())),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return BoostUntilEOT{
					Target:    id,
					Power:     -2,
					Toughness: -1,
					Label:     "Erebos, Bleak-Hearted — -2/-1",
				}.Apply(ctx)
			},
		}},
	})
}
