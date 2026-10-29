package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silvan Reveler — Creature — Elf Citizen {2}{G}{U}, 3/2 (The Hobbit):
//
//	"When this creature enters, draw a card, then discard a card. If you
//	 discard a land card this way, put it from your graveyard onto the
//	 battlefield tapped.
//	 Landfall — Whenever a land you control enters, you may pay {1}{G}{U}.
//	 If you do, return this card from your graveyard to your hand."
//
// The loot is Glacial Dragonhunt's draw-then-keyed-discard: the draw lands
// first so the drawn card is a legal discard, and the discard is the
// player's own choice. Its answer puts the land onto the battlefield tapped
// (Reassembling Skeleton's Tapped return) only while it is still in the
// graveyard: a discarded land that a replacement sent elsewhere is not
// "put from your graveyard", and a discarded nonland does nothing.
//
// The landfall trigger watches from the graveyard (Bloodghast's InGraveyard)
// and asks for the payment as it resolves (Forsaken Miner's offer, with its
// hand twin): the return happens only if the mana was really paid, and only
// for the very object that triggered (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11932191-4b19-49b1-bfe4-abb7b83b2e59",
		Name:         "Silvan Reveler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Silvan Reveler — draw a card, then discard a card; a land goes onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: silvanRevelerDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				}),
			InGraveyard(Landfall("Silvan Reveler — you may pay {1}{G}{U} to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return mayPayForGraveyardReturnToHand(g, item, "{1}{G}{U}", "Silvan Reveler — pay {1}{G}{U} to return it to your hand?")
				})),
		},
	})
}

var (
	// The loot's discard half, run once the draw has landed. The key is an
	// on-disk identity: never renamed, never reused.
	silvanRevelerDiscard = game.RegisterDrawThen("silvan-reveler/discard", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        d.N,
			Question: "Silvan Reveler — discard a card (a land card comes back tapped)",
		}, silvanRevelerDiscarded, nil)
	})

	// The discard's answer, keyed (#4253): a land card still in the graveyard
	// is put onto the battlefield tapped under its owner's control.
	silvanRevelerDiscarded = game.RegisterDiscardThen("choose-cards/silvan-reveler-discarded", func(g *game.Game, r game.DiscardsDone) error {
		for _, id := range r.Discarded.By(r.Player) {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !c.IsLand() {
				continue
			}
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Tapped: true}).Apply(keyedContext(g, r.Player, r.Source)); err != nil {
				return err
			}
		}
		return nil
	})
)
