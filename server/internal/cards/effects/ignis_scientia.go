package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ignis Scientia — Legendary Creature — Human Advisor {1}{G}{U}, 2/2 (Final
// Fantasy, #4462):
//
//	"When Ignis Scientia enters, look at the top six cards of your library.
//	 You may put a land card from among them onto the battlefield tapped.
//	 Put the rest on the bottom of your library in a random order.
//	 I've Come Up with a New Recipe! — {1}{G}{U}, {T}: Exile target card
//	 from a graveyard. If a creature card was exiled this way, create a Food
//	 token."
//
// The activated ability exiles the target card and, when it was a creature
// card (read as it was in the graveyard), makes a Food. If the target is
// illegal as it resolves the whole ability does nothing, so no Food.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23320d55-80cb-4da3-a53c-8c75abdc8c8f",
		Name:         "Ignis Scientia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ignis Scientia — look at the top six cards of your library; you may put a land card from among them onto the battlefield tapped; put the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					return ignisLookAtTopSix(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "I've Come Up with a New Recipe! — {1}{G}{U}, {T}: Exile target card from a graveyard. If a creature card was exiled this way, create a Food token.",
			Cost:    Plus(ManaCost("{1}{G}{U}"), TapCost()),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  ignisNewRecipe,
		}},
	})
}

// ignisLookAtTopSix is the enters trigger: look at the top six, a land card
// from among them may go onto the battlefield tapped, and the rest go on the
// bottom of the library in a random order.
func ignisLookAtTopSix(ctx *Context) error {
	player := ctx.Controller()
	looked := ctx.Game.LookAtTopOfLibraryForEffect(player, 6)
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    looked,
		Match:    Land(),
		Max:      1,
		Optional: true,
		Tapped:   true,
		Label:    "Ignis Scientia — you may put a land card from among them onto the battlefield tapped",
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}

func ignisNewRecipe(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		wasCreature := c.IsCreature()
		if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(t.ID); !wasCreature || z == nil || z.Kind != game.ZoneExile {
			return nil
		}
		return CreateToken{Template: FoodToken(), N: 1}.Apply(ctx)
	}
	return nil
}
