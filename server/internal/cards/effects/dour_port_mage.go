package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dour Port-Mage — Creature — Frog Wizard {1}{U}, 1/3 (EDHREC rank
// 1511):
//
//	"Whenever one or more other creatures you control leave the
//	 battlefield without dying, draw a card.
//	 {1}{U}, {T}: Return another target creature you control to its
//	 owner's hand."
//
// The blink deck's card-draw engine, with its own bounce built in. A
// flicker, a bounce, an exile of one of the controller's other
// creatures draws a card; a death does not (CR 700.4 — EventLTB
// carries the destination). The Port-Mage's own ability is the
// simplest way to fire it.
//
// "One or more": the engine emits one LTB event per creature, so a
// mass bounce would fire once per creature; the later events of the
// SAME event batch are declined — the b04-style dedup, narrowed to
// this ability's label so the Port-Mage's activated item is never
// mistaken for it. Without the dedup a Cyclonic Rift on the
// controller's own board would draw a hand, which is the #259
// direction. A separate, LATER batch triggers again, whatever is
// still on the stack from the last one: bounce, then bounce again in
// response, and the Port-Mage draws twice (#829, CR 603.2c).
//
// The bounce is the printed target clause: another creature the
// controller controls. "Another" is Noxious Gearhulk's shape — a
// target clause cannot see its own source, so the Port-Mage is
// excluded by name at announce (singleton format: the same creature)
// and by instance at resolution. Summoning sickness applies to the
// {T}.
//
// No simplification.
func init() {
	const label = "Dour Port-Mage — draw a card"
	Register(Spec{
		OracleID:     "cf58e309-00e8-438e-813e-2e1c1002db23",
		Name:         "Dour Port-Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13OtherCreatureYouControlLeftWithoutDying(ev, source, g)
			}, label, Do(DrawCards{N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Return another target creature you control to its owner's hand.",
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: TargetCreature("another target creature you control", YouControl(), b03NotNamed("Dour Port-Mage")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				if item.Targets[0].ID == item.SourceCardID {
					return nil
				}
				return BounceToHand{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
