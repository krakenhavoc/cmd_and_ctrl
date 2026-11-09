package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cecil, Dark Knight // Cecil, Redeemed Paladin (Final Fantasy, #4462).
// Front, Legendary Creature — Human Knight {B}, 2/3:
//
//	"Deathtouch
//	 Darkness — Whenever Cecil deals damage, you lose that much life.
//	 Then if your life total is less than or equal to half your starting
//	 life total, untap Cecil and transform it."
//
// Back, Legendary Creature — Human Knight, 4/4:
//
//	"Lifelink
//	 Protect — Whenever Cecil attacks, other attacking creatures gain
//	 indestructible until end of turn."
//
// Darkness is Spirit Link's damage trigger, read for any damage Cecil
// deals (combat or not, to a player, creature or planeswalker): the loss
// is the damage dealt, and "then if" is the keyed tail of the life loss
// (so it waits for a replacement prompt on the loss), reading the table's
// starting life (40 in Commander). Untap and transform are the in-place
// verbs on Cecil, so a Cecil that died with the damage (the ruling) is
// neither untapped nor turned over. Protect gives every other attacking
// creature indestructible, not only the controller's: the attackers are
// whoever is attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        cecilOracleID,
		Name:            "Cecil, Dark Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			},
			Key: "Cecil, Dark Knight — Darkness: you lose that much life; then if your life total is half your starting life total or less, untap Cecil and transform it",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return g.LoseLifeEachKeyedForEffect(item.SourceCardID, []uuid.UUID{item.Controller},
					item.Trigger.Event.Amount,
					game.KeyedTail{Ref: cecilDarknessThen, Player: item.Controller, Source: item.SourceCardID})
			},
		}},
	})
	Register(Spec{
		OracleID:        cecilOracleID + "#1",
		Name:            "Cecil, Redeemed Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Cecil, Redeemed Paladin — Protect: other attacking creatures gain indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(AttackingCreature(), OtherThan(item.SourceCardID)),
						Keywords: []string{"indestructible"},
						Label:    "Cecil, Redeemed Paladin — Protect: indestructible",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// cecilOracleID is shared by both faces.
const cecilOracleID = "e4719557-d94e-490f-8a56-c1867a01d2ef"

// cecilDarknessThen is "Then if your life total is less than or equal to
// half your starting life total, untap Cecil and transform it", run once
// the life loss is settled. An on-disk key.
var cecilDarknessThen = CardTail("then/cecil-dark-knight-darkness", func(ctx *Context, o game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(o.Tail.Player)
	if p == nil || p.Life*2 > startingLifeOf(ctx.Game) {
		return nil
	}
	// A second Darkness that resolves after the first already turned Cecil
	// over finds the Paladin, which is not the object that triggered it
	// (CR 701.28f): it does not untap or transform it back.
	if c, ok := ctx.Game.LookupCardForEffect(o.Tail.Source); !ok || c.ActiveFace != 0 {
		return nil
	}
	if err := (UntapTarget{Target: o.Tail.Source}).Apply(ctx); err != nil {
		return err
	}
	return Transform{Target: o.Tail.Source}.Apply(ctx)
})
