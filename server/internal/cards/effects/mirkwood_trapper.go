package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirkwood Trapper — Creature — Elf Scout {1}{G}{U}, 1/4 (Tales of
// Middle-earth Commander):
//
//	"Whenever a player attacks you, target attacking creature gets -2/-0
//	 until end of turn.
//	 Whenever a player attacks, if they aren't attacking you, that player
//	 chooses an attacking creature. It gets +2/+0 until end of turn."
//
// The first clause is Archnemesis's "a player attacks you": one trigger
// per declaration at the controller (the batch is keyed by the attacked
// player), its target any attacking creature, whoever controls it.
//
// The second is one trigger per declaration of any player, with an
// intervening "if" (CR 603.4) checked as it triggers and again as it
// resolves: none of that player's creatures is attacking the Trapper's
// controller. That player (not the Trapper's controller) chooses the
// creature, any attacking creature at all, and the choice is made as the
// ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "debbbdb2-2fa8-413f-b394-977212915c65",
		Name:         "Mirkwood Trapper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OncePerBatchPerPlayer(On(game.EventAttack, msc18APlayerAttacksYou,
				"Mirkwood Trapper — target attacking creature gets -2/-0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return BoostUntilEOT{Target: ts[0].ID, Power: -2, Label: "Mirkwood Trapper — -2/-0"}.Apply(ctx)
				})),
				TargetCreature("target attacking creature", AttackingCreature())),
			InterveningIf(s82GapB117TrapperNotAttackedCond,
				OncePerBatch(On(game.EventAttack, s82GapB117TrapperAttackerNotAttackingYou,
					"Mirkwood Trapper — that player chooses an attacking creature; it gets +2/+0 until end of turn",
					Do(Instruction(s82GapB117TrapperChoose))))),
		},
	})
}

// s82GapB117AttackingYou reports whether any creature `attacker`
// controls is attacking the player `you` (not a planeswalker or battle
// of theirs).
func s82GapB117AttackingYou(g *game.Game, attacker, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == attacker && c.IsCreature() && c.AttackingTarget == you {
			return true
		}
	}
	return false
}

func s82GapB117TrapperAttackerNotAttackingYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventAttack || ev.Actor == uuid.Nil {
		return false
	}
	return !s82GapB117AttackingYou(g, ev.Actor, source.Controller)
}

var s82GapB117TrapperNotAttackedCond = CondBoard{
	Name: "the attacking player isn't attacking you",
	Test: func(ctx *Context) bool {
		attacker := ctx.Trigger().Event.Actor
		return attacker != uuid.Nil && !s82GapB117AttackingYou(ctx.Game, attacker, ctx.Controller())
	},
}

func s82GapB117TrapperChoose(ctx *Context) error {
	return ChoosePermanents{
		Player:   ctx.Trigger().Event.Actor,
		Question: "Mirkwood Trapper — choose an attacking creature; it gets +2/+0 until end of turn",
		Candidates: func(g *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
			return AttackingCreatures(g), 1, 1
		},
		Keyed: s82GapB117TrapperPump,
	}.Apply(ctx)
}

var s82GapB117TrapperPump = RegisterPermanentsPick("prompt-run/mirkwood-trapper-pump", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (BoostUntilEOT{Target: id, Power: 2, Label: "Mirkwood Trapper — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
