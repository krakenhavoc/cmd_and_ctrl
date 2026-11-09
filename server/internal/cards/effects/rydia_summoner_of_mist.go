package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rydia, Summoner of Mist — Legendary Creature — Human Shaman {R}{G},
// 1/2 (Final Fantasy, #4462):
//
//	"Landfall — Whenever a land you control enters, you may discard a
//	 card. If you do, draw a card.
//	 Summon — {X}, {T}: Return target Saga card with mana value X from
//	 your graveyard to the battlefield with a finality counter on it. It
//	 gains haste until end of turn. Activate only as a sorcery."
//
// The landfall is Rubble Rouser's rummage (the discard optional, the draw
// only after one) on Landfall. The Summon ability's X is announced as the
// cost is paid and the target must be a Saga card of exactly that mana
// value (the XValue rule, judged at announce and again at resolution);
// the Saga returns with a finality counter through the shared entry-
// counter path (CR 614.1c), so it is exiled instead if it would die, and
// gains haste until end of turn. The haste is given only to the card that
// came back, as a new object (CR 400.7).
//
// No simplification.
func init() {
	const label = "Rydia, Summoner of Mist — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "a08f79b7-d66f-47da-9670-afd7409c0c5c",
		Name:         "Rydia, Summoner of Mist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall(label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label: "Summon — {X}, {T}: Return target Saga card with mana value X from your graveyard to the battlefield with a finality counter on it. It gains haste until end of turn. Activate only as a sorcery.",
			Cost:  Plus(ManaCost("{X}"), TapCost()),
			Targets: TargetCardInGraveyard("target Saga card with mana value X from your graveyard",
				YouOwn(), OfSubtype("Saga")).WithManaValueEqualsX(),
			SorcerySpeed: true,
			Effect:       rydiaSummonReturn,
		}},
	})
}

// rydiaSummonReturn returns the still-legal target Saga card with a
// finality counter, then gives the card that came back haste until end of
// turn (CR 400.7: the haste is the new object's, so a card that stayed in
// the graveyard gets none).
func rydiaSummonReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	id := ids[0]
	if err := g.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false,
		map[string]int{game.CounterFinality: 1}, game.KeyedTail{}); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"haste"},
		Label:    "Rydia, Summoner of Mist — the returned Saga has haste",
	}.Apply(ctx)
}
