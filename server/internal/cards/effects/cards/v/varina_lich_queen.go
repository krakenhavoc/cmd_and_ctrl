package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Varina, Lich Queen — Legendary Creature — Zombie Wizard {1}{W}{U}{B}, 4/4
// (Commander 2018, #5526):
//
//	"Whenever you attack with one or more Zombies, draw that many cards,
//	 then discard that many cards. You gain that much life.
//	 {2}, Exile two cards from your graveyard: Create a tapped 2/2 black
//	 Zombie creature token."
//
// The attack trigger is Screaming Swarm's counted batch (ThatManyEvents):
// one trigger per declaration, its amount the Zombies you declared as
// attackers (Sidar Jabari's youAttackWithAKnight shape, read for Zombie
// through the card's text). Varina is a Zombie, so her own attack counts.
// The draw and the discard are one loot of that many, the discard prompt
// opening only after every draw; the life gain is the same number.
//
// The activated ability is Grim Lavamancer's graveyard exile cost with
// Altar of Bhaal's tapped token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9859fbca-60c3-4a57-a377-51b56f03894f",
		Name:         "Varina, Lich Queen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && effects.AttackerHasSubtype(g, ev, source.TextSubtype("Zombie"))
			}, "Varina, Lich Queen — draw that many cards, then discard that many cards; gain that much life",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.SettledEventAmount(item)
					if err := effects.LootOne(g, item, n); err != nil {
						return err
					}
					return effects.GainLife{Amount: n}.Apply(effects.NewContext(g, item))
				}),
				effects.ThatManyEvents()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Exile two cards from your graveyard: Create a tapped 2/2 black Zombie creature token.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ExileFromGraveyard(2, "two cards", nil)),
			Effect:  effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.BlackZombieToken()).EntersTapped(), N: 1}),
		}},
	})
}
