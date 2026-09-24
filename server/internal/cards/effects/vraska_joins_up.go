package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vraska Joins Up — Legendary Enchantment {B}{G} (EDHREC rank 2594):
//
//	"When Vraska Joins Up enters, put a deathtouch counter on each
//	 creature you control.
//	 Whenever a legendary creature you control deals combat damage to
//	 a player, draw a card."
//
// The Golgari "Joins Up" card. The ETB puts a deathtouch counter on
// every creature the controller controls (snapshotted before the
// first counter lands, so a token made in response gets none); the
// draw trigger is combatDamageToPlayerBy narrowed to a legendary
// dealer, read post-layer (b24LegendaryCreatureYouControlDealtCombatDamageToPlayer),
// one card per legendary creature that connects.
//
// DECLARED SIMPLIFICATION, weaker than printed: the engine reads no
// keyword counters of its own (CR 122.1b), so Vraska carries that
// rule for the counters it places — a Layer 6 static granting
// deathtouch to every creature with a deathtouch counter
// (b24KeywordCounterGrant), any creature's, anyone's, exactly the
// rule and never more. It lives on Vraska, so it stops when Vraska
// leaves the battlefield: the counters stay on the creatures and go
// inert, where the printed card's would keep granting. The gap runs
// the weaker way; it closes when the engine reads keyword counters
// itself, at which point this static becomes a no-op and comes out.
func init() {
	Register(Spec{
		OracleID:     "c91b0dd5-4c63-49a5-95bf-c342b6ff2076",
		Name:         "Vraska Joins Up",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Deathtouch counters only grant deathtouch while Vraska Joins Up is on the battlefield — once it leaves, the counters stay but stop working."},
		Static:       []game.StaticAbility{b24KeywordCounterGrant("deathtouch")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Vraska Joins Up — put a deathtouch counter on each creature you control", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range b04CreatureIDsControlledBy(g, item.Controller) {
					if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
						continue
					}
					if err := (AddCounter{Target: id, Kind: "deathtouch", N: 1}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24LegendaryCreatureYouControlDealtCombatDamageToPlayer(ev, source, g)
			}, "Vraska Joins Up — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
