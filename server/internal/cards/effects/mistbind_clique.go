package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistbind Clique — Creature — Faerie Wizard {3}{U}, 4/4 (Special Guests):
//
//	"Flash
//	 Flying
//	 Champion a Faerie (When this enters, sacrifice it unless you exile
//	 another Faerie you control. When this leaves the battlefield, that card
//	 returns to the battlefield.)
//	 When a Faerie is championed with this creature, tap all lands target
//	 player controls."
//
// Champion is the shared champion helper (s82_bns_b05_helpers.go, CR
// 702.72). "Championed with this creature" (CR 702.72c) is the card the
// champion ability really exiled, so the tap trigger is a reflexive trigger
// queued from the exile's continuation: it goes on the stack above the
// champion ability with its target player chosen as it does (CR 603.3d), and
// it never happens when nothing was exiled (the Clique was sacrificed, or had
// already left).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d1946f91-0653-4647-b3de-95d78b9597c4",
		Name:            "Mistbind Clique",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered:       []game.TriggeredAbility{mistbindCliqueChampion.trigger("Mistbind Clique")},
	})
}

// mistbindCliqueChampion is "Champion a Faerie". Its pick key is an on-disk
// identity.
var mistbindCliqueChampion = newChampion("Mistbind Clique", "Faerie", "choose-cards/s82-bns-b05-mistbind-clique-champion", mistbindCliqueChampioned)

// mistbindCliqueChampioned queues the reflexive tap trigger once a Faerie was
// exiled by the champion ability. An on-disk key.
var mistbindCliqueChampioned = CardTail("s82-bns-b05/mistbind-clique-championed", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return WhenYouDo("Mistbind Clique — tap all lands target player controls", mistbindCliqueTapBody).Apply(ctx)
})

// mistbindCliqueTapBody is "tap all lands target player controls".
var mistbindCliqueTapBody = game.ReflexiveBody("s82-bns-b05/mistbind-clique-tap-lands",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if p := TargetedPlayer(ctx); p != uuid.Nil {
			return b24TapAllLandsControlledBy(ctx, p)
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec { return TargetPlayer("target player") }))
