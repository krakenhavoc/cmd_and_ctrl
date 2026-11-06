package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloomvine Regent // Claim Territory — Tarkir: Dragonstorm (#4463),
// an omen card (CR 720, #5175):
//
//	Bloomvine Regent — Creature — Dragon {3}{G}{G}, 4/5
//	  "Flying
//	   Whenever this creature or another Dragon you control enters, you
//	   gain 3 life."
//	Claim Territory — Sorcery — Omen {2}{G}
//	  "Search your library for up to two basic Forest cards, reveal them,
//	   put one onto the battlefield tapped and the other into your hand,
//	   then shuffle. (Also shuffle this card.)"
//
// Two keys, one card, as for an Adventure: the Dragon under the bare
// oracle ID and the Omen under "<oracle>#1". What happens to the card
// once the Omen resolves is the engine's (CR 720.3d, game/omen.go);
// this file says nothing about it.
//
// The Regent's trigger is Ganax's condition (the source itself, or
// another Dragon entering under your control). Claim Territory is
// Cultivate's two searches with a basic Forest for a basic land.
//
// No simplification.

const bloomvineRegentOracleID = "da1e019c-2ffb-412d-90d7-f2e5e5c44c4b"

// claimTerritoryHandHalf is Claim Territory's second search: the other
// basic Forest into your hand, then the shuffle. An on-disk key.
var claimTerritoryHandHalf = landToHandTail("claim-territory/forest-to-hand", "Claim Territory",
	"basic Forest", IsBasicLandWithSubtype("forest"))

func init() {
	Register(Spec{
		OracleID:        bloomvineRegentOracleID,
		Name:            "Bloomvine Regent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b14SelfOrDragonYouControlEntered(ev, source, g)
			}, "Bloomvine Regent — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
	Register(Spec{
		OracleID:     bloomvineRegentOracleID + "#1",
		Name:         "Claim Territory",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return twoLandSearch(ctx, "Claim Territory", "basic Forest", IsBasicLandWithSubtype("forest"), claimTerritoryHandHalf)
		},
	})
}
