package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Toshiro Umezawa — Legendary Creature — Human Samurai {1}{B}{B}, 2/2 (Secret
// Lair Drop):
//
//	"Bushido 1
//	 Whenever a creature an opponent controls dies, you may cast target instant
//	 card from your graveyard. If that spell would be put into a graveyard,
//	 exile it instead."
//
// Dreadhorde Arcanist's grant with the price paid: a cast-only, flash-speed
// permission over the targeted card, offered during the resolution (CR 608.2g),
// whose ExileInsteadOfGraveyard exiles the spell whether it resolves, is
// countered or fizzles.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11838086-db2f-4588-ae18-4129c9e2b67d",
		Name:         "Toshiro Umezawa",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Bushido(1, "Toshiro Umezawa"),
			effects.Targeting(
				effects.Optional(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.OldB11OpponentCreatureDied(ev, source, g)
				}, "Toshiro Umezawa — you may cast target instant card from your graveyard; if it would be put into a graveyard, exile it instead",
					effects.OldB12CastTargetInstantFromYourGraveyardExilingIt("Toshiro Umezawa")),
					"Toshiro Umezawa — cast target instant card from your graveyard?"),
				effects.TargetCardInGraveyard("target instant card from your graveyard", effects.YouOwn(), effects.Instant())),
		},
	})
}
