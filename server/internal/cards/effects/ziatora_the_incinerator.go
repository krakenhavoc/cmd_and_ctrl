package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ziatora, the Incinerator — Legendary Creature — Demon Dragon
// {3}{B}{R}{G}, 6/6 (EDHREC rank 3098):
//
//	"Flying
//	 At the beginning of your end step, you may sacrifice another
//	 creature. When you do, Ziatora deals damage equal to that
//	 creature's power to any target and you create three Treasure
//	 tokens."
//
// The Riveteers' Fling engine. Flying rides PrintedKeywords. The
// end-step trigger is Ruthless Technomancer's shape: "you may
// sacrifice another creature" is asked as the trigger RESOLVES
// (maySacrificeOne, CR 608.2c), over whatever you control then, and
// the creature is sacrificed (b29SacrificeChosenCreature).
//
// The printed "when you do" is a CR 603.12 REFLEXIVE trigger, and
// since #636 it is exactly that: the end-step ability creates it as
// it resolves, having actually sacrificed something. It goes on the
// stack above its parent with its own "any target" clause — chosen
// after the sacrifice, as printed — carrying the sacrificed
// creature's instance ID as its payload, and on resolution deals
// damage equal to that creature's power as it last stood
// (b29PowerAsItLastStood) and makes the three Treasures whether or
// not the damage lands.
//
// No simplification. The damage is the sacrificed creature's
// last-known power (CR 608.2h), so an anthem or an Equipment counts as
// a counter does (b29PowerAsItLastStood).
//
// History: the reflexive half used to be
// a second ability harvested off Ziatora on the battlefield, so a
// Ziatora removed in response to her own end-step trigger took the
// damage and the Treasures with her. The trigger now belongs to the
// resolving ability, not to the permanent, so it happens either way
// — which is what the card says.
func init() {
	Register(Spec{
		OracleID:        "d46c3fb6-f1c7-4a96-ae42-5bce17fc7c1d",
		Name:            "Ziatora, the Incinerator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep(b29ZiatoraSacrificeLabel, b29SacrificeChosenCreature),
		},
	})
}
