package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nakia, Wakandan Operative — Legendary Creature — Human Warrior Hero
// {2}{G}, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever your commander enters, you become the monarch.
//	 {2}, {T}: Put two +1/+1 counters on target creature or Vehicle.
//	 Activate only as a sorcery."
//
// "Your commander" is a commander its controller OWNS (Tome of Legends'
// reader), and Nakia herself counts when she is the commander: the
// trigger sees her own entry. The counters go on a creature or a
// Vehicle (a Vehicle that is not crewed is not a creature, which is why
// the target says both).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3f2ab9ff-10b8-494a-be7a-2fce172f62a3",
		Name:            "Nakia, Wakandan Operative",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCommander && c.Owner == source.Controller
			}, "Nakia, Wakandan Operative — you become the monarch", Do(BecomeTheMonarch{})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{2}, {T}: Put two +1/+1 counters on target creature or Vehicle. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}"), TapCost()),
			SorcerySpeed: true,
			Targets:      TargetPermanent("target creature or Vehicle", Or(Creature(), Subtype("Vehicle"))),
			Effect:       twoPlusOneCountersOnFirstLegalTarget,
		}},
	})
}
