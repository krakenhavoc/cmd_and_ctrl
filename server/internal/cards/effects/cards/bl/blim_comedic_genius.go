package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blim, Comedic Genius — Legendary Creature — Imp {2}{B}{R}, 4/3 (Commander
// Legends):
//
//	"Flying
//	 Whenever Blim deals combat damage to a player, that player gains control
//	 of target permanent you control. Then each player loses life and discards
//	 cards equal to the number of permanents they control but don't own."
//
// The permanent is targeted as the trigger goes on the stack and handed to the
// damaged player for good (CR 611.2a) if it is still legal as the trigger
// resolves. The loss is then counted per player, in turn order from the
// controller, after the control change (CR 608.2c): a permanent a player
// controls and does not own, whoever took it. Each player loses that much life
// and then chooses and discards that many cards (CR 701.9a); a player with
// none loses and discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7cf9574-e2cf-4137-a224-73f76c5d11d7",
		Name:            "Blim, Comedic Genius",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Blim, Comedic Genius — that player gains control of target permanent you control; then each player loses life and discards cards equal to the number of permanents they control but don't own",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Trigger == nil {
						return nil
					}
					damaged := item.Trigger.Event.Target
					if gift, ok := effects.B16FirstLegalTargetCard(ctx); ok {
						if err := (effects.GainControl{
							Target:     gift,
							Controller: damaged,
							Duration:   game.IndefiniteDuration(),
							Label:      "Blim, Comedic Genius — that player gains control of it",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					g.RecomputeLayersIfStaleLocked()
					notOwned := map[uuid.UUID]int{}
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != c.Owner {
							notOwned[c.Controller]++
						}
					}
					for _, id := range effects.TablePlayers(ctx) {
						n := notOwned[id]
						if n <= 0 {
							continue
						}
						if err := g.ChangePlayerLifeForEffect(item.SourceCardID, id, -n); err != nil {
							return err
						}
						g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
							Player:   id,
							Source:   item.SourceCardID,
							N:        n,
							Question: "Blim, Comedic Genius — discard a card for each permanent you control but don't own",
						})
					}
					return nil
				}),
				effects.TargetPermanent("target permanent you control", effects.YouControl())),
		},
	})
}
