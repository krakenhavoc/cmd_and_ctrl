package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret, Cruel Captain — Legendary Planeswalker — Tezzeret {3},
// loyalty 4:
//
//	"Whenever an artifact you control enters, put a loyalty counter on
//	 Tezzeret.
//	 0: Untap target artifact or creature. If it's an artifact
//	 creature, put a +1/+1 counter on it.
//	 −3: Search your library for an artifact card with mana value 1 or
//	 less, reveal it, put it into your hand, then shuffle.
//	 −7: You get an emblem with "At the beginning of combat on your
//	 turn, put three +1/+1 counters on target artifact you control. If
//	 it's not a creature, it becomes a 0/0 Robot artifact creature.""
//
// The emblem's target is chosen when its trigger goes on the stack and
// re-checked on resolution. "It becomes a 0/0 Robot artifact creature"
// states no duration, so it is the indefinite type change pinned to
// that object (CR 611.2a, as the Mirage hidden enchantments do). The
// artifact is already an artifact, so the type change adds Creature and
// the Robot subtype and sets the base size; the counters then make it
// the creature it is. The "not a creature" test is made on resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0eb11b2d-a397-48da-a0f3-9f0f83f42282",
		Name:            "Tezzeret, Cruel Captain",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Tezzeret, Cruel Captain emblem",
			Text:  "At the beginning of combat on your turn, put three +1/+1 counters on target artifact you control. If it's not a creature, it becomes a 0/0 Robot artifact creature.",
			Triggered: []game.TriggeredAbility{
				Targeting(
					AtBeginningOfYourCombat("Tezzeret, Cruel Captain emblem — three +1/+1 counters on an artifact", tezzeretEmblemBody),
					TargetPermanent("target artifact you control", Artifact(), YouControl())),
			},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact()
			}, "Tezzeret, Cruel Captain — put a loyalty counter on Tezzeret",
				func(g *game.Game, item *game.StackItem) error {
					if !b15OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterLoyalty, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "0: Untap target artifact or creature. If it's an artifact creature, put a +1/+1 counter on it.",
				Cost:    LoyaltyCost(0),
				Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						c, ok := g.LookupCardForEffect(t.ID)
						if !ok || !c.IsArtifact() || !c.IsCreature() {
							return nil
						}
						return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label: "−3: Search your library for an artifact card with mana value 1 or less, reveal it, put it into your hand, then shuffle.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() && c.ManaValue() <= 1 },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Tezzeret, Cruel Captain — search for an artifact card with mana value 1 or less",
						Source:    item.SourceCardID,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "−7: You get an emblem with \"At the beginning of combat on your turn, put three +1/+1 counters on target artifact you control. If it's not a creature, it becomes a 0/0 Robot artifact creature.\"",
				Cost:   LoyaltyCost(-7),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

// tezzeretEmblemBody is the emblem trigger's resolution: the creature
// test comes BEFORE the counters, so a noncreature artifact is
// animated as a 0/0 and then grows. Caller holds g.mu.
func tezzeretEmblemBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		if !c.IsCreature() {
			if err := (ScopedEffectFor{
				Target: t.ID,
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.AddSubtypesMod("Robot"),
					game.SetBasePowerMod(0),
					game.SetBaseToughnessMod(0),
				},
				Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
				Label:    "Tezzeret, Cruel Captain emblem — becomes a 0/0 Robot artifact creature",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
	}
	return nil
}
