package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// H.M.S. Bounty, Unlikely Timeship — Legendary Artifact — Spacecraft
// {4}{U}, 3/4 (Star Trek, #4973):
//
//	"When H.M.S. Bounty enters, you may search your library for a Whale
//	 card, reveal it, put it into your hand, then shuffle.
//	 Station
//	 8+ | Flying
//	 Whenever H.M.S. Bounty deals combat damage to a player, draw a
//	 card. If you control a Whale and haven't begun an extra turn this
//	 game, take an extra turn after this one."
//
// The enters search is Goblin Matron's (an optional search, so a player
// may find nothing and still shuffle). Station is the shared keyword
// (CR 702.184a); at eight charge counters the Spacecraft is a 3/4
// artifact creature with flying (CR 721.2a, 721.2b), and the combat
// damage trigger is part of the same striated block, so it exists only
// from eight counters on.
//
// The second sentence is checked as the trigger resolves, after the
// draw (it is not an intervening "if", CR 603.4): a Whale is read
// through the layered view, so a changeling counts, and "haven't begun
// an extra turn this game" is the engine's per-player count of extra
// turns that really began (Player.ExtraTurnsBegun, #4973) — so the
// trigger during the extra turn it gave does nothing more, and a turn
// that was queued but skipped never counted (CR 614.10a).
//
// No simplification.
func init() {
	t := WheneverThisDealsCombatDamageToAPlayer(
		"H.M.S. Bounty, Unlikely Timeship — draw a card; then, with a Whale and no extra turn begun this game, take an extra turn",
		hmsBountyCombatDamage)
	t.ActiveWhen = AtChargeCounters(8)
	Register(Spec{
		OracleID:     "eaa47a9b-79ad-4f42-a98b-dcf76138207c",
		Name:         "H.M.S. Bounty, Unlikely Timeship",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 3, 4),
			ThresholdKeywords(8, "flying"),
		),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("H.M.S. Bounty, Unlikely Timeship — search for a Whale card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					// "Whale" is a word of the source's text; SearchLibrary
					// lends it to each card (CR 612.2, #3984).
					Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Whale") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "H.M.S. Bounty — a Whale card, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
			t,
		},
	})
}

func hmsBountyCombatDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.ExtraTurnsBegunForEffect(item.Controller) > 0 || !trekB7YouControlA(g, item, "Whale") {
		return nil
	}
	return TakeExtraTurn{}.Apply(ctx)
}
