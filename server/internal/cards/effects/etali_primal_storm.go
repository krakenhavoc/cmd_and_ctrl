package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Etali, Primal Storm — Legendary Creature — Elder Dinosaur
// {4}{R}{R}, 6/6 (EDHREC rank 269):
//
//	"Whenever Etali attacks, exile the top card of each player's
//	 library, then you may cast any number of spells from among
//	 those cards without paying their mana costs."
//
// Impulse exile (`Game.ExileTopWithPermissionForEffect`, S21 sub-PR
// 6), the same primitive Bloodbraid Elf's cascade and Ragavan's
// impulse both ride, called once per seat with the permission handed
// to ETALI'S CONTROLLER rather than to each library's owner —
// exactly Ragavan's asymmetry ("usually NOT the owner"). `CastOnly`
// strands an exiled land, which is the printed sentence: it names
// SPELLS.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "078def07-ae5d-4591-8db6-d156834aab97",
		Name:         "Etali, Primal Storm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Etali, Primal Storm — exile the top card of each player's library",
				etaliExileAndOfferCasts),
		},
	})
}

// etaliExileAndOfferCasts is the trigger body: exile the top card of
// every seat's library and let Etali's controller cast any of them
// for free.
func etaliExileAndOfferCasts(g *game.Game, item *game.StackItem) error {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if _, err := g.ExileTopWithPermissionForEffect(p.ID, item.Controller, 1, game.CastPermission{
			Cost:     "{0}",
			CastOnly: true,
		}); err != nil {
			return err
		}
	}
	return nil
}
