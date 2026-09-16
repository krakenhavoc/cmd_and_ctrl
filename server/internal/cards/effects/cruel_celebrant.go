package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cruel Celebrant — Creature — Vampire, {W}{B}, 1/2 (EDHREC rank 952):
//
//	"Whenever this creature or another creature or planeswalker you
//	 control dies, each opponent loses 1 life and you gain 1 life."
//
// The Orzhov Zulaport Cutthroat: a drain on every death of your own,
// planeswalkers included, and on its own. One dies trigger. The
// self case is the ordinary "this dies" read on the LTB harvest; the
// other case reads the dead card post-move — controller and type
// line survive the trip to the graveyard — and admits creatures and
// planeswalkers you controlled. An opponent's creature is not "you
// control" and an exile or a bounce is not a death.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ee78cfc-0e9e-4737-a7e2-b42f94228040",
		Name:         "Cruel Celebrant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if cardDied(ev, source) {
					return true
				}
				if ev.NewZone != game.ZoneGraveyard {
					return false
				}
				dead, ok := g.LookupCardForEffect(ev.CardID)
				return ok && dead.Controller == source.Controller && (dead.IsCreature() || dead.IsPlaneswalker())
			}, "Cruel Celebrant — each opponent loses 1 life, you gain 1 life", b07DrainEachOpponent),
		},
	})
}
