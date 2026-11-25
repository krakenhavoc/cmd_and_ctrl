package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reaper of Night // Harvest Fear — an adventure card (Throne of Eldraine):
//
//	Reaper of Night {5}{B}{B}, Creature — Specter 4/5:
//	"Whenever this creature attacks, if defending player has two or fewer
//	 cards in hand, it gains flying until end of turn."
//
//	Harvest Fear {3}{B}, Sorcery — Adventure:
//	"Target opponent discards two cards. (Then exile this card. You may cast
//	 the creature later from exile.)"
//
// The adventure layout is Brazen Borrower's: the creature under the bare
// oracle ID and the Adventure half under "<oracle_id>#1" (CR 715.3d). The
// attack trigger is Spectral Bears' intervening "if" (CR 603.4): the
// defending player is fixed as it triggers and carried on the item, and their
// hand size is checked as it triggers and again as it resolves. Harvest Fear
// is Mire's Malice's discard: the opponent chooses two cards of their own.
//
// No simplification.
func init() {
	const oracleID = "d40e5a76-6c9a-47eb-b412-d3056cf36b3f"
	const label = "Reaper of Night — it gains flying until end of turn"
	t := effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.AttackDeclared(ev, source) && reaperDefenderHasTwoOrFewer(g, g.DefendingPlayerForAttackForEffect(ev.Target))
	}, label, func(g *game.Game, item *game.StackItem) error {
		if !reaperDefenderHasTwoOrFewer(g, item.Params.Player) {
			return nil
		}
		return effects.ThisGainsKeywordUntilEndOfTurn("flying", label)(g, item)
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Player = g.DefendingPlayerForAttackForEffect(ev.Target)
		return item
	}
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Reaper of Night",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Harvest Fear",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			p, ok := ctx.ClauseTarget(0)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   p.ID,
				Source:   item.SourceCardID,
				N:        2,
				Question: "Harvest Fear — discard two cards",
			})
			return nil
		},
	})
}

// reaperDefenderHasTwoOrFewer is "defending player has two or fewer cards in
// hand". False for no player: an attack aimed at nobody has no defender.
func reaperDefenderHasTwoOrFewer(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && p.Hand != nil && p.Hand.Size() <= 2
}
