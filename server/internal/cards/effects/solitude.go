package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solitude — Creature — Elemental Incarnation {3}{W}{W}, 3/2:
//
//	"Flash
//	 Lifelink
//	 When this creature enters, exile up to one other target creature.
//	 That creature's controller gains life equal to its power.
//	 Evoke—Exile a white card from your hand."
//
// Evoke with a CARD rather than mana, which is the Modern Horizons 2
// incarnation cycle's whole design: a five-drop that costs one other
// white card at instant speed.
//
// The evoke sacrifice is the part that makes the card work and the
// part a hand-written AlternativeCost would forget. CR 702.74b makes
// it a TRIGGERED ability, so the sequence is: Solitude enters → its
// own ETB trigger goes on the stack → the sacrifice trigger goes on
// the stack → the creature is exiled and its controller gains life →
// only then does Solitude die. Sacrificing inside the resolution
// would kill it before its own trigger resolved.
//
// "Up to one other target creature" is a zero-or-one clause, so an
// evoked Solitude with no legal target still enters and still dies —
// which is a real, if sad, play.
func init() {
	Register(Spec{
		OracleID:        "dcb9c2a7-ae54-4ddc-a567-640bf4bf4366",
		Name:            "Solitude",
		PrintedKeywords: []string{"flash", "lifelink"},
		AlternativeCosts: []game.AlternativeCost{
			EvokePitch(
				CardInYourHand("a white card from your hand", OfColor("W")),
				"a white card from your hand",
			),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCreature("up to one other target creature").WithCount(0, 1),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				self := source.InstanceID
				return game.NewTriggeredItem(source, "Solitude — exile a creature, its controller gains life",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						victim := item.Targets[0].ID
						if victim == self {
							// "OTHER target creature" — the picker
							// should never offer Solitude itself, and
							// refusing here costs one comparison.
							return nil
						}
						ctx := NewContext(g, item)
						// Power is read BEFORE the exile: once the
						// card leaves the battlefield its layered
						// characteristics are gone, and "life equal
						// to its power" means its power as it last
						// existed (CR 608.2h / LKI).
						card, ok := g.LookupCardForEffect(victim)
						if !ok {
							return nil
						}
						power := card.CurrentPower()
						controller := card.Controller
						if err := (ExileTarget{Target: victim}).Apply(ctx); err != nil {
							return err
						}
						if power <= 0 {
							return nil
						}
						return GainLife{Player: controller, Amount: power}.Apply(ctx)
					})
			},
		}},
	})
}
