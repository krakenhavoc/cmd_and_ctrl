package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outpost Siege — Enchantment {3}{R}:
//
//	"As this enchantment enters, choose Khans or Dragons.
//	 • Khans — At the beginning of your upkeep, exile the top card of
//	   your library. Until end of turn, you may play that card.
//	 • Dragons — Whenever a creature you control leaves the
//	   battlefield, this enchantment deals 1 damage to any target."
//
// A #1572 anchor-word card: each bullet is a trigger gated on its word
// (ADR 0071).
//
//   - Khans is Ragavan's impulse exile with the PLAY grant, so an
//     exiled land can be played (it says play, not cast).
//   - Dragons is any exit from the battlefield — dying, bounce, exile —
//     judged on the creature as it last existed on the battlefield (its
//     controller survives the move), and it fires once per creature.
//     The damage source is the Siege.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebb24fc7-dc71-4712-8c2a-b5920f78e55d",
		Name:         "Outpost Siege",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Outpost Siege", "Khans", "Dragons"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Khans", AtYourUpkeep("Outpost Siege — exile the top card of your library; you may play it this turn",
				outpostSiegeImpulse)),
			WhenChosen("Dragons", Targeting(
				On(game.EventLTB, aCreatureYouControlLeft, "Outpost Siege — 1 damage to any target",
					b33DamageChosenTargetFromSource(1)),
				TargetAny())),
		},
	})
}

// aCreatureYouControlLeft is "whenever a creature you control leaves
// the battlefield": any destination, the creature read after the move
// (the move does not reset its controller — see edea_possessed_sorceress.go).
func aCreatureYouControlLeft(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB {
		return false
	}
	left, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftAsType(ev, left, "creature") && left.Controller == source.Controller
}

// outpostSiegeImpulse is the Khans body.
func outpostSiegeImpulse(g *game.Game, item *game.StackItem) error {
	return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(NewContext(g, item))
}
