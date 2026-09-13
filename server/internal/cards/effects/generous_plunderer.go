package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Generous Plunderer — Creature — Human Rogue {1}{R}, 2/2 (EDHREC
// rank 2068):
//
//	"Menace
//	 At the beginning of your upkeep, you may create a Treasure token.
//	 When you do, target opponent creates a tapped Treasure token.
//	 Whenever this creature attacks, it deals damage to defending
//	 player equal to the number of artifacts they control."
//
// The Treasure-gifting Rogue: every upkeep a Treasure for you and a
// tapped one for the opponent you mean to hit, and the attack pays
// off every artifact they have accepted. Two triggers:
//
//   - The upkeep trigger is the controller's-upkeep condition with a
//     "you may" prompt and a target-opponent clause; on resolution
//     the controller gets a Treasure and the opponent a tapped one
//     (tappedTreasureToken, the S21 tapped entry).
//   - The attack trigger reads the defending player behind whatever
//     the Plunderer was declared at (b17DefendingPlayer, so a
//     planeswalker attack still hits its controller) and deals
//     b03ArtifactsControlled of theirs in damage, counted at
//     resolution as printed.
//
// Sandbox simplification, declared (the Overlook lands' posture): the
// printed "when you do" is a REFLEXIVE trigger — your Treasure is
// created, then a second ability targets the opponent — and here the
// two are one ability with the opponent chosen when it goes on the
// stack. Weaker, never stronger: with the target removed in response
// (an opponent leaving the game) the whole trigger is countered and
// you get no Treasure either, where printed you would keep yours.
func init() {
	Register(Spec{
		OracleID:        "91c835d1-22ca-4c90-9ba6-c8e01bbc0347",
		Name:            "Generous Plunderer",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"You pick the opponent when the upkeep trigger goes on the stack rather than after your Treasure is made, so the two Treasures are one ability instead of two."},
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginUpkeep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Generous Plunderer — create a Treasure (and give target opponent a tapped one)?"},
				Targets:        TargetPlayer("target opponent", Opponent()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Generous Plunderer — a Treasure for you, a tapped Treasure for target opponent",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
								return nil
							}
							ctx := NewContext(g, item)
							if err := (CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
								return err
							}
							return CreateToken{Controller: item.Targets[0].ID, Template: tappedTreasureToken(), N: 1}.Apply(ctx)
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					defender := b17DefendingPlayer(g, ev)
					return game.NewTriggeredItem(source, "Generous Plunderer — damage to defending player equal to their artifacts",
						func(g *game.Game, item *game.StackItem) error {
							n := b03ArtifactsControlled(g, defender)
							if n == 0 {
								return nil
							}
							return DealDamage{Source: item.SourceCardID, Target: defender, Amount: n}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
