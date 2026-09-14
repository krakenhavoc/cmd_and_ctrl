package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// On Wings of Gold — Enchantment {3}{W} (EDHREC rank 3609):
//
//	"Creatures you control that are Zombies and/or tokens get +1/+1
//	 and have flying.
//	 Whenever one or more cards leave your graveyard, create a 1/1
//	 white Zombie creature token."
//
// The Aetherdrift Zombie anthem with Tormod's trigger. The static
// is two layer effects over one scope — a creature the controller
// controls that is a Zombie, a token, or both (a Zombie token counts
// once) — +1/+1 in layer 7c and flying in layer 6; the Zombie it
// makes is both and lifts itself. The trigger is Tormod, the
// Desecrator's condition (b16CardLeftYourGraveyard — a move out of
// the controller's own graveyard, or a flashback cast from it) with
// the per-label "one or more" dedup, so a Bojuka Bog on the
// graveyard makes one Zombie and a single regrowth makes one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b6f608d-f27d-44cd-b0a1-0e1e1aa1c98a",
		Name:         "On Wings of Gold",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b34ZombiesAndTokensYouControlGetPlusOne(),
			b34ZombiesAndTokensYouControlHaveFlying(),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventZoneMove, game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16CardLeftYourGraveyard(ev, source, g) && !b12TriggerPendingOrOnStack(g, source, b34OnWingsOfGoldLabel)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, b34OnWingsOfGoldLabel, b34CreateTokens(b34WhiteZombieToken, 1))
			},
		}},
	})
}

// b34OnWingsOfGoldLabel is the stack label of On Wings of Gold's
// graveyard trigger — the "one or more" dedup keys on it.
const b34OnWingsOfGoldLabel = "On Wings of Gold — create a 1/1 white Zombie"
