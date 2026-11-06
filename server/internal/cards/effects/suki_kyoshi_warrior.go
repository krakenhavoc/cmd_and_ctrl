package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suki, Kyoshi Warrior — Legendary Creature — Human Warrior Ally
// {2}{G/W}{G/W}, */4 (Avatar: The Last Airbender, #4459):
//
//	"Suki's power is equal to the number of creatures you control.
//	 Whenever Suki attacks, create a 1/1 white Ally creature token that's
//	 tapped and attacking."
//
// The power is Ironroot Warlord's characteristic-defining ability (layer
// 7a). The token is put onto the battlefield attacking through
// CreateTokensAttackingYourChoice, General Kreat's path: its controller
// chooses which opponent, planeswalker or battle it attacks (the ruling),
// with the one Suki attacks offered first, and it was never declared, so
// it fires no attack trigger (CR 508.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9223488e-5206-4d43-90ef-387189dcee16",
		Name:         "Suki, Kyoshi Warrior",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			jmpB5PowerIsCreaturesYouControl("Suki's power is equal to the number of creatures you control."),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Suki, Kyoshi Warrior — create a 1/1 Ally that's tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Template: TokenCard("1/1 white Ally"),
						N:        1,
						Tapped:   true,
						Prefer:   triggerAttackTarget(item),
						Label:    "Suki, Kyoshi Warrior",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
