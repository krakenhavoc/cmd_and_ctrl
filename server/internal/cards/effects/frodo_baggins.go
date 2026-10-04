package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frodo Baggins — Legendary Creature — Halfling Scout {G}{W}, 1/3:
//
//	"Whenever Frodo Baggins or another legendary creature you control
//	 enters, the Ring tempts you.
//	 As long as Frodo Baggins is your Ring-bearer, it must be blocked
//	 if able."
//
// The block requirement is Gaea's Protector's, switched on by the
// designation: only one creature is required to block it, and a
// defending player whose creatures can't block it (the Ring's first
// line, for one) owes nothing (2023-06-16 rulings, CR 509.1c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4dd6edf-1e9e-46fa-92b5-5df9aa0e4338",
		Name:         "Frodo Baggins",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			BlockRequirementWhere(game.BlockRequirementMustBeBlocked, func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && game.IsRingBearerOf(*target, source.Controller)
			}),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, frodoOrAnotherLegendaryCreatureYouControlEntered, "Frodo Baggins — the Ring tempts you", Do(TheRingTemptsYou{})),
		},
	})
}

// frodoOrAnotherLegendaryCreatureYouControlEntered is "Frodo Baggins
// or another legendary creature you control enters".
func frodoOrAnotherLegendaryCreatureYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.IsCreature() && c.IsLegendary()
}
