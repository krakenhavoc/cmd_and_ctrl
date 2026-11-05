package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Combadge — Artifact — Communicator Equipment {3} (Star Trek, #4973):
//
//	"When you cast this spell, copy it. (The copy becomes a token.)
//	 Equipped creature has "Whenever this creature attacks while you
//	 control another creature with a Communicator attached to it, this
//	 creature can't be blocked this turn."
//	 Equip {2}"
//
// The cast trigger is Lumaret's Favor's (WhenYouCastThisSpell) with the
// copy unconditional: CR 707.10f makes the copy of a permanent spell a
// token as it resolves, a Combadge token under the caster's control. The
// copy names the spell without targeting it, so a Combadge countered in
// response is still copied from last-known information (CR 608.2h,
// CopySpell.FromLastKnown, storm's reading).
//
// The granted trigger is the equipped creature's own (ADR 0093,
// Conformer Shuriken's shape). "While" is a condition of the attack, not
// an intervening "if" (CR 603.4): as the creature is declared an
// attacker, its controller must control another creature that has a
// permanent with the Communicator subtype attached to it, read through
// the layered view (trekB26CommunicatorOnAnotherCreature). Two creatures
// each wearing a Combadge make both unblockable. The restriction ends at
// cleanup.
//
// No simplification.
const combadgeGrant = "combadge/unblockable-with-another-communicator"

func init() {
	Register(Spec{
		OracleID:     "ccd0e613-17c1-4a97-b519-de0376d596f0",
		Name:         "Combadge",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: combadgeGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attackDeclared(ev, source) && trekB26CommunicatorOnAnotherCreature(g, source)
				}, "Combadge — this creature can't be blocked this turn",
					func(g *game.Game, item *game.StackItem) error {
						return RestrictUntilEOT{
							Target:       item.SourceCardID,
							Restrictions: game.CantBeBlocked,
							Label:        "Combadge — can't be blocked",
						}.Apply(NewContext(g, item))
					}),
			},
			Text: "Whenever this creature attacks while you control another creature with a Communicator attached to it, this creature can't be blocked this turn.",
		}},
		Static: []game.StaticAbility{
			GrantAbilitiesToAttached(combadgeGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Combadge — copy this spell",
				func(g *game.Game, item *game.StackItem) error {
					return CopySpell{StackID: item.SourceCardID, Controller: item.Controller, FromLastKnown: true}.Apply(NewContext(g, item))
				}),
		},
	})
}
