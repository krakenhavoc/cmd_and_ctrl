package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathrite Shaman — Creature — Elf Shaman {B/G}, 1/2 (EDHREC rank
// 605):
//
//	"{T}: Exile target land card from a graveyard. Add one mana of
//	 any color. (Activate only as an instant.)
//	 {B}, {T}: Exile target instant or sorcery card from a graveyard.
//	 Each opponent loses 2 life.
//	 {G}, {T}: Exile target creature card from a graveyard. You gain
//	 2 life."
//
// A one-drop that is a mana dork, a graveyard hoser and a win
// condition. Three CR 602 activated abilities, each with a graveyard
// target clause over EVERY graveyard (no YouOwn — "from a graveyard"
// is the whole point), the same TargetCardInGraveyard the zone
// browser answers for Sun Titan.
//
// The first ability adds mana but is NOT a mana ability, because it
// targets (CR 605.1a): it goes on the stack, can be responded to, and
// the mana lands on resolution through the AddMana primitive — the
// five-colour pipe queues the same colour pick a Treasure does,
// narrowed to the commander's identity. "Activate only as an instant"
// is the default timing for an activated ability and needs nothing.
//
// Summoning sickness applies to all three (a creature source with a
// tap cost); the engine enforces it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22f1a4a4-c423-4d1c-8775-0ed604a9fa51",
		Name:         "Deathrite Shaman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Exile target land card from a graveyard. Add one mana of any color.",
				Cost:    TapCost(),
				Targets: TargetCardInGraveyard("target land card in a graveyard", Land()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					if err := (ExileTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
						return err
					}
					return AddMana{Produced: "{W|U|B|R|G}"}.Apply(ctx)
				},
			},
			{
				Label:   "{B}, {T}: Exile target instant or sorcery card from a graveyard. Each opponent loses 2 life.",
				Cost:    Plus(ManaCost("{B}"), TapCost()),
				Targets: TargetCardInGraveyard("target instant or sorcery card in a graveyard", Or(Instant(), Sorcery())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					if err := (ExileTarget{Target: item.Targets[0].ID}).Apply(NewContext(g, item)); err != nil {
						return err
					}
					return eachOpponentLosesLife(g, item, 2)
				},
			},
			{
				Label:   "{G}, {T}: Exile target creature card from a graveyard. You gain 2 life.",
				Cost:    Plus(ManaCost("{G}"), TapCost()),
				Targets: TargetCardInGraveyard("target creature card in a graveyard", Creature()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					if err := (ExileTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
				},
			},
		},
	})
}
