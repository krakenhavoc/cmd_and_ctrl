package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pantlaza, Sun-Favored — Legendary Creature — Dinosaur {2}{R}{G}{W},
// 4/4:
//
//	"Whenever Pantlaza or another Dinosaur you control enters, you may
//	 discover X, where X is that creature's toughness. Do this only
//	 once each turn."
//
// "Do this only once each turn" limits the DISCOVER, not the trigger:
// declining the first Dinosaur's offer leaves the turn's discover for
// the next one. So the gate counts this ability's resolutions this turn
// per object (#936) — an accepted "you may" is the only way it
// resolves — and is asked both before the prompt and as it resolves.
//
// X is the entering creature's toughness as the trigger resolves, from
// last-known information if it has already left (CR 608.2h,
// ctx.TriggeringPermanent). Discover is ADR 0099's (game/discover.go).
func init() {
	const label = "Pantlaza, Sun-Favored — discover X"
	Register(Spec{
		OracleID:     "b828ba28-e98d-498b-9ab4-d4aa7143e407",
		Name:         "Pantlaza, Sun-Favored",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, AllOf(pantlazaDinosaurEntered, pantlazaNotYetThisTurn(label)), label,
				func(g *game.Game, item *game.StackItem) error {
					if g.ResolvedThisTurn(item.SourceCardID, label) > 1 {
						return nil
					}
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok {
						return nil
					}
					return Discover{N: info.Toughness}.Apply(ctx)
				}), "Pantlaza, Sun-Favored — discover X, where X is that creature's toughness?"),
		},
	})
}

// pantlazaDinosaurEntered is "Pantlaza or another Dinosaur you control
// enters".
func pantlazaDinosaurEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil {
		return false
	}
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok {
		return false
	}
	return c.InstanceID == source.InstanceID || c.HasSubtype("Dinosaur")
}

// pantlazaNotYetThisTurn keeps the prompt from being asked once the
// turn's discover has happened.
func pantlazaNotYetThisTurn(label string) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return source != nil && g.ResolvedThisTurn(source.InstanceID, label) == 0
	}
}
