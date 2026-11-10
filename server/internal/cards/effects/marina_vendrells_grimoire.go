package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marina Vendrell's Grimoire — Legendary Artifact — Book {5}{U} (Duskmourn):
//
//	"When Marina Vendrell's Grimoire enters, if you cast it, draw five
//	 cards.
//	 You have no maximum hand size and don't lose the game for having 0 or
//	 less life.
//	 Whenever you gain life, draw that many cards.
//	 Whenever you lose life, discard that many cards. Then if you have no
//	 cards in hand, you lose the game."
//
// Five pieces, each an existing seam:
//
//   - the enters trigger is Cresting Mosasaurus's "if you cast it"
//     (b16EnteredFromStack), a condition on the trigger, so a Grimoire put
//     onto the battlefield draws nothing;
//   - "no maximum hand size" is a hand-size static for you
//     (HandSizeYou / HandSizeNoMaximum);
//   - "don't lose for having 0 or less life" is Phyrexian Unlife's game-end
//     gate narrowed to that one cause (LossLife): other ways to lose still
//     work, and the gate is read at every check, so it stops when the
//     Grimoire leaves;
//   - the gain trigger captures the amount from the event as it triggers
//     (Sanguine Bond's reading), lifelink included, and draws that many;
//   - the loss trigger is the life that was really lost (s22PlayerLostLife:
//     damage counts only the life it cost) and discards that many cards of
//     your choice (a keyed discard), THEN checks the hand: it is empty
//     afterwards, even if it was empty to begin with, and you lose the game
//     (CR 104.3). Discarding is "as many as you can" (CR 701.9b).
func init() {
	const gainLabel = "Marina Vendrell's Grimoire — draw that many cards"
	const loseLabel = "Marina Vendrell's Grimoire — discard that many cards, then if you have no cards in hand, you lose the game"
	gain := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Target == source.Controller && ev.Amount > 0
		},
		Key: gainLabel,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, gainLabel)
			item.Params.Amount = ev.Amount
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Params.Amount <= 0 {
				return nil
			}
			return g.DrawNForEffect(item.Controller, item.Params.Amount)
		},
	}
	lose := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			_, ok := s22PlayerLostLife(ev, source.Controller, g)
			return ok
		},
		Key: loseLabel,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, loseLabel)
			item.Params.Amount, _ = s22PlayerLostLife(ev, source.Controller, g)
			return item
		},
		Effect: marinaVendrellsGrimoireDiscard,
	}
	Register(Spec{
		OracleID:     "096eac52-00c2-4439-b2f3-ff555d15ec88",
		Name:         "Marina Vendrell's Grimoire",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 5},
		HandSize:     []game.HandSizeStatic{{Players: game.HandSizeYou, Kind: game.HandSizeNoMaximum}},
		GameEndGates: []game.GameEndGate{{Scope: game.GateYou, CantLose: true, Causes: []game.LossCause{game.LossLife}}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, marinaVendrellsGrimoireYouCastIt),
				"Marina Vendrell's Grimoire — draw five cards", Do(DrawCards{N: 5})),
			gain,
			lose,
		},
	})
}

// marinaVendrellsGrimoireYouCastIt is the enter trigger's "if you cast it".
func marinaVendrellsGrimoireYouCastIt(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && b16EnteredFromStack(g, source.InstanceID)
}

// marinaVendrellsGrimoireDiscard discards as many cards as life was lost,
// then checks the hand.
func marinaVendrellsGrimoireDiscard(g *game.Game, item *game.StackItem) error {
	if item.Params.Amount <= 0 {
		return nil
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        item.Params.Amount,
		Question: "Marina Vendrell's Grimoire — discard that many cards",
	}, marinaVendrellsGrimoireLose, nil)
}

// marinaVendrellsGrimoireLose is "then if you have no cards in hand, you
// lose the game", keyed. An on-disk key: never renamed, never reused.
var marinaVendrellsGrimoireLose = game.RegisterDiscardThen("choose-cards/marina-vendrells-grimoire-lose-if-empty", func(g *game.Game, r game.DiscardsDone) error {
	p := g.PlayerByIDForEffect(r.Player)
	if p == nil || p.Hand == nil || p.Hand.Size() > 0 {
		return nil
	}
	return LoseTheGame{Player: r.Player}.Apply(keyedContext(g, r.Player, r.Source))
})
