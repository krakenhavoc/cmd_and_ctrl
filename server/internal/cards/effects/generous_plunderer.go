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
// off every artifact they have accepted. Two printed abilities and
// three stack items:
//
//   - The upkeep trigger is the controller's-upkeep condition with a
//     "you may" prompt. On resolution the controller gets a Treasure,
//     and "when you do" is a CR 603.12 REFLEXIVE trigger (#636) that
//     goes on the stack above it and targets an opponent, who gets a
//     tapped one (tappedTreasureToken, the S21 tapped entry).
//   - The attack trigger reads the defending player behind whatever
//     the Plunderer was declared at (b17DefendingPlayer, so a
//     planeswalker attack still hits its controller) and deals
//     b03ArtifactsControlled of theirs in damage, counted at
//     resolution as printed.
//
// The reflexive half is what makes the timing printed rather than
// convenient: the opponent is chosen after your Treasure exists
// rather than before you decided to make one, the table gets a window
// to respond to the gift on its own, and an opponent who leaves in
// that window costs you the second Treasure only — where the folded
// version used to counter the whole trigger and cost you yours too.
func init() {
	Register(Spec{
		OracleID:        "91c835d1-22ca-4c90-9ba6-c8e01bbc0347",
		Name:            "Generous Plunderer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Optional(
				AtYourUpkeep("Generous Plunderer — create a Treasure", generousPlundererTreasure),
				"Generous Plunderer — create a Treasure (and give target opponent a tapped one)?"),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				Key: "Generous Plunderer — damage to defending player equal to their artifacts",
				// The defending player is a board read at trigger time
				// (ADR 0041 P9's fill-in Build): b17DefendingPlayer resolves
				// the attack's target to a player once, when the attack was
				// declared, and the resolving item cannot re-derive it later.
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Generous Plunderer — damage to defending player equal to their artifacts")
					item.Params.Player = b17DefendingPlayer(g, ev)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					defender := item.Params.Player
					n := b03ArtifactsControlled(g, defender)
					if n == 0 {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: defender, Amount: n}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// generousPlundererTreasure is the upkeep body: your Treasure, and —
// because you did — the reflexive trigger that gives an opponent a
// tapped one. "When you do" is unconditional here: the "you may" was
// the trigger's own prompt, and answering yes is the doing.
//
// Caller holds g.mu.
func generousPlundererTreasure(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return ReflexiveTrigger{
		Label: "Generous Plunderer — a tapped Treasure for target opponent",
		Body:  generousPlundererGiftBody,
	}.Apply(ctx)
}

// generousPlundererGift is the reflexive half: the opponent chosen
// when this trigger went on the stack creates a tapped Treasure.
//
// Caller holds g.mu.
func generousPlundererGift(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	return CreateToken{Controller: item.Targets[0].ID, Template: tappedTreasureToken(), N: 1}.
		Apply(NewContext(g, item))
}
