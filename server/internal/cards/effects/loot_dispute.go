package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loot Dispute — Enchantment {3}{R}:
//
//	"When this enchantment enters, you take the initiative and create a
//	 Treasure token.
//	 Whenever you attack the player who has the initiative, create a
//	 Treasure token.
//	 Loud Ruckus — Whenever you complete a dungeon, create a 5/5 red
//	 Dragon creature token with flying."
//
// The ETB is TakeTheInitiative then a Treasure (ADR 0203, #4096).
//
// "Whenever you attack the player who has the initiative" is the
// "whenever you attack a player" batch, Horizon Explorer's shape: one
// declaration that attacks that player with any number of creatures is
// one trigger (OncePerBatchPerPlayer), and the player must have the
// initiative as attackers are declared — it is the trigger condition,
// not an intervening "if", so the Treasure comes even if the initiative
// has moved by the time the trigger resolves.
//
// Loud Ruckus watches EventDungeonCompleted (CR 309.7): completing the
// Undercity, by the state-based action or by venturing out of its last
// room, makes the Dragon.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c885698-69f2-44b5-abfe-9a2833822794",
		Name:         "Loot Dispute",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Loot Dispute — you take the initiative and create a Treasure token",
				Do(TakeTheInitiative{}, CreateToken{Template: TreasureToken(), N: 1})),
			OncePerBatchPerPlayer(On(game.EventAttack, lootDisputeAttackedTheInitiative,
				"Loot Dispute — you attacked the player who has the initiative: create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1}))),
			WheneverYouCompleteADungeon("Loot Dispute — Loud Ruckus: create a 5/5 red Dragon with flying",
				Do(CreateToken{Template: TokenCard("5/5 red Dragon with flying"), N: 1})),
		},
	})
}

// lootDisputeAttackedTheInitiative: a creature you control was declared
// attacking the player who has the initiative.
func lootDisputeAttackedTheInitiative(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	return YouHaveTheInitiative(g, ev.Target)
}
