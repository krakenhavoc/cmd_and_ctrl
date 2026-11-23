package cardsr

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ravenloft Adventurer — {3}{B} Creature — Human Rogue Assassin 3/4
//
//	"When this creature enters, you take the initiative.
//	 If a creature an opponent controls would die, instead exile it and put
//	 a hit counter on it.
//	 Whenever this creature attacks, if you've completed a dungeon, defending
//	 player loses 1 life for each card they own in exile with a hit counter
//	 on it."
//
// The replacement is Dauthi Voidwalker's shape (GraveyardBecomesExile's
// ExileCounter): the creature never dies (CR 700.4), and the counter is put
// on the card as it lands in exile. A token is exiled and ceases to exist.
// The attack trigger's intervening "if" (CR 603.4) is read as it triggers
// and again as it resolves; the defending player is fixed as the attack is
// declared (CR 506.3) and the cards are counted as the ability resolves.
func init() {
	attack := effects.IfVentureCond(effects.CondYouCompletedADungeon, game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventAttack},
		AppliesTo: effects.ThisAttacked,
		Key:       "Ravenloft Adventurer — defending player loses 1 life for each card they own in exile with a hit counter on it",
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, "Ravenloft Adventurer — defending player loses 1 life for each card they own in exile with a hit counter on it")
			item.Params.Player = effects.B17DefendingPlayer(g, ev)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			defender := item.Params.Player
			if defender == uuid.Nil {
				return nil
			}
			n := effects.GapB213HitCountersOwnedBy(g, defender)
			if n <= 0 {
				return nil
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -n)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "1822f924-51ba-4f81-92cc-6c5d741a33f9",
		Name:         "Ravenloft Adventurer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Ravenloft Adventurer"),
			attack,
		},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.RfOpponentCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				ev.ExileCounter = "hit"
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Ravenloft Adventurer: exile instead of dying, with a hit counter",
		}},
	})
}
