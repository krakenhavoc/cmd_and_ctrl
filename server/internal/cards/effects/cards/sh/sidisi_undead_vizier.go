package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sidisi, Undead Vizier — Legendary Creature — Zombie Snake {3}{B}{B}, 4/6 (Dragons of Tarkir, #5551):
//
//	"Deathtouch
//	 Exploit (When this creature enters, you may sacrifice a creature.)
//	 When Sidisi exploits a creature, you may search your library for a card, put it into your hand, then shuffle."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff is a reflexive trigger raised only when a creature was really sacrificed.
// The search is Profane Tutor's, made optional: "you may search" lets the
// controller decline before looking.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "3406db41-6b95-48fc-b58e-4476220fd34d",
		Name:            "Sidisi, Undead Vizier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Sidisi, Undead Vizier", sidisiUndeadVizierSearchBody),
		},
	})
}

// sidisiUndeadVizierSearchBody is "When Sidisi exploits a creature, you may search your library for a card, put it into your hand, then shuffle." An on-disk key:
// never renamed, never reused.
var sidisiUndeadVizierSearchBody = game.ReflexiveBody("sidisi-undead-vizier/search", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.SearchLibrary{
		Player:   ctx.Controller(),
		Dest:     game.ZoneHand,
		Limit:    1,
		Shuffle:  true,
		Optional: true,
		Reason:   "Sidisi, Undead Vizier — search your library for a card",
	}.Apply(ctx)
}), nil)
