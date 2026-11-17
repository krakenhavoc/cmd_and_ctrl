package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pia Nalaar, Consul of Revival — Legendary Creature — Human Artificer
// {R}{W}, 2/3 (March of the Machine: The Aftermath):
//
//	"Thopters you control have haste.
//	 Whenever you play a land from exile or cast a spell from exile,
//	 create a 1/1 colorless Thopter artifact creature token with
//	 flying."
//
// The haste is Saheeli, Jewel of Avishkar's layer-6 grant to the
// controller's Thopters, "Thopter" read through the source's text (CR
// 612.2). The trigger is Prosper, Tome-Bound's: a cast event whose origin
// is exile, or a land played from exile (b20PlayedACardFromExile), by the
// controller. Each such cast or play makes one Thopter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65d4ebdb-8757-4e5e-9411-d78597e20d57",
		Name:         "Pia Nalaar, Consul of Revival",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Thopter"))
			}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b20PlayedACardFromExile(ev, source, g)
			}, "Pia Nalaar, Consul of Revival — create a 1/1 Thopter with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
