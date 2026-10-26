package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Realmbreaker, the Invasion Tree — Legendary Artifact {3}:
//
//	"{2}, {T}: Target opponent mills three cards. Put a land card from
//	 their graveyard onto the battlefield tapped under your control. It
//	 gains 'If this land would leave the battlefield, exile it instead of
//	 putting it anywhere else.'
//	 {10}, {T}, Sacrifice Realmbreaker: Search your library for any
//	 number of Praetor cards, put them onto the battlefield, then
//	 shuffle."
//
// The first ability targets the opponent, mills, and THEN picks: the
// land is chosen as the ability resolves (CR 608.2c) from the whole of
// that graveyard, so a land the mill itself dropped is a legal pick, and
// the pick is not a target (hexproof on a land card in a graveyard means
// nothing). It is mandatory — with no land card there, nothing is asked.
// The land enters tapped under the activator's control on the entry
// event, not tapped a beat later (#1872).
//
// The granted "exile it instead" is the same indefinite replacement
// Unearth registers (ExileInsteadOfLeavingBattlefield), pinned to the
// land as an object. It is modelled as a replacement attached to the
// permanent rather than as a gained ability, so a land that loses its
// abilities would still be exiled; no card at this table does that to a
// land, and the direction is the one the card prints.
//
// The second ability is the Praetor tutor: the searcher may find any
// number, including none, up to the whole library (Protean Hulk's
// shape), and the library is shuffled afterwards. The sacrifice is part
// of the cost, so the Tree is gone before the search resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77398866-eebb-4091-b566-7726f8194a59",
		Name:         "Realmbreaker, the Invasion Tree",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}: Target opponent mills three cards. Put a land card from their graveyard onto the battlefield tapped under your control. It gains \"If this land would leave the battlefield, exile it instead of putting it anywhere else.\"",
				Cost:    Plus(ManaCost("{2}"), TapCost()),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect:  realmbreakerMillAndSteal,
			},
			{
				Label:   "{10}, {T}, Sacrifice Realmbreaker: Search your library for any number of Praetor cards, put them onto the battlefield, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{10}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := ctx.PlayerByID(item.Controller)
					if p == nil || p.Library == nil {
						return nil
					}
					return SearchLibrary{
						Player: item.Controller,
						// SearchLibrary lends the source's text (CR 612.2, #3984).
						Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Praetor") },
						Dest:      game.ZoneBattlefield,
						Limit:     p.Library.Size(),
						Shuffle:   true,
						Reason:    "Realmbreaker — any number of Praetor cards",
						Source:    item.SourceCardID,
					}.Apply(ctx)
				},
			},
		},
	})
}

// realmbreakerMillAndSteal is the first ability's body.
func realmbreakerMillAndSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	victim := targets[0].ID
	if err := (MillCards{Player: victim, N: 3}).Apply(ctx); err != nil {
		return err
	}
	candidates := graveyardCardIDs(ctx, victim, game.Card.IsLand)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: victim,
		Source:     item.SourceCardID,
		Question:   "Realmbreaker — put a land card from their graveyard onto the battlefield tapped under your control",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    realmbreakerSteal,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// realmbreakerSteal is the land chosen entering under the controller's
// control, tapped, keyed (#4253): the asking item rides the prompt as data.
var realmbreakerSteal = askedPick("choose-cards/realmbreaker-steal-land", func(ctx *Context, r game.CardsChosen) error {
	g, controller := ctx.Game, ctx.Controller()
	for _, id := range r.Picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: controller, Tapped: true}).Apply(ctx); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneBattlefield {
			ExileInsteadOfLeavingBattlefield(g, id, controller, "Realmbreaker: if this land would leave the battlefield, exile it instead")
		}
	}
	return nil
})
