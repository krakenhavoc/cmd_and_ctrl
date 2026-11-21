package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wreck Hunter — Creature — Human Artificer {B}{B}, 2/2 (The Brothers'
// War):
//
//	"Flash
//	 When this creature enters, choose target player. You create a tapped
//	 Powerstone token for each nonland card in that player's graveyard
//	 that was put there from the battlefield this turn."
//
// The count is made as the trigger resolves, off the target player's
// graveyard: each card's newest arrival must be this turn and from the
// battlefield (putIntoGraveyardFromBattlefieldThisTurn, Cry of the
// Carnarium's read). A token is not a card (CR 108.2) and a land card is
// excluded; "nonland" is read off the card as it sits in the graveyard.
// The Powerstones are the Hunter's controller's, whoever the target is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "08dcda8e-704d-444e-a7ca-34586a7f42a7",
		Name:            "Wreck Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Targets:   effects.TargetPlayer("target player"),
			Key:       "Wreck Hunter — create a tapped Powerstone token for each nonland card put into that player's graveyard from the battlefield this turn",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var who uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						who = t.ID
					}
				}
				n := 0
				for _, p := range g.Seats {
					if p == nil || p.ID != who || p.Graveyard == nil {
						continue
					}
					for _, c := range p.Graveyard.Cards {
						if !c.IsLand() && !c.IsToken() && effects.PutIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID) {
							n++
						}
					}
				}
				if n == 0 {
					return nil
				}
				return effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: n}.Apply(ctx)
			},
		}},
	})
}
