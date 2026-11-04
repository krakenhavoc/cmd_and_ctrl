package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Killian, Decisive Mentor — Legendary Creature — Human Warlock
// {1}{W}{B}, 2/3:
//
//	"Whenever an enchantment you control enters, tap up to one target
//	 creature and goad it. (Until your next turn, that creature attacks
//	 each combat if able and attacks a player other than you if able.)
//	 Whenever one or more creatures that are enchanted by an Aura you
//	 control attack, draw a card."
//
// The first trigger taps the target and adds Killian's controller's goad
// to it (CR 701.15a), which the engine ends as that player's next turn
// begins; with no target picked it does nothing. The second is one
// trigger per attack declaration however many such creatures attack
// (CR 603.2c), and counts any player's creature enchanted by an Aura
// Killian's controller controls, whoever it attacks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c54bd34-667d-415c-ad26-82e0eb0ff409",
		Name:         "Killian, Decisive Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, EnchantmentEnteredUnderYourControl,
				"Killian, Decisive Mentor — tap up to one target creature and goad it", killianDecisiveMentorTapAndGoad),
				TargetCreature("up to one target creature").WithCount(0, 1)),
			OncePerBatch(On(game.EventAttack, killianDecisiveMentorEnchantedAttacker,
				"Killian, Decisive Mentor — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}

// killianDecisiveMentorTapAndGoad taps the still-legal target and goads it.
func killianDecisiveMentorTapAndGoad(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return goadAndScheduleClear(ctx, []uuid.UUID{id})
}

// killianDecisiveMentorEnchantedAttacker is "a creature that is enchanted
// by an Aura you control" attacking.
func killianDecisiveMentorEnchantedAttacker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventAttack && socB7EnchantedByAuraOf(g, ev.CardID, source.Controller)
}
