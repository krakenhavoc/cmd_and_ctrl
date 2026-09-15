package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Engineer — Creature — Goblin Artificer {1}{R}, 1/2 (EDHREC
// rank 1185):
//
//	"When this creature enters, you may search your library for an
//	 artifact card, put it into your graveyard, then shuffle.
//	 {R}, {T}, Sacrifice an artifact: Return target artifact card with
//	 mana value 3 or less from your graveyard to the battlefield."
//
// The artifact deck's Entomb-and-Welder. The ETB is Entomb's search
// with an artifact predicate and a graveyard destination, optional
// (the prompt always opens; "fail to find" is an answer). The
// activated ability is a three-component cost — mana, tap (summoning
// sickness applies), sacrifice an artifact — with a graveyard target
// clause validated BEFORE the cost is paid, so the artifact you
// sacrifice can never be the one you return; that is the printed
// ordering (CR 601.2c before 601.2h) and the reason the card reads
// "an artifact" and not "another".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1d6cce8-085f-42cb-8b0c-b6fbbf88b16a",
		Name:         "Goblin Engineer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Goblin Engineer — search for an artifact card, put it into your graveyard", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsArtifact() },
					Dest:      game.ZoneGraveyard,
					Limit:     1,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Goblin Engineer — an artifact card, into your graveyard",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}, Sacrifice an artifact: Return target artifact card with mana value 3 or less from your graveyard to the battlefield.",
			Cost:    Plus(ManaCost("{R}"), TapCost(), b10SacrificeAnArtifact()),
			Targets: TargetCardInGraveyard("target artifact card with mana value 3 or less in your graveyard", YouOwn(), Artifact(), ManaValueLE(3)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
			},
		}},
	})
}
