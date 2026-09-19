package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of Fire and Ice — Artifact — Equipment for {3}:
//
//	"Equipped creature gets +2/+2 and has protection from red and
//	 from blue.
//	 Whenever equipped creature deals combat damage to a player, this
//	 Equipment deals 2 damage to any target and you draw a card.
//	 Equip {2}"
//
// The other half of the Sword cycle's flagship pair, and the reason
// it is in this batch alongside Feast and Famine: it is the first
// TARGETED trigger on an attachment. The 2 damage is chosen as the
// ability goes on the stack (CR 603.3d) — the engine computes the
// legal set at trigger time, drops the trigger silently if there is
// nothing to point at, and re-checks at resolution — so the whole
// "kill the blocker that just traded" line works without the card
// file knowing any of it.
//
// Note the damage source is the EQUIPMENT, not the creature. That is
// the printed text and it matters for the damage attribution in the
// event log; NewTriggeredItem already binds the source correctly.
//
// PROTECTION FROM RED AND FROM BLUE is a layer-6 grant to the
// equipped creature (GrantToAttached), and #662 made it real: the
// tokens are read by game.ProtectionQualities and enforced at all
// four DEBT points. It is the Sword's evasion (a red or blue blocker
// cannot block the wearer, CR 702.16f) and its removal protection (a
// red or blue spell cannot target it and red or blue damage is
// prevented, CR 702.16b and e).
//
// Note the grant is on the EQUIPPED CREATURE and the Sword itself is
// a colourless artifact, which is the distinction CR 702.16b turns
// on: the Sword's own damage trigger below is a source with no
// colour, so it can point its 2 damage at a creature with protection
// from red quite happily. That asymmetry — an ability's source is the
// permanent, not its controller and not the creature it is attached
// to — is exactly what the TargetSource refactor bought.
func init() {
	Register(Spec{
		OracleID:     "2ccdc60a-49a9-44b9-a7af-0ebf18b26785",
		Name:         "Sword of Fire and Ice",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("protection from red", "protection from blue"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Targets: TargetAny(),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Sword of Fire and Ice — 2 damage, draw a card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if err := (DealDamage{Target: t.ID, Amount: 2, Source: item.SourceCardID}.Apply(ctx)); err != nil {
								return err
							}
						}
						// The draw is not conditional on the damage
						// landing — "and you draw a card" is a
						// separate clause, so a fizzled target still
						// draws.
						return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
					})
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
