package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sefris of the Hidden Ways — Legendary Creature — Human Wizard
// {W}{U}{B}, 2/3:
//
//	"Whenever one or more creature cards are put into your graveyard
//	 from anywhere, venture into the dungeon. This ability triggers only
//	 once each turn.
//	 Create Undead — Whenever you complete a dungeon, return target
//	 creature card from your graveyard to the battlefield."
//
// "From anywhere": a dying creature, a discard, a mill, a countered
// spell. A creature card is a card (a token is not, CR 108.2), and it is
// your graveyard by ownership. "Only once each turn" is the turn tally
// of this object's triggers, so the second event of one batch finds the
// first already counted.
//
// No simplification.
func init() {
	label := ventureLabel("Sefris of the Hidden Ways")
	Register(Spec{
		OracleID:     "9b21e056-57e7-4e39-b2b4-2fdf8719d0d2",
		Name:         "Sefris of the Hidden Ways",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventZoneMove, game.EventMill, game.EventLTB},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if source == nil || ev.NewZone != game.ZoneGraveyard {
						return false
					}
					if b11TriggeredThisTurn(g, source.InstanceID, label) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && !IsToken(c) && c.IsCreature() && c.Owner == source.Controller
				}, label, Venture),
			Targeting(WheneverYouCompleteADungeon("Create Undead — return target creature card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				}), TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())),
		},
	})
}
