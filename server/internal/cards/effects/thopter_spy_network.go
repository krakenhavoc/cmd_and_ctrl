package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thopter Spy Network — Enchantment, {2}{U}{U} (EDHREC rank 936):
//
//	"At the beginning of your upkeep, if you control an artifact,
//	 create a 1/1 colorless Thopter artifact creature token with
//	 flying.
//	 Whenever one or more artifact creatures you control deal combat
//	 damage to a player, draw a card."
//
// The artifact deck's engine: a Thopter every upkeep as long as any
// artifact is around (the first Thopter keeps it going by itself),
// and a card whenever the fliers connect.
//
// The upkeep half is Land Tax's intervening-if posture: "if you
// control an artifact" is checked when the trigger would go on the
// stack and not again at resolution, so an artifact removed in
// response still yields a Thopter — the same declared corner every
// intervening-if card in the catalog takes.
//
// The combat half is "one or more", so it is deduplicated with the
// batch 04 helper: the engine emits one damage event per creature,
// and the second artifact creature's event is declined while the
// first trigger is still queued or on the stack. Without that the
// card would ship STRONGER than printed (#259).
func init() {
	Register(Spec{
		OracleID:     "49be65fd-3755-410d-b0dc-2e5861ea2552",
		Name:         "Thopter Spy Network",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The artifact check happens only when the upkeep trigger goes on the stack, so losing your last artifact in response won't stop the Thopter."},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b03ArtifactsControlled(g, source.Controller) > 0
			}, "Thopter Spy Network — create a 1/1 Thopter", Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !combatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				src, ok := g.LookupCardForEffect(ev.Source)
				return ok && src.IsArtifact()
			}, "Thopter Spy Network — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
