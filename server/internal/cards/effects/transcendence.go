package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Transcendence — Enchantment {3}{W}{W}{W}:
//
//	"You don't lose the game for having 0 or less life.
//	 When you have 20 or more life, you lose the game.
//	 Whenever you lose life, you gain 2 life for each 1 life you lost.
//	 (Damage dealt to you causes you to lose life.)"
//
// ADR 0107 PR 1 (#1858). Three lines, three existing shapes:
//
//   - The first is a "can't lose" gate narrowed to one cause, CR
//     704.5a's 0 or less life (ADR 0057, LossLife). Poison, an empty
//     library, commander damage and "you lose the game" effects still
//     end the game.
//   - The second is a CR 603.8 state trigger over its controller's life
//     total. The loss is an effect's (LossEffect), so the gate above does
//     not stop it. It triggers once and, if something stops the loss,
//     again after it has left the stack.
//   - The third watches both shapes a life loss takes (damage, and a
//     loss or payment), Vilis's s22PlayerLostLife, and carries the amount
//     lost on the item as it triggers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9279f8f-3d19-4f5b-b194-cb8c65313b7e",
		Name:         "Transcendence",
		Completeness: CompletenessFull,
		GameEndGates: []game.GameEndGate{{Scope: game.GateYou, CantLose: true, Causes: []game.LossCause{game.LossLife}}},
		Triggered: []game.TriggeredAbility{
			WhenState("Transcendence — you lose the game",
				func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
					p := g.PlayerByIDForEffect(controller)
					return p != nil && !p.Eliminated && p.Life >= 20
				},
				Do(LoseTheGame{})),
			{
				Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
				Key:     "Transcendence — gain 2 life for each 1 life you lost",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := s22PlayerLostLife(ev, source.Controller, g)
					return ok
				},
				// A fill-in Build (ADR 0041 P9): the amount lost is a fact
				// of the moment the ability triggered.
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Transcendence — gain 2 life for each 1 life you lost")
					item.Params.Amount, _ = s22PlayerLostLife(ev, source.Controller, g)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 2 * item.Params.Amount}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
