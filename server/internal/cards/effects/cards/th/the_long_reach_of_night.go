package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Long Reach of Night // Animus of Night's Reach — a transforming
// Saga (Kamigawa: Neon Dynasty, #5493).
//
//	I, II — Each opponent sacrifices a creature of their choice unless
//	        they discard a card.
//	III   — Exile this Saga, then return it to the battlefield
//	        transformed under your control.
//
// Animus of Night's Reach, Enchantment Creature — Spirit 0/4: "Menace.
// Whenever this creature attacks, it gets +X/+0 until end of turn, where
// X is the number of creature cards in defending player's graveyard."
//
// Each opponent, in turn order from the controller's left, is asked
// whether to discard a card; whoever doesn't (or can't, with an empty
// hand) sacrifices a creature they choose. X is counted as the trigger
// resolves, in the graveyard of the player the Animus attacked (a
// planeswalker's or battle's defender too, CR 508.5).
func init() {
	chapter := func(n int) game.TriggeredAbility {
		return effects.ChapterTrigger(n, effects.SagaChapterLabel("The Long Reach of Night", n, "each opponent sacrifices a creature of their choice unless they discard a card"),
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var victims []uuid.UUID
				for _, id := range effects.ApnapPlayers(g) {
					if id != item.Controller {
						victims = append(victims, id)
					}
				}
				return longReachOfNightStep(ctx, victims)
			})
	}
	effects.Register(effects.Spec{
		OracleID:     TheLongReachOfNightOracleID,
		Name:         "The Long Reach of Night",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(chapter(1), game.Purpose{Answers: game.AnswerRemove}),
			effects.TriggerWithPurpose(chapter(2), game.Purpose{Answers: game.AnswerRemove}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("The Long Reach of Night", 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        TheLongReachOfNightOracleID + "#1",
		Name:            "Animus of Night's Reach",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Animus of Night's Reach — it gets +X/+0 until end of turn, where X is the number of creature cards in defending player's graveyard",
				func(g *game.Game, item *game.StackItem) error {
					defender := effects.B17DefendingPlayer(g, item.Trigger.Event)
					x := len(effects.B42CardsInGraveyardOf(g, defender, effects.IsCreatureCard))
					if x == 0 || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  x,
						Label:  "Animus of Night's Reach — +X/+0",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})

	longReachOfNightAnswer = effects.OptionPickThen("option-pick/s82-neo-b18-long-reach-of-night",
		func(ctx *effects.Context, r game.OptionPicked) error {
			if len(r.Carry) == 0 {
				return nil
			}
			victim, rest := r.Carry[0], r.Carry[1:]
			if r.Index == 1 {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   victim,
					Source:   ctx.Source(),
					N:        1,
					Question: "The Long Reach of Night — discard a card",
				})
			} else {
				longReachOfNightSacrifice(ctx, victim)
			}
			return longReachOfNightStep(ctx, rest)
		})
}

// theLongReachOfNightOracleID is shared with the back face ("#1").
const TheLongReachOfNightOracleID = "8b1e453a-35a6-4dff-bd37-498f34d0470e"

var longReachOfNightAnswer game.OptionPickThen

func longReachOfNightSacrifice(ctx *effects.Context, victim uuid.UUID) {
	ctx.Game.PlayerSacrificesForEffect(ctx.Source(), victim,
		effects.SacrificeSpec("a creature", effects.Creature()), "Sacrifice a creature")
}

// longReachOfNightStep asks the next opponent: sacrifice a creature, or
// discard a card. An empty hand can't discard, so that player sacrifices.
func longReachOfNightStep(ctx *effects.Context, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim, rest := remaining[0], remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(victim)
		if p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		if p.Hand == nil || p.Hand.Size() == 0 {
			longReachOfNightSacrifice(ctx, victim)
			remaining = rest
			continue
		}
		return effects.PickOption{
			Player:   victim,
			Question: "The Long Reach of Night — sacrifice a creature, or discard a card?",
			Options: []game.ChoiceOption{
				{Label: "Sacrifice a creature"},
				{Label: "Discard a card"},
			},
			ThenKey: longReachOfNightAnswer,
			Carry:   append([]uuid.UUID{victim}, rest...),
		}.Apply(ctx)
	}
	return nil
}
