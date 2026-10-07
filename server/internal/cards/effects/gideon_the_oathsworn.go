package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gideonTheOathswornLabel = "Gideon, the Oathsworn — put a +1/+1 counter on each of the non-Gideon creatures that attacked"

// Gideon, the Oathsworn — Legendary Planeswalker — Gideon {4}{W}{W},
// starting loyalty 4:
//
//	"Whenever you attack with two or more non-Gideon creatures, put a
//	 +1/+1 counter on each of those creatures.
//	 +2: Until end of turn, Gideon, the Oathsworn becomes a 5/5 white
//	 Soldier creature that's still a planeswalker. Prevent all damage
//	 that would be dealt to him this turn. (He can't attack if he was
//	 cast this turn.)
//	 −9: Exile Gideon, the Oathsworn and each creature your opponents
//	 control."
//
// THE TRIGGER is Chivalric Alliance's batch shape (OncePerBatch on the
// per-creature EventAttack): it fires on the first attacker event of a
// declaration that finds two non-Gideon creatures of yours attacking,
// and the rest of that declaration's events are declined, so a
// five-creature attack triggers once. "Non-Gideon" reads the subtype off
// the effective characteristics, so a Gideon that has become a creature
// and attacks is not counted. The counters go on the non-Gideon
// creatures of yours that are attacking as the trigger resolves, which
// is the declared set unless one was removed from combat in between.
// THE +2 is animateGideon (gideon_animate.go; #2046, ADR 0032 amendment
// of 2026-10-07): white and Soldier only, no indestructible. The
// parenthetical is summoning sickness (CR 302.6), which the engine
// enforces for any creature. THE −9 exiles him and every creature an
// opponent controls in one move; a Gideon that left in response is a new
// object and is not exiled again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "19a81aa7-823b-43fa-abc2-b2700a122bc1",
		Name:            "Gideon, the Oathsworn",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor == source.Controller &&
					len(attackingNonGideonCreatures(g, source.Controller)) >= 2
			}, gideonTheOathswornLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return b13PutCounterOnEach(ctx, attackingNonGideonCreatures(g, item.Controller))
			})),
		},
		Activated: []ActivatedAbility{
			{
				Label: "+2: Until end of turn, Gideon, the Oathsworn becomes a 5/5 white Soldier creature that's still a planeswalker. Prevent all damage that would be dealt to him this turn.",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return animateGideon(g, item, gideonAnimation{
						Label:     "Gideon, the Oathsworn — a 5/5 white Soldier creature until end of turn",
						Subtypes:  []string{"Soldier"},
						Colors:    []string{"W"},
						Power:     5,
						Toughness: 5,
					})
				},
			},
			{
				Label: "−9: Exile Gideon, the Oathsworn and each creature your opponents control.",
				Cost:  LoyaltyCost(-9),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					self := item.SourceCardID
					exileSelf := !ctx.isNewSourceObject(self)
					return ExileAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						if c.InstanceID == self {
							return exileSelf
						}
						return c.IsCreature() && c.Controller != item.Controller
					}}.Apply(ctx)
				},
			},
		},
	})
}

// attackingNonGideonCreatures lists the creatures `controller` controls
// that are attacking and are not Gideons, in battlefield order.
func attackingNonGideonCreatures(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.AttackingTarget != uuid.Nil && !c.HasSubtype("Gideon") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
