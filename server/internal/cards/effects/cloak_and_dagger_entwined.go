package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloak and Dagger, Entwined — Legendary Creature — Human Hero {1}{W}{B}
// 2/2 (#4585, Marvel Super Heroes #4455):
//
//	"Deathtouch, lifelink
//	 When Cloak and Dagger enter, choose target opponent and up to one
//	 target creature they control. They reveal their hand. You may exile
//	 a nonland card from their hand or the chosen creature until Cloak
//	 and Dagger leave the battlefield."
//
// THE TARGETS. The creature clause is related to the opponent clause
// (ControlledByTargetsController, whose anchor is a player here as in
// Down for Repairs and Ravager of the Fells): it must be controlled by
// the opponent chosen, judged as the ability goes on the stack (CR
// 601.2c, 603.3d) and again as it resolves (CR 608.2b). "Up to one" lets
// the trigger go on the stack with an opponent alone.
//
// THE REVEAL AND THE CHOICE. The hand is revealed to the whole table
// first (CR 701.20a), then the controller chooses one of: a nonland card
// from that hand, the chosen creature, or neither ("you may"). The hand
// is public by then, so choosing the kind before the card loses nothing.
// A hand with no nonland card, or a creature that is no longer a legal
// target, simply leaves that choice out.
//
// THE "UNTIL". The creature is exiled through ExileUntil{ThisLeaves} and
// comes back to the battlefield under its owner's control (CR 610.3c).
// The hand card is exiled linked to Cloak and Dagger, and its return is
// recorded afterwards as an UntilReturn with ToHand: CR 610.3 returns an
// object to its previous zone, so it goes back into its owner's hand,
// not onto the battlefield. If Cloak and Dagger left before the trigger
// resolved, nothing is exiled (CR 610.3b) and the hand is only revealed.
//
// No simplification.
const (
	cloakDaggerRevealReason = "Cloak and Dagger, Entwined — reveals their hand"
	cloakDaggerNothing      = "Exile nothing"
	cloakDaggerFromHand     = "Exile a nonland card from their hand"
	cloakDaggerHeldLabel    = "Cloak and Dagger, Entwined — the exiled card returns when it leaves the battlefield"
)

func init() {
	enters := WhenThisEnters(
		"Cloak and Dagger, Entwined — choose target opponent and up to one target creature they control; they reveal their hand; you may exile a nonland card from their hand or the chosen creature until Cloak and Dagger leave the battlefield",
		cloakAndDaggerEnters)
	enters.Targets = Clauses(
		TargetPlayer("target opponent", Opponent()),
		ControlledByTargetsController(
			TargetCreature("up to one target creature they control").WithCount(0, 1), 0),
	)
	Register(Spec{
		OracleID:        "64132f93-66ac-4794-8c07-a88f9ed0e22b",
		Name:            "Cloak and Dagger, Entwined",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

// cloakAndDaggerEnters reveals the opponent's hand and asks which of the
// exiles, if either, the controller wants.
func cloakAndDaggerEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := TargetedPlayer(ctx)
	if player == uuid.Nil {
		return nil
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	hand := make([]uuid.UUID, 0, len(p.Hand.Cards))
	nonland := false
	nonlandOK := Nonland()
	for _, c := range p.Hand.Cards {
		hand = append(hand, c.InstanceID)
		if nonlandOK(g, ctx.Controller(), c) {
			nonland = true
		}
	}
	g.RevealForEffect(game.RevealSpec{
		Player: player,
		Source: ctx.Source(),
		Reason: cloakDaggerRevealReason,
		Cards:  hand,
	})
	// CR 610.3b: Cloak and Dagger left before this resolved, so neither
	// exile happens.
	ref, ok := ctx.SourceRef()
	if !ok || g.UntilEventHappenedForEffect(game.UntilReturn{Leaves: ref}, item) {
		return nil
	}
	options := []game.ChoiceOption{{Label: cloakDaggerNothing}}
	if nonland {
		options = append(options, game.ChoiceOption{Label: cloakDaggerFromHand})
	}
	if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
		name := "the chosen creature"
		if c, found := g.LookupCardForEffect(t.ID); found {
			name = c.Name
		}
		options = append(options, game.ChoiceOption{Label: "Exile " + name, Cards: []uuid.UUID{t.ID}})
	}
	if len(options) == 1 {
		return nil
	}
	return PickOption{
		Question: "Cloak and Dagger, Entwined — exile a nonland card from their hand or the chosen creature?",
		Options:  options,
		ThenKey:  cloakDaggerChose,
		WithItem: true,
	}.Apply(ctx)
}

// cloakDaggerChose carries out the choice: the hand card (linked to
// Cloak and Dagger, returned to the hand by cloakDaggerHandReturn), the
// creature, or nothing.
var cloakDaggerChose = OptionPickThen("option-pick/cloak-and-dagger-entwined-choice",
	func(ctx *Context, r game.OptionPicked) error {
		if r.Option == nil {
			return nil
		}
		switch r.Option.Label {
		case cloakDaggerNothing:
			return nil
		case cloakDaggerFromHand:
			return ChooseFromRevealedHand{
				Player:   TargetedPlayer(ctx),
				Filter:   Nonland(),
				Label:    "nonland card",
				Exile:    true,
				Linked:   true,
				Revealed: true,
				Then:     cloakDaggerHandReturn,
			}.Apply(ctx)
		}
		t, ok := ctx.ClauseTarget(1)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		return ExileUntil{
			Target:     t.ID,
			ThisLeaves: true,
			Label:      "Cloak and Dagger, Entwined — the exiled creature returns when it leaves the battlefield",
		}.Apply(ctx)
	})

// cloakDaggerHandReturn records that the exiled hand card goes back to
// its owner's hand when Cloak and Dagger leave the battlefield. The
// exile linked the card to the Cloak and Dagger object it was exiled by
// (CR 607.2a), which is the object the return waits on.
var cloakDaggerHandReturn = RevealedPickThen("revealed-pick/cloak-and-dagger-entwined-return",
	func(ctx *Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		id := pick.Chosen[0].InstanceID
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.ExiledWith.ID == uuid.Nil {
			return nil
		}
		ctx.Game.ScheduleUntilReturnForEffect(id, game.UntilReturn{
			Controller:   pick.Chooser,
			SourceCardID: pick.Source,
			Label:        cloakDaggerHeldLabel,
			Leaves:       game.ObjectRef(c.ExiledWith),
			ToHand:       true,
		})
		return nil
	})
