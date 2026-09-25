package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PuPu UFO — Artifact Creature — Construct Alien {2}, 0/4:
//
//	"Flying
//	 {T}: You may put a land card from your hand onto the battlefield.
//	 {3}: Until end of turn, this creature's base power becomes equal
//	 to the number of Towns you control."
//
// Flying rides PrintedKeywords.
//
// The tap ability is the shared "you may put a land card from your
// hand onto the battlefield" clause (MayPutALandFromHand, the
// pick-from-hand prompt with a zero floor). It is a {T} ability on a
// creature, so summoning sickness gates it (CR 302.6) — the engine
// enforces that, the card does not declare it — and the land is PUT,
// not played: it does not use up a land drop.
//
// The {3} ability is a layer 7b set of POWER only (CR 613.4b); the
// toughness stays the printed 4, and counters and anthems still apply
// on top. The count of Towns is read ONCE, when the ability resolves
// (CR 608.2h), and the affected object is pinned then too (CR
// 611.2c): a Town that enters or leaves later in the turn does not
// change it, and a UFO that is flickered comes back as its printed
// 0/4. A UFO that left the battlefield before resolution sets nothing.
// Towns are counted post-layer (MatchLandSubtype), so a creature type
// like Townsfolk does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b1412eda-e3a4-41b2-932e-795a0ba0f7f8",
		Name:            "PuPu UFO",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:  "{T}: You may put a land card from your hand onto the battlefield.",
				Cost:   TapCost(),
				Effect: Do(MayPutALandFromHand("PuPu UFO")),
			},
			{
				Label:  "{3}: Until end of turn, this creature's base power becomes equal to the number of Towns you control.",
				Cost:   ManaCost("{3}"),
				Effect: pupuUFOBasePowerFromTowns,
			},
		},
	})
}

// pupuUFOBasePowerFromTowns sets the source's base power to the Towns
// its controller controls, until end of turn.
func pupuUFOBasePowerFromTowns(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	towns := countControlled(g, ctx.Controller(), MatchLandSubtype("Town"))
	return untilEndOfTurn(ctx, ctx.Source(), nil,
		fmt.Sprintf("PuPu UFO — base power %d until end of turn", towns),
		game.SetBasePowerMod(towns))
}
