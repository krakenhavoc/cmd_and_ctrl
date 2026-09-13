package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teshar, Ancestor's Apostle — Legendary Creature — Bird Cleric
// {3}{W}, 2/2 (EDHREC rank 1916):
//
//	"Flying
//	 Whenever you cast a historic spell, return target creature card
//	 with mana value 3 or less from your graveyard to the battlefield.
//	 (Artifacts, legendaries, and Sagas are historic.)"
//
// The cheap-artifact recursion engine. The trigger is an ordinary
// "whenever you cast" gated on the spell being historic
// (b09IsHistoric reads the spell off the stack, where its type line
// is intact); the return is a targeted trigger over the controller's
// own graveyard — creature cards at mana value 3 or less — answered
// through the zone browser, the Sun Titan shape. With no legal
// target the trigger is removed (CR 603.3d), as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4193f50-33b5-4201-ba03-eb135eced7ff",
		Name:            "Teshar, Ancestor's Apostle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17HistoricSpellCastByYou(ev, source, g)
			},
			Targets: TargetCardInGraveyard(
				"target creature card with mana value 3 or less in your graveyard",
				YouOwn(), Creature(), ManaValueLE(3),
			),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Teshar, Ancestor's Apostle — return a creature card to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return ReturnFromGraveyard{
							Target: item.Targets[0].ID,
							Dest:   game.ZoneBattlefield,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
