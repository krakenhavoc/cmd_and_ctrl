package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baneblade Scoundrel // Baneclaw Marauder — {3}{B} Creature — Human
// Rogue Werewolf 4/3 // Creature — Werewolf 5/4 (#2586, ADR 0132):
//
//	Front: "Whenever this creature becomes blocked, each creature
//	        blocking it gets -1/-1 until end of turn.
//	        Daybound"
//	Back:  "Whenever this creature becomes blocked, each creature
//	        blocking it gets -1/-1 until end of turn.
//	        Whenever a creature blocking this creature dies, that
//	        creature's controller loses 1 life.
//	        Nightbound"
//
// "Becomes blocked" is the once-per-attacker EventBecomesBlocked (CR
// 506.4), so a double block is one trigger that shrinks both blockers.
// The blockers are read as the trigger resolves (Feint's "each creature
// blocking it"). The back face's second trigger reads the dying
// creature's combat state off the leave event (CR 603.10a): it must have
// been blocking THIS creature, and the life is lost by the player who
// controlled it as it died.
//
// No simplification.
func init() {
	const oracle = "4f5da665-e880-4325-a9e2-6fc4ca1d807a"
	shrink := func(name string) game.TriggeredAbility {
		label := name + " — each creature blocking it gets -1/-1 until end of turn"
		return On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID
		}, label, func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, id := range blockersOf(g, item.SourceCardID) {
				if err := (BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: label}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		})
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Baneblade Scoundrel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered:       []game.TriggeredAbility{shrink("Baneblade Scoundrel")},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Baneclaw Marauder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered: []game.TriggeredAbility{
			shrink("Baneclaw Marauder"),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					dead, ok := diedWhileBlocking(ev, g)
					return ok && ev.BlockingTarget == source.InstanceID && leftAsType(ev, dead, "creature")
				},
				Key: "Baneclaw Marauder — that creature's controller loses 1 life",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Baneclaw Marauder — that creature's controller loses 1 life")
					if dead, ok := g.LookupCardForEffect(ev.CardID); ok {
						item.Params.Object = game.ObjectRef{ID: leftUnderControlOf(ev, dead)}
					}
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					victim := item.Params.Object.ID
					if g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1)
				},
			},
		},
	})
}
