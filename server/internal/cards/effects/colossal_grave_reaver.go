package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const b17GraveReaverReturnLabel = "Colossal Grave-Reaver — put a milled creature card onto the battlefield"

// Colossal Grave-Reaver — Creature — Dragon {6}{B}{G}, 7/6 (EDHREC
// rank 1890):
//
//	"Flying
//	 Whenever this creature enters or attacks, mill three cards.
//	 Whenever one or more creature cards are put into your graveyard
//	 from your library, put one of them onto the battlefield."
//
// The self-mill Dragon that reanimates as it goes. The first trigger
// is Sun Titan's "enters or attacks" shape with a three-card mill.
// The second watches EventMill — the engine emits one per card, so
// "one or more" is the per-label dedup: the first creature card of
// a mill fires the trigger and the rest of that mill are declined
// while it is queued or on the stack. At resolution the batch is
// read back off the event log (b17MilledCreatureCards) — every
// creature card that mill put into the controller's graveyard and
// is still there — and one of them comes back under the
// controller's control.
//
// Sandbox simplification, declared: "put ONE OF THEM" is a choice,
// and the engine has no resolution-time pick-a-card prompt for a
// trigger (the pick_target prompt freezes its legal set when the
// trigger fires, which is after the FIRST card of the mill and
// before the rest), so the engine picks the creature card with the
// greatest mana value, the first milled on a tie. Never stronger
// than printed — every pick is one the printed card allows — only
// less controllable. Two mills in one resolution that are not one
// batch fire once, not twice: weaker, never stronger.
func init() {
	Register(Spec{
		OracleID:        "df8e0d1b-b47c-4807-9c9b-84dcec835254",
		Name:            "Colossal Grave-Reaver",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"When creature cards are milled, the one with the greatest mana value comes back automatically rather than one you choose."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB, game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Colossal Grave-Reaver — mill three cards",
						func(g *game.Game, item *game.StackItem) error {
							return MillCards{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventMill},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Actor != source.Controller || ev.NewZone != game.ZoneGraveyard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					if !ok || !c.IsCreature() || c.Owner != source.Controller {
						return false
					}
					return !b12TriggerPendingOrOnStack(g, source, b17GraveReaverReturnLabel)
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					seq := ev.Seq
					return game.NewTriggeredItem(source, b17GraveReaverReturnLabel,
						func(g *game.Game, item *game.StackItem) error {
							pick, ok := b17GreatestManaValue(g, b17MilledCreatureCards(g, item.Controller, seq))
							if !ok {
								return nil
							}
							return ReturnFromGraveyard{
								Target:     pick,
								Dest:       game.ZoneBattlefield,
								Controller: item.Controller,
							}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
