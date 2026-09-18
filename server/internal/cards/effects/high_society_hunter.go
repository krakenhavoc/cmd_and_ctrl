package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// High-Society Hunter — 5/3 Creature — Vampire Noble for {3}{B}{B}
// (EDHREC rank 4427):
//
//	"Flying
//	 Whenever this creature attacks, you may sacrifice another
//	 creature. If you do, put a +1/+1 counter on this creature.
//	 Whenever another nontoken creature dies, draw a card."
//
// A five-mana evasive body that turns the whole table's attrition into
// cards. The second trigger is the reason it plays in Commander: it
// watches ANY nontoken creature die, anyone's, so every board wipe and
// every combat the table has without you is a fistful of cards.
// Roadmap batch 42 (#449), "no new machinery".
//
// "Another NONTOKEN creature" excludes tokens, which is the printed
// card's own brake — the Hunter is not a sacrifice-deck payoff, it is
// a grindy-table payoff, and a Bitterblossom board does nothing for
// it. It also excludes itself, so the Hunter's own death draws
// nothing.
//
// The attack trigger's counter follows ONLY from the sacrifice: "if
// you do" is a real condition, and declining the prompt leaves the
// Hunter a 5/3.
//
// DECLARED SIMPLIFICATION (#259), the Chitterspitter caveat: the
// creature to sacrifice is picked when the attack trigger goes on the
// stack rather than when it resolves, through the engine's picker for
// a choice among your own permanents. Two consequences, both weaker
// than printed and never stronger — an opponent can see the choice
// and remove the creature in response, in which case the trigger
// fizzles and no counter is placed where the printed card would let
// you pick another; and because that picker is a target clause, a
// creature you control with hexproof or shroud cannot be chosen,
// where a printed sacrifice targets nothing and could take it.
// Resolution-time permanent choice is the missing engine piece.
func init() {
	Register(Spec{
		OracleID:     "0dc158ba-7ec8-4558-91f5-0a87ef8380d4",
		Name:         "High-Society Hunter",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You choose the creature to sacrifice when the attack trigger goes on the stack rather than on resolution, so opponents can respond to the choice — and if they remove it, no counter is placed.",
			"A creature you control with hexproof or shroud can't be chosen for the sacrifice.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WheneverThisAttacks("High-Society Hunter — sacrifice another creature, put a +1/+1 counter on this creature",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
								return nil
							}
							ctx := NewContext(g, item)
							if err := (SacrificePermanent{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
								return err
							}
							return AddCounter{
								Target: item.SourceCardID,
								Kind:   game.CounterPlusOne,
								N:      1,
							}.Apply(ctx)
						}),
					PermanentYouControl("another creature you control", Creature(), b03NotNamed("High-Society Hunter"))),
				"High-Society Hunter — sacrifice another creature to put a +1/+1 counter on it?"),
			On(game.EventLTB,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if !AnotherCreatureDied(ev, source, lki, g) {
						return false
					}
					dead, ok := diedCreature(ev, g)
					return ok && !IsToken(dead)
				},
				"High-Society Hunter — draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
