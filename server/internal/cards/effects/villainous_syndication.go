package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Villainous Syndication — {1}{B} Enchantment — Plan (Marvel Super Heroes
// Commander, #4455):
//
//	"Tap an untapped Villain you control: Mill a card. Put a plan counter on
//	 this enchantment. Activate only as a sorcery.
//	 When the fourth plan counter is put on this enchantment, sacrifice it.
//	 When you do, return target creature card from your graveyard to the
//	 battlefield."
//
// The cost is Dungeon Descent's "tap an untapped <thing> you control" (not
// {T}, so a Villain that arrived this turn may pay it, CR 302.6). The fourth
// counter is Claim the Kingdom's plan recipe: sacrifice, then the reflexive
// trigger (CR 603.12) picks its graveyard target after the sacrifice, so
// the sacrificed enchantment is never a legal one (it is not a creature).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4777fa96-12a5-42d5-840f-ddd85f4a4b68",
		Name:         "Villainous Syndication",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "Tap an untapped Villain you control: Mill a card. Put a plan counter on this enchantment. Activate only as a sorcery.",
			Cost:         TapAnUntapped("an untapped Villain you control", Creature(), Subtype("Villain")),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (MillCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				return AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterPlaced, mshB2PlanCounterReached(4),
				"Villainous Syndication — sacrifice it; when you do, return target creature card from your graveyard to the battlefield",
				mshB2SacrificeThisThenWhenYouDo("Villainous Syndication — return target creature card from your graveyard to the battlefield", villainousSyndicationReturnBody)),
		},
	})
}

// villainousSyndicationReturnBody is the reflexive half: the chosen creature
// card comes back under its owner's control (CR 400.3). An on-disk key.
var villainousSyndicationReturnBody = game.ReflexiveBody("villainous-syndication/return",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx)
			}
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())
	}))
