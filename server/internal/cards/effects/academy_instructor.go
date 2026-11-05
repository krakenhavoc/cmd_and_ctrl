package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Academy Instructor — Creature — Bolian Officer {2}{R}, 0/3 (Star Trek,
// #4973):
//
//	"This creature gets +1/+0 for each Officer token you control.
//	 Whenever you cast a noncreature spell, create a 1/1 red Officer
//	 creature token."
//
// The pump is a layer 7c modify that counts on every pass
// (trekB14OfficerTokensYouControl), so it rises with each Officer the
// trigger makes and falls as one leaves. The Instructor is an Officer but
// not a token, so it never counts itself. The trigger is Kykar's cast
// trigger with the printed ttrk Officer: it fires on casting, so a
// countered spell still makes the token.
//
// No simplification.
func init() {
	pump := PumpSelfCreatureOrAttachedPer(1, 0, trekB14OfficerTokensYouControl)
	pump.Label = "This creature gets +1/+0 for each Officer token you control."
	Register(Spec{
		OracleID:     "8802b7f0-1ed1-4872-9939-182afb8441ca",
		Name:         "Academy Instructor",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{pump},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Academy Instructor — create a 1/1 red Officer creature token",
				Do(CreateToken{Template: TokenCard("1/1 red Officer"), N: 1})),
		},
	})
}
