package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thor, God of Thunder — {3}{R}{R} Legendary Creature — God Warrior Hero 5/5
// (Marvel Super Heroes, #4455):
//
//	"Flying
//	 When Thor enters, exile target Equipment, instant, or sorcery card
//	 from your graveyard. Until the end of your next turn, you may play that
//	 card.
//	 Whenever you cast a noncreature spell, Thor deals damage equal to that
//	 spell's mana value to any target."
//
// The enters trigger exiles the chosen card with a play permission that ends
// with the controller's next turn (ADR 0063, Gundabad Opportunist's
// duration); "play" and not "cast", so nothing is stranded. No target in the
// graveyard means no trigger (CR 603.3d). The cast trigger is Kaervek's
// shape: the spell's mana value is read as the trigger fires, with its
// announced X (CR 202.3e), because the spell may be gone by the time the
// damage resolves; Thor is the damage source.
//
// No simplification.
func init() {
	enters := Targeting(
		WhenThisEnters("Thor — exile target Equipment, instant, or sorcery card from your graveyard; you may play it until the end of your next turn", thorExileAndGrant),
		TargetCardInGraveyard("target Equipment, instant, or sorcery card in your graveyard",
			Or(OfSubtype("Equipment"), Instant(), Sorcery()), YouOwn()))
	Register(Spec{
		OracleID:        "2842217a-173e-4879-8b25-fefe2fc1123f",
		Name:            "Thor, God of Thunder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			enters,
			{
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: YouCast(Noncreature()),
				Targets:   TargetAny(),
				Key:       "Thor — deal damage equal to that spell's mana value to any target",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					amount := 0
					if spell, ok := g.LookupCardForEffect(ev.CardID); ok {
						amount, _ = g.ManaValueForEffect(spell)
					}
					item := game.NewTriggeredItem(source, "Thor — deal damage equal to that spell's mana value to any target")
					item.Params.Amount = amount
					return item
				},
				Effect: DealAmountFromParamsToFirstTarget,
			},
		},
	})
}

func thorExileAndGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	controller := item.Controller
	return g.ExileCardWithPermissionForEffect(id, game.CastPermission{
		Player:   controller,
		Duration: g.UntilEndOfYourNextTurnDuration(controller),
		Label:    "Thor — play it until the end of your next turn",
	})
}
