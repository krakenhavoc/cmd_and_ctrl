package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mindblade Render — Creature — Azra Warrior {1}{B}, 1/3 (Tarkir: Dragonstorm Commander)
// Commander):
//
//	"Whenever your opponents are dealt combat damage, if any of that
//	 damage was dealt by a Warrior, you draw a card and you lose 1 life."
//
// One trigger per combat damage step however many opponents were hit and
// however many Warriors hit them (OncePerBatch, the printed ruling). An
// event qualifies when it is combat damage to one of YOUR opponents from
// a Warrior, whoever controls that Warrior: a Warrior an opponent controls
// hitting another opponent of yours counts, and is not optional. The
// damage source is read as it stands when the damage is dealt, with the
// Render's own text for the word "Warrior" (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "028dcef2-3de9-4dde-ba17-f9e077b3190d",
		Name:         "Mindblade Render",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, mindbladeRenderWarriorHitAnOpponent,
				"Mindblade Render — you draw a card and you lose 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -1)
				})),
		},
	})
}

// mindbladeRenderWarriorHitAnOpponent is "your opponents are dealt combat
// damage … by a Warrior": combat damage to a player other than the
// Render's controller, from a Warrior.
func mindbladeRenderWarriorHitAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Target == source.Controller {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.HasSubtype(source.TextSubtype("Warrior"))
}
