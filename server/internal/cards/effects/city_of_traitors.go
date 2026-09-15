package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// City of Traitors — Land (EDHREC rank 1176):
//
//	"When you play another land, sacrifice this land.
//	 {T}: Add {C}{C}."
//
// Ancient Tomb without the pain and with a shelf life: two colourless
// until the next land drop. The mana is Sol Ring's; the drawback is a
// trigger on the controller's next land play, resolved by sacrificing
// the City (SacrificePermanent, so a "whenever you sacrifice" payoff
// sees it). A City already gone when the trigger resolves does
// nothing.
//
// The engine emits no land-PLAY event, so the trigger reads the zone
// a land arrived FROM (b10AnotherLandPlayedByYou): a land played from
// hand, from an impulse-exile grant or from a graveyard permission
// counts, and a land put onto the battlefield from the LIBRARY —
// Cultivate, a fetchland, Rampant Growth — does not, which is the
// printed distinction for every common case.
//
// Sandbox simplification, declared: a land an effect RETURNS to the
// battlefield from a graveyard or exile (Titania's ETB, Splendid
// Reclamation, a flickered land) also reads as a play and costs the
// City. Weaker than printed for the City's controller, never
// stronger.
func init() {
	Register(Spec{
		OracleID:     "f161111d-9747-47b3-bb10-3c8bded32e21",
		Name:         "City of Traitors",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A land returned to the battlefield from your graveyard or from exile also counts as playing a land."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherLandPlayedByYou(ev, source, g)
			}, "City of Traitors — sacrifice it", func(g *game.Game, item *game.StackItem) error {
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
	})
}
