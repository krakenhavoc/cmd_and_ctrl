package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kalitas, Traitor of Ghet — Legendary Creature — Vampire Warrior
// {2}{B}{B}, 3/4:
//
//	"Lifelink
//	 If a nontoken creature an opponent controls would die, instead
//	 exile that card and create a 2/2 black Zombie creature token.
//	 {2}{B}, Sacrifice another Vampire or Zombie: Put two +1/+1
//	 counters on Kalitas."
//
// The replacement is The Darkness Crystal's: "if a nontoken creature
// an opponent controls would die, instead exile it"
// (opponentNontokenCreatureWouldDie). The creature never died (CR 700.4),
// so its dies-triggers do not fire. A commander is exiled too and its
// owner then decides about the command zone (CR 903.9a).
//
// "And create a 2/2 black Zombie creature token" is the second
// instruction of the modified event (CR 608.2c, 614.6; #4522), not a
// trigger: nothing goes on the stack between the exile and the Zombie.
// It is declared with CreateTokensAfterReplacedMoveForEffect, lands
// after the exile through the token pipeline (a doubler doubles it), and
// whatever the destroying effect does next waits for it, including a
// CR 616 question between two token replacements. The Zombie is
// Kalitas' controller's, fixed as the replacement applies, so Kalitas
// leaving afterwards changes nothing. The Zombie is still made when the
// card ends up somewhere else (a commander moved to the command zone):
// only the impossible instruction is ignored.
//
// The sacrifice is a cost of "another" Vampire or Zombie you control,
// so Kalitas cannot pay with itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e1cfd1cb-44a5-429f-a5c1-e6d29bad1c71",
		Name:            "Kalitas, Traitor of Ghet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: opponentNontokenCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return g.CreateTokensAfterReplacedMoveForEffect(ev, game.TokenCreation{
					Controller: src.Controller,
					Source:     src.InstanceID,
					Groups: []game.TokenGroup{{
						Template: src.TokenThroughText(TokenCard("2/2 black Zombie")),
						Count:    1,
					}},
				})
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Kalitas, Traitor of Ghet: exile instead of dying, and create a 2/2 black Zombie",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice another Vampire or Zombie: Put two +1/+1 counters on Kalitas.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{2}{B}"), SacrificeAnotherN(1, "another Vampire or Zombie",
				Or(HasSubtype("Vampire"), HasSubtype("Zombie")))),
			Effect: Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
