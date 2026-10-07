package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Suppressor Skyguard — Creature — Human Knight {2}{W}{U}, 2/3:
//
//	"Flying
//	 Whenever a player attacks you, if that player has another opponent
//	 who isn't being attacked, prevent all combat damage that would be
//	 dealt to you this combat."
//
// ADR 0108 amendment 2026-10-07 (#2027): the shield lasts "this combat"
// (game.UntilEndOfCombat), the combat phase the attack was declared in,
// not the turn. The attack declaration is one occurrence however many
// creatures it holds (OncePerBatch, CR 603.2c). A player is "being
// attacked" when a creature is attacking them or a planeswalker they
// control (CR 506.3d), whoever controls the creature. The condition is an
// intervening if (CR 603.4): it is checked as the attack is declared and
// again as the trigger resolves, so a second opponent who is attacked in
// between (or, through combat removal, no longer) is read as it stands.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "06c47664-2b4b-468e-a38d-04b62e6edd67",
		Name:            "Suppressor Skyguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
					ev.Target == source.Controller &&
					anotherOpponentNotBeingAttacked(g, ev.Actor, source.Controller)
			}, "Suppressor Skyguard — prevent all combat damage dealt to you this combat", func(g *game.Game, item *game.StackItem) error {
				// CR 603.4: the intervening if holds on resolution too.
				if !anotherOpponentNotBeingAttacked(g, item.Trigger.Event.Actor, item.Controller) {
					return nil
				}
				return PreventDamageFromSource{
					Protect: ShieldYou, CombatOnly: true, Lasts: ShieldThisCombat,
					Label: "Suppressor Skyguard — prevent all combat damage dealt to you this combat",
				}.Apply(NewContext(g, item))
			})),
		},
	})
}

// anotherOpponentNotBeingAttacked reports whether `attacker` has an
// opponent other than `defender`, still in the game, who is not being
// attacked: no creature is attacking that player or a planeswalker they
// control. In a game with only the two of them there is no such opponent.
func anotherOpponentNotBeingAttacked(g *game.Game, attacker, defender uuid.UUID) bool {
	attacked := map[uuid.UUID]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget != uuid.Nil {
			if p := g.DefendingPlayerForAttackForEffect(c.AttackingTarget); p != uuid.Nil {
				attacked[p] = true
			}
		}
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == attacker || p.ID == defender {
			continue
		}
		if !attacked[p.ID] {
			return true
		}
	}
	return false
}
