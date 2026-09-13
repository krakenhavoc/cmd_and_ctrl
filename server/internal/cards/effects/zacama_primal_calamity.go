package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zacama, Primal Calamity — Legendary Creature — Elder Dinosaur
// {6}{R}{G}{W}, 9/9 (EDHREC rank 1797):
//
//	"Reach, vigilance, trample
//	 When Zacama enters, if you cast it, untap all lands you control.
//	 {2}{R}: Zacama deals 3 damage to target creature.
//	 {2}{G}: Destroy target artifact or enchantment.
//	 {2}{W}: You gain 3 life."
//
// The nine-mana Naya commander that pays for itself. Three keywords
// on PrintedKeywords; three ordinary activated abilities, each one
// primitive with its own target clause; and an ETB whose intervening
// "if you cast it" is read off the event log (b16EnteredFromStack —
// the zone move that preceded the ETB came from the stack), so a
// reanimated or flickered Zacama untaps nothing, as printed. The
// untap is every tapped land the controller controls.
//
// Intervening-if caveat, shared with every such card in the catalog:
// the condition is checked when the trigger is put on the stack and
// not again on resolution, which for "if you cast it" can never
// change between the two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "23270e4a-a222-4ffb-a946-d0e20d665187",
		Name:            "Zacama, Primal Calamity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b16EnteredFromStack(g, source.InstanceID)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Zacama, Primal Calamity — untap all lands you control",
					func(g *game.Game, item *game.StackItem) error {
						return b16UntapAllYouControlMatching(NewContext(g, item), item.Controller, func(c game.Card) bool { return c.IsLand() })
					})
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{R}: Zacama deals 3 damage to target creature.",
				Cost:    ManaCost("{2}{R}"),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: id, Amount: 3}.Apply(ctx)
				},
			},
			{
				Label:   "{2}{G}: Destroy target artifact or enchantment.",
				Cost:    ManaCost("{2}{G}"),
				Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return DestroyTarget{Target: id}.Apply(ctx)
				},
			},
			{
				Label: "{2}{W}: You gain 3 life.",
				Cost:  ManaCost("{2}{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
