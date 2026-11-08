package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghitu Lavarunner — Creature — Human Wizard {R}, 1/2 (Foundations,
// #4465):
//
//	"As long as there are two or more instant and/or sorcery cards in your
//	 graveyard, this creature gets +1/+0 and has haste."
//
// Undercover Skrull's graveyard gate on two statics, both read live on
// every layer pass: the +1/+0 is a layer 7c modification and haste a
// layer 6 grant, so the two come and go together as the graveyard fills
// and empties. A card that is both an instant and a sorcery counts once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e3f4987-6f31-4bc5-8bcd-6adb0440ee7c",
		Name:         "Ghitu Lavarunner",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as there are two or more instant and/or sorcery cards in your graveyard, this creature gets +1/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: ghituTwoInstantOrSorceryCardsInYourGraveyard,
				Apply:     pumpBy(1, 0),
			},
			b16GrantKeywords(ghituTwoInstantOrSorceryCardsInYourGraveyard, "haste"),
		},
	})
}

// ghituTwoInstantOrSorceryCardsInYourGraveyard is the gate: the source
// itself, while its controller's graveyard holds two or more instant
// and/or sorcery cards.
func ghituTwoInstantOrSorceryCardsInYourGraveyard(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID &&
		GraveyardAtLeast(2, MatchInstantOrSorcery)(g, source.Controller, source.InstanceID)
}
