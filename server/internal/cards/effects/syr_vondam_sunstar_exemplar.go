package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syr Vondam, Sunstar Exemplar — Legendary Creature — Human Knight {W}{B},
// 2/2 (Edge of Eternities, #4461):
//
//	"Vigilance, menace
//	 Whenever another creature you control dies or is put into exile, put
//	 a +1/+1 counter on Syr Vondam and you gain 1 life.
//	 When Syr Vondam dies or is put into exile while its power is 4 or
//	 greater, destroy up to one target nonland permanent."
//
// Both triggers are leaves-the-battlefield look-backs (CR 603.10a). The
// first is one trigger per creature that left the battlefield for a
// graveyard or exile (a creature card put into exile from a hand or a
// graveyard is not a permanent that left, per the rulings); the creature
// is judged by what it was and who controlled it as it last existed, and
// Syr Vondam leaving at the same time still triggers it for the others.
// The counter goes on Syr Vondam only if he is still on the battlefield;
// the life is gained either way. The second is Syr Vondam's own dies or
// exile, with his power as it last existed (counters included) at 4 or
// more, and it targets up to one nonland permanent as it goes on the
// stack. Vigilance and menace are printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e597063-c077-4236-a9e4-b0721b24677f",
		Name:            "Syr Vondam, Sunstar Exemplar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, stdB3AnotherCreatureYouControlDiedOrExiled,
				"Syr Vondam, Sunstar Exemplar — put a +1/+1 counter on Syr Vondam and gain 1 life",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, GainLife{Amount: 1})),
			Targeting(
				On(game.EventLTB, stdB3SyrVondamLeftWithPower4,
					"Syr Vondam, Sunstar Exemplar — destroy up to one target nonland permanent", stdB3DestroyUpToOneTarget),
				TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1)),
		},
	})
}

// stdB3AnotherCreatureYouControlDiedOrExiled is the first trigger's
// condition: a creature other than the source, controlled by the source's
// controller as it left, went from the battlefield to a graveyard or to
// exile.
func stdB3AnotherCreatureYouControlDiedOrExiled(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.CardID == source.InstanceID {
		return false
	}
	if ev.NewZone != game.ZoneGraveyard && ev.NewZone != game.ZoneExile {
		return false
	}
	c, _ := g.LookupCardForEffect(ev.CardID)
	return leftAsType(ev, c, "creature") && leftUnderControlOf(ev, c) == source.Controller
}

// stdB3SyrVondamLeftWithPower4 is the second trigger's condition: the
// source died or was exiled from the battlefield with power 4 or more as
// it last existed.
func stdB3SyrVondamLeftWithPower4(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return b22SelfDiedOrWasExiledFromBattlefield(ev, source) && b13LastKnownPower(g, source.InstanceID, lki) >= 4
}

// stdB3DestroyUpToOneTarget destroys the chosen permanent if it is still
// legal as the trigger resolves.
func stdB3DestroyUpToOneTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return DestroyTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}
