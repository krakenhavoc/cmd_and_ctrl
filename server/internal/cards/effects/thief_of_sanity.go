package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thief of Sanity — Creature — Specter {1}{U}{B}, 2/2 (Outlaws of Thunder
// Junction Commander):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, look at the top
//	 three cards of that player's library, exile one of them face down, then
//	 put the rest into their graveyard. You may cast that card for as long as
//	 it remains exiled, and mana of any type can be spent to cast that spell."
//
// Black Cat's look-and-exile (the looker alone knows the cards, the pick is
// keyed, the exile is face down with the any-type grant) with the rest going
// to the damaged player's graveyard. "Cast", not "play": a land exiled this
// way is not offered as a land drop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "68b51d58-7566-4a26-8c31-8da32799bea7",
		Name:            "Thief of Sanity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Thief of Sanity — look at the top three cards of that player's library, exile one face down, put the rest into their graveyard",
				cmdr2S45ThiefOfSanityTrigger),
		},
	})
}

func cmdr2S45ThiefOfSanityTrigger(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return cmdr2S45LookExileOneRestToGraveyard(g, item, item.Trigger.Event.Target, 3, "Thief of Sanity")
}
