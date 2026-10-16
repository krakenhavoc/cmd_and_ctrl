package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veyran, Voice of Duality — Legendary Creature — Efreet Wizard
// {1}{U}{R}, 2/2 (EDHREC rank 745):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 Veyran gets +1/+1 until end of turn.
//	 If you casting or copying an instant or sorcery spell causes a
//	 triggered ability of a permanent you control to trigger, that
//	 ability triggers an additional time."
//
// The pump is the shared Magecraft trigger (CR 207.2c): EventCast for
// a cast, EventSpellCopied for a copy put on the stack.
//
// The second paragraph is a CR 603.2d trigger doubler (#752) whose
// cause is those same two events: the actor is Veyran's controller and
// the spell is an instant or a sorcery. The doubled ability must
// belong to a PERMANENT that player controls, so a spell's own "when
// you cast this spell" trigger is not doubled (FromSpell), nor is an
// opponent's permanent watching your cast. Veyran's own magecraft is a
// permanent's triggered ability and is doubled too, as its rulings say.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d57a6b19-bd06-4b07-8ee1-0b45612ba9e4",
		Name:         "Veyran, Voice of Duality",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Magecraft("Veyran, Voice of Duality — gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Veyran, Voice of Duality — magecraft pump")),
		},
		TriggerDoublers: []game.TriggerDoubler{{
			Label:    "Veyran, Voice of Duality",
			RowLabel: "Abilities of your permanents that trigger when you cast or copy an instant or sorcery spell trigger an additional time.",
			Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
				if q.FromSpell || q.SourceLKI.Controller != q.DoublerLKI.Controller {
					return false
				}
				doubler := q.Doubler
				doubler.Controller = q.DoublerLKI.Controller
				return instantOrSorceryCastOrCopiedByYou(q.Event, &doubler, g)
			},
		}},
	})
}
