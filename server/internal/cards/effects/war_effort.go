package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War Effort — Enchantment {3}{R} (Tarkir: Dragonstorm, #4463):
//
//	"Creatures you control get +1/+0.
//	 Whenever you attack, create a 1/1 red Warrior creature token that's
//	 tapped and attacking. Sacrifice it at the beginning of the next end
//	 step."
//
// The anthem is a layer 7c static over the creatures its controller controls
// (b16Anthem), Warriors included. "Whenever you attack" is ONE trigger per
// attack declaration (OncePerBatch, Adeline's shape), for the controller's
// own attack. The Warrior is Dalkovan Encampment's token: it enters tapped and
// attacking, the controller choosing what it attacks (CR 508.4,
// CreateTokensAttackingYourChoice, the first declared attacker's target
// preferred), and was never declared, so "whenever a creature attacks"
// triggers do not fire for it. Its ID goes to a delayed trigger that
// sacrifices it at the beginning of the next end step (CR 603.7), and one
// that has already left is not sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9bc6841-9f71-46af-ac1f-a01caee290c1",
		Name:         "War Effort",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static:       []game.StaticAbility{b16Anthem(b16CreaturesYouControl, 1, 0)},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "War Effort — create a 1/1 red Warrior token tapped and attacking; sacrifice it at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Template: b33TappedAttackingRedWarrior(),
						N:        1,
						Tapped:   true,
						Prefer:   triggerAttackTarget(item),
						Label:    "War Effort",
						Tail:     game.KeyedTail{Ref: warEffortWarriorSacrifice},
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// warEffortWarriorSacrifice is the Warrior's tail, keyed: the end step
// trigger that sacrifices it. An on-disk key.
var warEffortWarriorSacrifice = AttackingTokensThen("attacking-tokens/war-effort-sacrifice-the-warrior",
	func(ctx *Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label: "War Effort — sacrifice the Warrior",
			Cards: created,
			Body:  sacrificeListedCardsBody,
		}.Apply(ctx)
	})
