package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hazel's Brewmaster — Creature — Squirrel Warlock {3}{B}, 3/4
// (EDHREC rank 2326):
//
//	"Menace
//	 Whenever this creature enters or attacks, exile up to one target
//	 card from a graveyard and create a Food token.
//	 Foods you control have all activated abilities of all creature
//	 cards exiled with this creature."
//
// The Food deck's graveyard hate. Menace rides PrintedKeywords. The
// enters-or-attacks trigger is one printed ability with two trigger
// conditions (b21SelfEnteredOrAttacked, Sun Titan's shape), and its
// "up to one target card from a graveyard" is a zero-or-one target
// clause over every graveyard at the table.
//
// It is declared as TWO TriggeredAbility entries with the same label
// because of how a targeted trigger is dispatched: the engine drops
// a targeted trigger outright when its legal set is empty (CR
// 603.3d), which is right for "exile target card" and wrong for "up
// to one … AND create a Food" — with every graveyard empty the Food
// must still come. So the targeted entry fires only while some
// graveyard holds a card (b21AnyGraveyardHasACard, the same question
// the legal-set walk answers), and the untargeted entry fires only
// when none does. Exactly one of the two applies to any event. With
// a target chosen that then leaves the graveyard in response, the
// ability is countered by game rules and makes no Food — CR 608.2b,
// as printed; a pick of no target resolves and makes the Food.
//
// The exile is linked (ExileTarget.Linked, ADR 0155 §8), so the card
// lands with Card.ExiledWith naming this Brewmaster object. The third
// ability is a borrow whose recipients are the Foods you control
// (GrantActivatedAbilitiesOf over ExiledWithLenders, ADR 0155 PR 3): each
// of your Foods has every activated ability of every creature card linked
// to this Brewmaster, beside its own sacrifice ability. A lent "{T}" taps
// the Food and "sacrifice this creature" sacrifices the Food (CR 201.5b).
// A creature card that leaves exile stops lending at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8180024-1979-4677-9a0d-e08d4b7c825a",
		Name:            "Hazel's Brewmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static:          []game.StaticAbility{b21BrewmasterBorrow()},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB, game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b21SelfEnteredOrAttacked(ev, source) && b21AnyGraveyardHasACard(g)
				},
				Targets: TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
				Key:     b21BrewmasterLabel,
				Effect:  b21BrewmasterExileAndFood,
			},
			OnAny([]game.EventKind{game.EventETB, game.EventAttack}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b21SelfEnteredOrAttacked(ev, source) && !b21AnyGraveyardHasACard(g)
			}, b21BrewmasterLabel, b21BrewmasterExileAndFood),
		},
	})
}

// b21BrewmasterLabel is the stack label both declarations share —
// one printed ability, one name on the stack.
const b21BrewmasterLabel = "Hazel's Brewmaster — exile up to one card from a graveyard, create a Food"

// b21BrewmasterExileAndFood is the trigger's body: exile the chosen
// graveyard card, linked to the Brewmaster, if one was chosen and is
// still there, then create a Food for the controller.
func b21BrewmasterExileAndFood(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ExileTarget{Target: t.ID, Linked: true}).Apply(ctx); err != nil {
			return err
		}
		break
	}
	return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(ctx)
}

// b21BrewmasterBorrow is "Foods you control have all activated abilities
// of all creature cards exiled with this creature."
func b21BrewmasterBorrow() game.StaticAbility {
	borrow := GrantActivatedAbilitiesOf(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.HasSubtype("Food") && target.Controller == source.Controller
	}, game.AbilityBorrow{
		Label:   "Foods you control have all activated abilities of all creature cards exiled with this creature.",
		Lenders: ExiledWithLenders(ByController(Creature())),
	})
	borrow.DependsOnExile = true
	return borrow
}
