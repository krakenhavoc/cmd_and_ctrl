package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duchess, Wayward Tavernkeep — Legendary Creature — Human Citizen {3}{R},
// 4/3 (Fallout, #5948):
//
//	"Hunters for Hire — Whenever a creature you control deals combat damage
//	 to a player, put a quest counter on it.
//	 {1}, Remove a quest counter from a permanent you control: Create a Junk
//	 token. (It's an artifact with "{T}, Sacrifice this token: Exile the top
//	 card of your library. You may play that card this turn. Activate only as
//	 a sorcery.")"
//
// The trigger is one per creature and per player it hit; the counter goes on
// only if the creature is still the same object on the battlefield as the
// ability resolves (CR 400.7). The activated ability's cost removes the
// counter from any permanent its controller controls (RemoveCountersFrom,
// chosen, not targeted) before the ability goes on the stack, so removing
// the last counter from a creature already in combat is an ordinary payment.
// The token is the printed Junk (JunkToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b47d91a1-f2ac-4604-95a7-c247ac2fab18",
		Name:         "Duchess, Wayward Tavernkeep",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, gapB63CreatureYouControlDealtCombatDamageToAPlayer,
				"Duchess, Wayward Tavernkeep — put a quest counter on it",
				func(g *game.Game, item *game.StackItem) error {
					id, ok := gapB63DamageDealer(g, item)
					if !ok {
						return nil
					}
					return AddCounter{Target: id, Kind: "quest", N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Remove a quest counter from a permanent you control: Create a Junk token.",
			Cost:    Plus(ManaCost("{1}"), RemoveCountersFrom("quest", 1, "a permanent you control")),
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Effect:  Do(CreateToken{Template: JunkToken(), N: 1}),
		}},
	})
}
