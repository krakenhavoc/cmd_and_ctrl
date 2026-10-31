package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Director Nick Fury — Legendary Creature — Human Spy Hero {U}{R}{W}, 2/4
// (Marvel Super Heroes Commander, #4455):
//
//	"Hero spells you cast cost {1} less to cast.
//	 Whenever you attack, look at the top four cards of your library. You
//	 may reveal a Hero card from among them and put that card into your
//	 hand. Put the rest on the bottom of your library in a random order."
//
// The discount is the cast modifier over Hero spells the controller casts
// (a changeling spell counts; "Hero" is a word of Fury's text). "Whenever
// you attack" is one trigger per attack declaration (Great Gilded Boat's
// shape); the look is Bucky Barnes's LookAtTopThenMayTakeToHand, which puts
// the rest underneath in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04b907bf-c80c-4e92-9fea-26dc3d1aaf2a",
		Name:         "Director Nick Fury",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Hero spells you cast cost {1} less to cast.",
				YourSpell(), SpellOfCreatureType("Hero")),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Director Nick Fury — look at the top four cards; you may take a Hero",
				LookAtTopThenMayTakeToHand(4, AnySubtype("Hero"), 1, "Director Nick Fury — you may reveal a Hero card and put it into your hand"))),
		},
	})
}
