package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nezumi Shortfang // Stabwhisker the Odious — {1}{B} Creature — Rat
// Rogue 1/1 // Legendary Creature — Rat Shaman 3/3, a flip card (#3216,
// ADR 0151):
//
//	Top:     "{1}{B}, {T}: Target opponent discards a card. Then if
//	          that player has no cards in hand, flip this creature."
//	Flipped: "At the beginning of each opponent's upkeep, that player
//	          loses 1 life for each card fewer than three in their
//	          hand."
//
// The discard is the opponent's own choice (DiscardPrompt), and the
// "then" runs after the card has really left their hand, or at once if
// the hand was already empty (CR 701.9a). Stabwhisker is black and {1}
// {B} (CR 710.1c); its upkeep trigger counts the hand as it resolves.
//
// No simplification.
const NezumiShortfangOracleID = "45c8e860-5ce5-4a1c-aebe-97ddffcb2f74"

func init() {
	effects.Register(effects.Spec{
		OracleID:     NezumiShortfangOracleID,
		Name:         "Nezumi Shortfang",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Target opponent discards a card. Then if that player has no cards in hand, flip this creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					victim := t.ID
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player: victim,
						Source: item.SourceCardID,
						N:      1,
						Asked:  game.NewAskedItem(item),
					}, nezumiShortfangDiscarded, nil)
				}
				return nil
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     NezumiShortfangOracleID + "#1",
		Name:         "Stabwhisker the Odious",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachOpponentsUpkeep("Stabwhisker the Odious — that player loses 1 life for each card fewer than three in their hand",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.NewContext(g, item).Trigger().Event.Actor
					if fewer := 3 - effects.B14HandSize(g, player); player != uuid.Nil && fewer > 0 {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, player, -fewer)
					}
					return nil
				}),
		},
	})
}

// nezumiShortfangDiscarded is "then if that player has no cards in hand, flip
// this creature", keyed (#4253): the ability's item rides the prompt as data,
// and the player who discarded is the one whose hand is read.
var nezumiShortfangDiscarded = game.RegisterDiscardThen("choose-cards/nezumi-shortfang-discarded", func(g *game.Game, r game.DiscardsDone) error {
	ctx, _ := effects.AskedContext(g, r.Item, r.Player, r.Source, nil)
	return effects.FlipThisWhen(g, ctx.Item, func(g *game.Game, _ uuid.UUID) bool {
		return effects.B14HandSize(g, r.Player) == 0
	})
})
