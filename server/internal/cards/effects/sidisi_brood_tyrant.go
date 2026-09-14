package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// b33SidisiZombieLabel is the stack label of Sidisi's mill trigger —
// the "one or more" dedup keys on it.
const b33SidisiZombieLabel = "Sidisi, Brood Tyrant — create a 2/2 black Zombie"

// Sidisi, Brood Tyrant — Legendary Creature — Snake Shaman
// {1}{B}{G}{U}, 3/3 (EDHREC rank 3524):
//
//	"Whenever Sidisi enters or attacks, mill three cards.
//	 Whenever one or more creature cards are put into your graveyard
//	 from your library, create a 2/2 black Zombie creature token."
//
// The self-mill commander. The first trigger is Sun Titan's "enters
// or attacks" shape with a three-card mill; the second is Colossal
// Grave-Reaver's: it watches EventMill — the engine emits one per
// card — so the first creature card of a mill fires it and the rest
// of that mill are declined while the trigger is queued or on the
// stack. Two separate mills in one turn make two Zombies, as
// printed. A creature card that reaches the graveyard from the
// library some other way — a search that puts it there — is not a
// mill and does not fire it; no card in the catalog does that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fad7072-21e3-446e-a28f-615038c8bfea",
		Name:         "Sidisi, Brood Tyrant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB, game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b21SelfEnteredOrAttacked(ev, source)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Sidisi, Brood Tyrant — mill three cards", b33MillN(3))
				},
			},
			{
				Watches: []game.EventKind{game.EventMill},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b33CreatureCardMilledIntoYourGraveyard(ev, source, g) && !b12TriggerPendingOrOnStack(g, source, b33SidisiZombieLabel)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b33SidisiZombieLabel, b33CreateTokenBody(BlackZombieToken, 1))
				},
			},
		},
	})
}
