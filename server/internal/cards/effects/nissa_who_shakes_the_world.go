package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa, Who Shakes the World — Legendary Planeswalker — Nissa {3}{G}{G},
// starting loyalty 5:
//
//	"Whenever you tap a Forest for mana, add an additional {G}.
//	 +1: Put three +1/+1 counters on up to one target noncreature land you
//	 control. Untap it. It becomes a 0/0 Elemental creature with vigilance
//	 and haste that's still a land.
//	 −8: You get an emblem with "Lands you control have indestructible."
//	 Search your library for any number of Forest cards, put them onto the
//	 battlefield tapped, then shuffle."
//
// The mana ability is a triggered mana ability (Crypt Ghast's shape). The
// +1 states no duration, so the land stays a 0/0 Elemental creature for
// good (CR 611.2a), pinned to that object, and the three counters make
// it 3/3. The −8 makes the emblem first, then searches: "any number" is
// the whole stack of Forest cards the player picks, put onto the
// battlefield together, then the library is shuffled.
func init() {
	Register(Spec{
		OracleID:        "5aa3abf1-d56b-4f42-8c84-7e5a2c15ee0f",
		Name:            "Nissa, Who Shakes the World",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		ManaTriggers: []game.ManaTrigger{
			WheneverYouTapALandOfSubtypeForMana(
				"Nissa, Who Shakes the World — add an additional {G}", "Forest", AddsFixedMana("{G}")),
		},
		Emblem: &EmblemSpec{
			Label: "Nissa, Who Shakes the World emblem",
			Text:  "Lands you control have indestructible.",
			Static: []game.StaticAbility{
				KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsLand() && target.Controller == source.Controller
				}, "indestructible"),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Put three +1/+1 counters on up to one target noncreature land you control. Untap it. It becomes a 0/0 Elemental creature with vigilance and haste that's still a land.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPermanent("up to one target noncreature land you control", Land(), Noncreature(), YouControl()).WithCount(0, 1),
				Effect:  nissaWhoShakesAnimate,
			},
			{
				Label: "−8: You get an emblem with \"Lands you control have indestructible.\" Search your library for any number of Forest cards, put them onto the battlefield tapped, then shuffle.",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateEmblem{}).Apply(ctx); err != nil {
						return err
					}
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.IsLand() && c.HasSubtype("Forest") },
						Dest:          game.ZoneBattlefield,
						TappedOnEntry: true,
						Unbounded:     true,
						Shuffle:       true,
						Source:        item.SourceCardID,
						Reason:        "Nissa, Who Shakes the World — search for any number of Forest cards",
					}.Apply(ctx)
				},
			},
		},
	})
}

// nissaWhoShakesAnimate is the +1: counters, untap, then the permanent
// type change, in the printed order.
func nissaWhoShakesAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var id uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			id = t.ID
			break
		}
	}
	if id == uuid.Nil {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 3}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target: id,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
			game.SetBasePowerMod(0),
			game.SetBaseToughnessMod(0),
			game.AddKeywordsMod("vigilance", "haste"),
		},
		Duration: g.PinnedTo(game.IndefiniteDuration(), id),
		Label:    "Nissa, Who Shakes the World — becomes a 0/0 Elemental creature with vigilance and haste",
	}.Apply(ctx)
}
