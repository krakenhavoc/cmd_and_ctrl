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
// "YOU MAY PAY 2 LIFE. IF YOU DO, DRAW" is a COST (CR 118.3), so it
// goes through the life-payment path: a controller at 2 or less life
// cannot pay and does not draw, and the payment finishes before the
// draw begins.
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
// SANDBOX SIMPLIFICATION, DECLARED — Crossway Troublemakers'. The
// "you may pay 2 life" decision is taken when the trigger goes on the
// stack rather than as it resolves, because the engine's optional
// trigger prompt is the announce-time one. Opponents therefore see
// the answer before the trigger resolves. It never changes what you
// can pay, only when you say so.
func init() {
	Register(Spec{
		OracleID:     "a5690dc9-f12e-4ecc-b528-f2f83c79c3c8",
		Name:         "Erebos, Bleak-Hearted",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Whether to pay 2 life for the card is decided when the trigger goes on the stack, not as it resolves.",
		},
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{{
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					devotionTo(g, source.Controller, "B") < 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				// "Isn't a creature" removes only the creature type;
				// the Legendary and Enchantment halves of the printed
				// line stay, and so does the God subtype, because the
				// clause says nothing about them.
				c.Types = []string{"Enchantment"}
			},
		}},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b39AnotherCreatureYouControlDied(ev, source, g)
			}, "Erebos, Bleak-Hearted — pay 2 life, draw a card", func(g *game.Game, item *game.StackItem) error {
				return payLifeThenDraw(NewContext(g, item), 2, 1)
			}),
				"Erebos, Bleak-Hearted — another creature died. Pay 2 life to draw a card?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature: Target creature gets -2/-1 until end of turn.",
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeN(1, "another creature", Creature(), b03NotNamed("Erebos, Bleak-Hearted"))),
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
